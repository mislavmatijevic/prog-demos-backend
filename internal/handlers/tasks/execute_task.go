package tasks

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

type ExecutionErrorCode int

const (
	EXEC_ERR_CODE_NO_TESTS ExecutionErrorCode = iota + 1
	EXEC_ERR_ARTEFACT_CONTENT_MISMATCH
	EXEC_ERR_TEST_FAILED
	EXEC_ERR_TIMEOUT
)

const CONTAINER_TIMEOUT_MARK = "timeout"

func (execErrCode ExecutionErrorCode) String() string {
	return [...]string{
		"Can't test this task.",
		"Artefact files did not contain expected contents.",
		"Program did not output expected test data.",
		"Execution took too long.",
	}[execErrCode-1]
}

func (execErrCode ExecutionErrorCode) EnumIndex() int {
	return int(execErrCode)
}

type taskExecutionRequest = struct {
	SolutionCode string `json:"solutionCode"`
}

type testDataMismatchReason = struct {
	TestInput      string `json:"testInput,omitempty"`
	Output         string `json:"output,omitempty"`
	ExpectedOutput string `json:"expectedOutput,omitempty"`
}

type taskExecutionResponse = struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	ErrorCode    int         `json:"errorCode"`
	ReasonFailed interface{} `json:"reason,omitempty"`
}

func ExecuteTask(w http.ResponseWriter, r *http.Request) {
	requestBody, err := getRequestBody(r)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	solutionCodeHasValue, solutionCode := utils.GetTrimmedStringWithValue(requestBody.SolutionCode)
	if !solutionCodeHasValue {
		api.RequestErrorHandlerCustomMsg(w, "Request body does not contain solution code.")
		return
	}

	var originalParamId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalParamId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}
	if !database.CheckTaskExists(taskId) {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Task with id %v not found.", taskId))
		return
	}

	userId, err := authentication.GetUserIdFromToken(r)
	if err != nil || userId == 0 {
		api.InternalErrorHandlerCustomMsg(w, "Couldn't get user from JWT token.")
		return
	}

	alreadyHasRunningTask := checkUserHasRunningTasks(userId)
	if alreadyHasRunningTask {
		api.TooEarlyErrorHandlerCustomMsg(w, "Previously submitted task still in progress!")
		return
	}

	taskExecution, err := markTaskExecutionStartForUserId(taskId, userId, solutionCode)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	cppFileForSyntaxChecking, err := createTempCppFile("", solutionCode)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		setTaskExecutionStatusFailed(taskExecution)
		return
	} else if !checkIsValidSolutionCode(cppFileForSyntaxChecking) {
		api.RequestErrorHandlerCustomMsg(w, "Code is not valid C++.")
		os.Remove(cppFileForSyntaxChecking.Name())
		setTaskExecutionStatusFailed(taskExecution)
		return
	} else {
		os.Remove(cppFileForSyntaxChecking.Name())
	}

	var tests []database.Test = database.GetTestsForTask(taskId)
	if len(tests) == 0 {
		sendTaskExecutionFailedResponse(w, EXEC_ERR_CODE_NO_TESTS, nil)
		setTaskExecutionStatusFailed(taskExecution)
		return
	}

	var res taskExecutionResponse

	for _, test := range tests {
		omitOutputsCheck := false
		testInput := test.Input

		cppFile, err := storeTempFiles(solutionCode, testInput)
		var tempDirPath = filepath.Dir(cppFile.Name())
		if err != nil {
			handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = runFileInIsolatedDockerContainerTask(ctx, cppFile)
		if err != nil {
			handleDockerContainerFail(err, w, taskExecution, tempDirPath)
			return
		}

		actualOutputs, err := getOutputs(tempDirPath)
		if err != nil {
			log.Error(err)
			handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
			return
		}

		var hasArtefacts bool = len(test.ArtefactSHA256) != 0

		if hasArtefacts {
			hashMatches, err := checkHashMatch(test, tempDirPath)
			if err != nil {
				log.Error(err)
				handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
				return
			} else if !hashMatches {
				sendTaskExecutionFailedResponse(w, EXEC_ERR_ARTEFACT_CONTENT_MISMATCH, testDataMismatchReason{TestInput: testInput})
				setTaskExecutionStatusFailed(taskExecution)
				omitOutputsCheck = true
			}
		}

		os.RemoveAll(tempDirPath)

		if omitOutputsCheck {
			return
		}

		if actualOutputs != test.ExpectedOutput {
			sendTaskExecutionFailedResponse(w, EXEC_ERR_TEST_FAILED, testDataMismatchReason{
				TestInput:      testInput,
				Output:         actualOutputs,
				ExpectedOutput: test.ExpectedOutput,
			})
			setTaskExecutionStatusFailed(taskExecution)
			return
		}
	}

	res = taskExecutionResponse{Success: true, Message: "Test data matches output!"}
	sendResponse(w, res)
	setTaskExecutionStatusSucceeded(taskExecution)
}

func checkIsValidSolutionCode(cppFile *os.File) bool {
	cmd := exec.Command("g++", "-fsyntax-only", cppFile.Name())
	_, err := cmd.CombinedOutput()
	return err == nil
}

func markTaskExecutionStartForUserId(taskId, userId int, code string) (*database.TaskExecution, error) {
	var taskExecution database.TaskExecution = database.TaskExecution{
		InitiatorID:   userId,
		TaskID:        taskId,
		StartedAt:     time.Now(),
		IsFinished:    false,
		SubmittedCode: code,
		WasSuccessful: false,
	}

	return database.SaveTaskExecution(taskExecution)
}

func setTaskExecutionStatusFailed(taskExecution *database.TaskExecution) {
	taskExecution.WasSuccessful = false
	saveFinishedTaskExecution(taskExecution)
}

func setTaskExecutionStatusSucceeded(taskExecution *database.TaskExecution) {
	taskExecution.WasSuccessful = true
	saveFinishedTaskExecution(taskExecution)
}

func saveFinishedTaskExecution(taskExecution *database.TaskExecution) {
	taskExecution.IsFinished = true
	taskExecution.FinishedAt = time.Now()
	database.SaveTaskExecution(*taskExecution)
}

func checkUserHasRunningTasks(userId int) bool {
	currentlyRunningTaskExecution := database.GetRunningTaskExecutionForUserId(userId)
	return currentlyRunningTaskExecution != nil
}

func sendTaskExecutionFailedResponse(w http.ResponseWriter, execErrCode ExecutionErrorCode, reasonFailed interface{}) {
	res := taskExecutionResponse{Success: false, ErrorCode: execErrCode.EnumIndex(), Message: execErrCode.String(), ReasonFailed: reasonFailed}
	sendResponse(w, res)
}

func handleTestExecutionInternalFail(w http.ResponseWriter, tempDirPath string, err error, taskExecution *database.TaskExecution) {
	api.InternalErrorHandlerGenericMsg(w, err)
	os.RemoveAll(tempDirPath)
	setTaskExecutionStatusFailed(taskExecution)
}

func checkHashMatch(test database.Test, tempDirPath string) (hashMatches bool, err error) {
	actualSha256, err := calculateOutputArtefactsHash(tempDirPath)
	if err != nil {
		return false, err
	}

	if actualSha256 != test.ArtefactSHA256 {
		return false, nil
	}

	return true, nil
}

func sendResponse(w http.ResponseWriter, res taskExecutionResponse) {
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func getRequestBody(r *http.Request) (*taskExecutionRequest, error) {
	if r.Body == nil {
		return nil, errors.New("body is missing task's data")
	}

	var requestBody taskExecutionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil || requestBody.SolutionCode == "" {
		return nil, errors.New("body is not in correct format")
	}
	return &requestBody, nil
}

// Stores temp CPP source code file and "stdin.txt" file which serves as stdin mock.
// If it fails, the function deletes whatever it created.
// Returns: new CPP file
func storeTempFiles(cppCode string, inputs string) (*os.File, error) {
	createdTempPath, _ := os.MkdirTemp("/var/temp_tasks/", "temp_cpp_solutions_*")
	createdTempCppFile, err := createTempCppFile(createdTempPath, cppCode)
	if err != nil {
		return nil, err
	}

	var fullPathToNewTempDir = filepath.Dir(createdTempCppFile.Name())

	createdTempInputsFile, err := os.Create(fullPathToNewTempDir + "/stdin.txt")
	if err != nil {
		log.Error("Failed to create temp inputs file!", err)
		os.Remove(createdTempCppFile.Name())
		return nil, err
	}
	err = fillFileWithData(createdTempInputsFile, inputs)
	if err != nil {
		os.Remove(createdTempCppFile.Name())
		os.Remove(createdTempInputsFile.Name())
		return nil, err
	}

	return createdTempCppFile, nil
}

// No file gets created if I fail.
func createTempCppFile(path string, code string) (*os.File, error) {
	createdTempCppFile, err := os.CreateTemp(path, "solution_*.cpp")

	if err != nil {
		log.Error("Failed to create temp cpp file!", err)
		return nil, err
	}

	err = fillFileWithData(createdTempCppFile, code)
	if err != nil {
		os.Remove(createdTempCppFile.Name())
		return nil, err
	}

	return createdTempCppFile, nil
}

func fillFileWithData(tempFile *os.File, inputs string) error {
	_, err := tempFile.Write([]byte(inputs))
	if err != nil {
		log.Errorf("Failed to insert data in the temp file (%s)\n%s", tempFile.Name(), err)
	}
	return err
}

func runFileInIsolatedDockerContainerTask(context context.Context, cppFile *os.File) error {
	errChan := make(chan error, 1)
	sourceCodePath := cppFile.Name()

	go func() {
		errChan <- runDockerRunnerImage(sourceCodePath, true)
	}()

	select {
	case <-context.Done():
		var containerName = filepath.Base(filepath.Dir(sourceCodePath))

		err := removeRunningDockerContainer(containerName)
		if err != nil {
			log.Error(err)
		}

		return errors.New(CONTAINER_TIMEOUT_MARK)
	case err := <-errChan:
		log.Errorf("Failed to run Docker container: %v", err)
		return err
	}
}

// Names the new container after the temp parent folder in which the CPP file is stored.
func runDockerRunnerImage(fullFilePath string, allowBuildingImageIfNotFound bool) error {
	var sourceCodePath = filepath.Dir(fullFilePath)
	var parentFolderName = filepath.Base(sourceCodePath)
	var sourceFileName = filepath.Base(fullFilePath)

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	var volumeName = os.Getenv("TASKS_VOLUME_NAME")
	var containerName = parentFolderName

	var dockerRunArguments = fmt.Sprintf(
		"%s run --rm "+
			"--name %s "+
			"-v %s:/var/temp_tasks/ "+
			"--memory 30m --cpus 0.15 "+
			"--security-opt no-new-privileges --network none "+
			"-e SOURCE_CODE_FOLDER=%s "+
			"-e SOURCE_FILE_NAME=%s "+
			"task-runner:latest", dockerPath, containerName, volumeName, parentFolderName, sourceFileName,
	)

	cmd := exec.Command("bash", "-c", dockerRunArguments)

	output, err := cmd.CombinedOutput()
	var stringOutput = string(output)
	if stringOutput != "" {
		if strings.Contains(stringOutput, "Unable to find image 'task-runner:latest'") {
			if allowBuildingImageIfNotFound {
				buildRunnerImage(dockerPath)
				return runDockerRunnerImage(fullFilePath, false)
			} else {
				log.Errorf("Failed to build Docker container: %s", err)
			}
		} else {
			log.Warningf("Suspicious output from task-runner container: %s", stringOutput)
		}
	}

	return err
}

func buildRunnerImage(dockerPath string) error {
	cmd := exec.Command("bash", "-c", dockerPath+" build -t task-runner:latest -f ./task-runner.Dockerfile .")
	return cmd.Run()
}

func handleDockerContainerFail(err error, w http.ResponseWriter, taskExecution *database.TaskExecution, tempDirPath string) {
	switch err.Error() {
	case CONTAINER_TIMEOUT_MARK:
		{
			sendTaskExecutionFailedResponse(w, EXEC_ERR_TIMEOUT, nil)
			setTaskExecutionStatusFailed(taskExecution)
		}
	default:
		{
			log.Error(err)
			handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
		}
	}
}

func removeRunningDockerContainer(containerName string) error {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(dockerPath+" rm --force "+containerName))
	output, err := cmd.CombinedOutput()
	var stringOutputs = string(output)
	if stringOutputs != containerName {
		if err == nil {
			return fmt.Errorf("container %s could not be stopped: %s", containerName, stringOutputs)
		} else {
			return fmt.Errorf(fmt.Sprintf("Container %s could not be stopped: error:%v, output:%s", containerName, err, stringOutputs))
		}
	}

	return nil
}

func getOutputs(tempDirPath string) (string, error) {
	bytes, err := os.ReadFile(filepath.Join(tempDirPath, "stdout.txt"))
	var stringOutput = string(bytes)
	stringOutput = strings.TrimFunc(stringOutput, func(r rune) bool { return r == '\n' || r == ' ' })
	return stringOutput, err
}

func calculateOutputArtefactsHash(tempDirPath string) (string, error) {
	var rawContents, err = readOutputFilesFromDir(tempDirPath)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(rawContents)
	var hashedContents string = fmt.Sprintf("%x", h.Sum(nil))

	return hashedContents, err
}

func readOutputFilesFromDir(tempDirPath string) ([]byte, error) {
	artefactsFilePath := filepath.Join(tempDirPath, "artefacts.txt")
	bytes, errMain := os.ReadFile(artefactsFilePath)
	var stringOutput = string(bytes)
	var fileNames = strings.Split(stringOutput, "\n")
	sort.Strings(fileNames)

	var allBytes []byte = make([]byte, 0)

	for _, relativeFileName := range fileNames {
		if len(relativeFileName) == 0 {
			continue
		}

		var actualFileName = filepath.Base(relativeFileName)
		fullFilePath := filepath.Join(tempDirPath, actualFileName)

		currentFileBytes, err := os.ReadFile(fullFilePath)
		if err != nil {
			errMain = err
			break
		}
		allBytes = append(allBytes, currentFileBytes...)
	}

	return allBytes, errMain
}

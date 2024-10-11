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
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/lizard"
	log "github.com/sirupsen/logrus"
)

var tasksVolumeName = os.Getenv("TASKS_VOLUME_NAME")
var tempTasksFolderPath = "/var/temp_tasks/"

type ExecutionErrorCode int

const CONTAINER_TIMEOUT_MARK = "timeout"
const CONTAINER_FORCEFULLY_KILLED_MARK = "forcefully killed"

const (
	EXEC_ERR_BAD_SYNTAX ExecutionErrorCode = iota + 1
	EXEC_ERR_TEST_FAILED
	EXEC_ERR_TIMEOUT
	EXEC_ERR_KILLED
	EXEC_ERR_ARTEFACT_CONTENT_MISMATCH
)

func (execErrCode ExecutionErrorCode) String() string {
	return [...]string{
		"Solution code has syntax errors.",
		"Program did not output expected test data.",
		"Execution took too long.",
		"Execution was forcefully killed. Most probably a memory leak.",
		"Artefact files did not contain expected contents.",
	}[execErrCode-1]
}

func (execErrCode ExecutionErrorCode) EnumIndex() int {
	return int(execErrCode)
}

type taskExecutionRequest struct {
	SolutionCode string `json:"solutionCode"`
}

type testDataMismatchReason struct {
	TestInput      string `json:"testInput,omitempty"`
	Output         string `json:"output,omitempty"`
	ExpectedOutput string `json:"expectedOutput,omitempty"`
}

type taskExecutionFailedResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	ErrorCode    int         `json:"errorCode"`
	ReasonFailed interface{} `json:"reason,omitempty"`
}

type successfulTaskExecutionResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Score   lizard.CodeScore `json:"score"`
}

func executeTask(w http.ResponseWriter, r *http.Request) {
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

	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil || userId == 0 {
		api.InternalErrorHandlerCustomMsg(w, "Couldn't get user from JWT token.")
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't get user from JWT token during task execution.")
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
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't mark execution as started!")
		return
	}

	cppFileForSyntaxChecking, err := createTempCppFile("", solutionCode)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		setTaskExecutionStatusFailed(taskExecution)
		return
	}

	reportedErrors, err := findAllErrorsInSolutionCode(cppFileForSyntaxChecking)
	os.Remove(cppFileForSyntaxChecking.Name())
	if err != nil {
		api.InternalErrorHandlerCustomMsg(w, fmt.Sprintf("Couldn't check code syntax: %v", err))
		setTaskExecutionStatusFailed(taskExecution)
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't check code syntax!")
		return
	} else if len(reportedErrors) != 0 {
		sendTaskExecutionFailedResponse(taskExecution, w, EXEC_ERR_BAD_SYNTAX, reportedErrors)
		return
	}

	var tests []database.TaskTest = database.GetTestsForTask(taskId)
	if len(tests) == 0 {
		api.InternalErrorHandlerCustomMsg(w, fmt.Sprintf("No tests defined for task %d!", taskId))
		setTaskExecutionStatusFailed(taskExecution)
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't test the solution!")
		return
	}

	for _, test := range tests {
		omitOutputsCheck := false
		testInput := test.Input

		cppFile, err := storeTempFiles(solutionCode, testInput)
		var tempDirPath = filepath.Dir(cppFile.Name())
		if err != nil {
			handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Couldn't store temp files during testing!")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = runFileInIsolatedDockerContainerTask(ctx, cppFile)
		if err != nil {
			switch err.Error() {
			case CONTAINER_TIMEOUT_MARK:
				{
					sendTaskExecutionFailedResponse(taskExecution, w, EXEC_ERR_TIMEOUT, nil)
					break
				}
			case CONTAINER_FORCEFULLY_KILLED_MARK:
				{
					sendTaskExecutionFailedResponse(taskExecution, w, EXEC_ERR_KILLED, nil)
					break
				}
			default:
				{
					handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
					log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId, "temp_dir": tempDirPath}).Error("Task runner failed to run in Docker!")
				}
			}
			return
		}

		outputPath := filepath.Join(tempDirPath, "stdout.txt")
		actualOutputs, err := getOutputs(outputPath)
		if err != nil {
			handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId, "file": outputPath}).Error("Couldn't read output file!")
			return
		}

		var hasArtefacts bool = test.ArtefactSHA256.Valid

		if hasArtefacts {
			hashMatches, err := checkHashMatch(test, tempDirPath)
			if err != nil {
				handleTestExecutionInternalFail(w, tempDirPath, err, taskExecution)
				log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Artefact SHA256 comparison failed!")
				return
			} else if !hashMatches {
				sendTaskExecutionFailedResponse(taskExecution, w, EXEC_ERR_ARTEFACT_CONTENT_MISMATCH, testDataMismatchReason{TestInput: testInput})
				omitOutputsCheck = true
			}
		}

		os.RemoveAll(tempDirPath)

		if omitOutputsCheck {
			return
		}

		if actualOutputs != test.ExpectedOutput {
			sendTaskExecutionFailedResponse(taskExecution, w, EXEC_ERR_TEST_FAILED, testDataMismatchReason{
				TestInput:      testInput,
				Output:         actualOutputs,
				ExpectedOutput: test.ExpectedOutput,
			})
			return
		}
	}

	cppFileForScoreCalculation, err := createTempCppFile("", solutionCode)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		setTaskExecutionStatusFailed(taskExecution)
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Artefact SHA256 comparison failed!")
		return
	}

	var solvedTask = database.GetSingleFullTasks(taskId)
	numbericComplexity, err := strconv.Atoi(solvedTask.BasicInfo.Complexity)
	if err != nil {
		numbericComplexity = 0
		log.WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Task complexity could not be converted to integer!")
	}

	score, err := lizard.CalculateScore(cppFileForScoreCalculation, numbericComplexity)
	if err != nil {
		api.InternalErrorHandlerCustomMsg(w, fmt.Sprintf("Could not calculate score: %v", err))
		setTaskExecutionStatusFailed(taskExecution)
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Could not calculate score after task execution.")
		return
	}
	err = os.Remove(cppFileForScoreCalculation.Name())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Could not delete temp file created for scoring!")
	}

	setTaskExecutionStatusSucceeded(taskExecution, solvedTask, *score)

	var successResponse = successfulTaskExecutionResponse{Success: true, Score: *score, Message: "Test data matches output!"}
	api.RespondOk(w, successResponse)
}

func findAllErrorsInSolutionCode(cppFile *os.File) ([]utils.GppCompilerReportedSyntaxError, error) {
	cmd := exec.Command("g++", "-fsyntax-only", "-o /dev/null", cppFile.Name())
	output, err := cmd.CombinedOutput()

	var foundErrors []utils.GppCompilerReportedSyntaxError

	if err != nil {
		foundErrors, err = utils.ExtractGppSyntaxErrors(string(output))
	}

	return foundErrors, err
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

func setTaskExecutionStatusSucceeded(taskExecution *database.TaskExecution, solvedTask *database.FullTask, score lizard.CodeScore) {
	taskExecution.WasSuccessful = true
	taskExecution.CodeScore = &score

	var previousBestScoreExecutionFromThisUserForThisTask = database.GetBestScoreExecutionForUserAndTask(
		taskExecution.InitiatorID,
		taskExecution.TaskID,
	)

	var newUsersBestScore = false
	var newTaskAllTimeBestScore = false

	if previousBestScoreExecutionFromThisUserForThisTask == nil {
		newUsersBestScore = true
	} else {
		if taskExecution.CodeScore.HasBetterScoreThan(previousBestScoreExecutionFromThisUserForThisTask.CodeScore) {
			subtractLastScore(solvedTask, *previousBestScoreExecutionFromThisUserForThisTask.CodeScore)
			newUsersBestScore = true
		}
	}

	if newUsersBestScore {
		taskExecution.BestScore = true
		increaseAverageScoreOnTaskItself(solvedTask, score)

		if previousBestScoreExecutionFromThisUserForThisTask != nil {
			previousBestScoreExecutionFromThisUserForThisTask.BestScore = false
			database.SaveTaskExecution(*previousBestScoreExecutionFromThisUserForThisTask)
		}
	}

	if taskExecution.CodeScore.HasBetterScoreThan(&solvedTask.AllTimeBestScore) {
		solvedTask.AllTimeBestScore = *taskExecution.CodeScore
		newTaskAllTimeBestScore = true
	}

	if newUsersBestScore || newTaskAllTimeBestScore {
		database.SaveTask(solvedTask)
	}

	saveFinishedTaskExecution(taskExecution)
}

func subtractLastScore(fullTask *database.FullTask, score lizard.CodeScore) {
	scoresCountSoFar, tokensSum, scoresSum, complexitySum := getTaskScoreSums(fullTask)

	tokensSum -= score.Tokens
	scoresSum -= score.TotalScore
	complexitySum -= score.Complexity
	scoresCountSoFar--

	setAverageTaskScore(fullTask, scoresCountSoFar, tokensSum, scoresSum, complexitySum)
}

func increaseAverageScoreOnTaskItself(fullTask *database.FullTask, score lizard.CodeScore) {
	scoresCountSoFar, tokensSum, scoresSum, complexitySum := getTaskScoreSums(fullTask)

	tokensSum += score.Tokens
	scoresSum += score.TotalScore
	complexitySum += score.Complexity
	scoresCountSoFar++

	setAverageTaskScore(fullTask, scoresCountSoFar, tokensSum, scoresSum, complexitySum)
}

func getTaskScoreSums(task *database.FullTask) (int, int, float32, int) {
	var scoresCountSoFar = task.ScoresCount
	var tokensSum = task.AverageScore.Tokens * scoresCountSoFar
	var scoresSum = task.AverageScore.TotalScore * float32(scoresCountSoFar)
	var complexitySum = task.AverageScore.Complexity * scoresCountSoFar
	return scoresCountSoFar, tokensSum, scoresSum, complexitySum
}

func setAverageTaskScore(task *database.FullTask, newScoresCount int, tokensSum int, scoresSum float32, complexitySum int) {
	task.ScoresCount = newScoresCount
	if newScoresCount == 0 {
		task.AverageScore.Tokens = 0
		task.AverageScore.TotalScore = 0
		task.AverageScore.Complexity = 0
	} else {
		task.AverageScore.Tokens = tokensSum / newScoresCount
		task.AverageScore.TotalScore = utils.RoundNumberDownToTwoDecimals(scoresSum / float32(newScoresCount))
		task.AverageScore.Complexity = complexitySum / newScoresCount
	}
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

func sendTaskExecutionFailedResponse(taskExecution *database.TaskExecution, w http.ResponseWriter, execErrCode ExecutionErrorCode, reasonFailed interface{}) {
	setTaskExecutionStatusFailed(taskExecution)
	var errorResponse = taskExecutionFailedResponse{Success: false, ErrorCode: execErrCode.EnumIndex(), Message: execErrCode.String(), ReasonFailed: reasonFailed}

	api.RespondWithStatus(w, errorResponse, http.StatusUnprocessableEntity)
}

func handleTestExecutionInternalFail(w http.ResponseWriter, tempDirPath string, err error, taskExecution *database.TaskExecution) {
	api.InternalErrorHandlerGenericMsg(w, err)
	os.RemoveAll(tempDirPath)
	setTaskExecutionStatusFailed(taskExecution)
}

func checkHashMatch(test database.TaskTest, tempDirPath string) (hashMatches bool, err error) {
	actualSha256, err := calculateOutputArtefactsHash(tempDirPath)
	if err != nil {
		return false, err
	}

	if actualSha256 != test.ArtefactSHA256.String {
		return false, nil
	}

	return true, nil
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
	createdTempPath, _ := os.MkdirTemp(tempTasksFolderPath, "temp_cpp_solutions_*")
	createdTempCppFile, err := createTempCppFile(createdTempPath, cppCode)
	if err != nil {
		return nil, err
	}

	var fullPathToNewTempDir = filepath.Dir(createdTempCppFile.Name())

	createdTempInputsFile, err := os.Create(fullPathToNewTempDir + "/stdin.txt")
	if err != nil {
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
	return err
}

func runFileInIsolatedDockerContainerTask(timeoutContext context.Context, cppFile *os.File) error {
	var errChan = make(chan error, 1)
	var sourceCodePath = cppFile.Name()
	var containerName = utils.CreateShortHash(sourceCodePath)

	log.Debugf("Source code path: %s", sourceCodePath)

	go func() {
		errChan <- runDockerRunnerImage(containerName, sourceCodePath, true)
	}()

	select {
	case <-timeoutContext.Done():
		log.WithFields(log.Fields{"context": "task_execution"}).Warning("Timeout reached - forcefully removing Docker container!")

		err := removeRunningDockerContainer(containerName)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution"}).Error("Could not delete temp container after task execution timeout!")
		}

		return errors.New(CONTAINER_TIMEOUT_MARK)
	case err := <-errChan:
		return err
	}
}

func runDockerRunnerImage(containerName string, fullFilePath string, allowBuildingImageIfNotFound bool) error {
	var sourceCodePath = filepath.Dir(fullFilePath)
	var sourceFileName = filepath.Base(fullFilePath)

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	var volumeName = ""
	var parentFolderName = ""
	if utils.IsProd() {
		volumeName = tasksVolumeName
		parentFolderName = filepath.Base(sourceCodePath)
	} else {
		volumeName = sourceCodePath
		parentFolderName = ""
	}

	var dockerRunArguments = fmt.Sprintf(
		"%s run --rm "+
			"--name %s "+
			"-v %s:%s "+
			"--memory 50m --cpus 0.15 "+
			"--security-opt no-new-privileges --network none "+
			"-e SOURCE_CODE_FOLDER=%s "+
			"-e SOURCE_FILE_NAME=%s "+
			"task-runner:latest", dockerPath, containerName, volumeName, tempTasksFolderPath, parentFolderName, sourceFileName,
	)

	cmd := exec.Command("bash", "-c", dockerRunArguments)

	output, err := cmd.CombinedOutput()
	var stringOutput = string(output)
	if stringOutput != "" {
		if strings.Contains(stringOutput, "Unable to find image 'task-runner:latest'") {
			if allowBuildingImageIfNotFound {
				buildRunnerImage(dockerPath)
				return runDockerRunnerImage(containerName, fullFilePath, false)
			} else {
				log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution"}).Error("Failed to build Docker container!")
			}
		} else if strings.Contains(stringOutput, "Killed") {
			return errors.New(CONTAINER_FORCEFULLY_KILLED_MARK)
		} else {
			log.WithFields(log.Fields{"suspicious_output": stringOutput, "context": "task_execution"}).Warning("Suspicious output from task-runner container!")
		}
	}

	return err
}

func buildRunnerImage(dockerPath string) error {
	var buildCommand string = fmt.Sprintf("%s build -t task-runner:latest -f ", dockerPath)
	if utils.IsProd() {
		buildCommand += "./task-runner.Dockerfile ."
	} else {
		buildCommand += "./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/"
	}

	log.Infof("Building Docker image using command: %s", buildCommand)

	cmd := exec.Command("bash", "-c", buildCommand)
	return cmd.Run()
}

func removeRunningDockerContainer(containerName string) error {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf(dockerPath+" rm --force "+containerName))
	output, err := cmd.CombinedOutput()
	var stringOutputs = strings.Trim(string(output), "\n")
	if stringOutputs != containerName {
		if err == nil {
			return fmt.Errorf("container %s could not be stopped: %s", containerName, stringOutputs)
		} else {
			return fmt.Errorf("container %s could not be stopped: error:%v, output:%s", containerName, err, stringOutputs)
		}
	}

	return nil
}

func getOutputs(outputFilePath string) (string, error) {
	bytes, err := os.ReadFile(outputFilePath)
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

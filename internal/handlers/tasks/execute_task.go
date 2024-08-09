package tasks

import (
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

	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	log "github.com/sirupsen/logrus"
)

type taskExecutionRequest = struct {
	SolutionCode string `json:"solution_code"`
}

type testDataMismatch = struct {
	TestInput      string `json:"test_input"`
	Output         string `json:"output,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
}

type taskExecutionResponse = struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	ReasonFailed *testDataMismatch `json:"reason,omitempty"`
}

func ExecuteTask(w http.ResponseWriter, r *http.Request) {
	requestBody, failedToGetRequestBody := getRequestBody(r, w)
	if failedToGetRequestBody {
		return
	}

	var originalParamId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalParamId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	var tests []database.Test = database.GetTestsForTask(taskId)
	if len(tests) == 0 {
		res := taskExecutionResponse{Success: false, Message: "Can't test this task."}
		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(res)
		return
	}

	var res taskExecutionResponse

	for _, test := range tests {
		omitOutputsCheck := false
		testInput := test.Input

		cppFile, err := storeTempFiles(requestBody.SolutionCode, testInput)
		var tempDirPath = filepath.Dir(cppFile.Name())
		if err != nil {
			handleTestExecutionFail(w, tempDirPath, err)
			return
		}

		err = runFileInIsolatedDockerContainer(cppFile)
		if err != nil {
			log.Error(err)
			handleTestExecutionFail(w, tempDirPath, err)
			return
		}

		actualOutputs, err := getOutputs(tempDirPath)
		if err != nil {
			log.Error(err)
			handleTestExecutionFail(w, tempDirPath, err)
			return
		}

		var hasArtefacts bool = len(test.ArtefactSHA256) != 0

		if hasArtefacts {
			hashMatches, err := checkHashMatch(test, tempDirPath)
			if err != nil {
				log.Error(err)
				handleTestExecutionFail(w, tempDirPath, err)
				return
			} else if !hashMatches {
				res = taskExecutionResponse{
					Success:      false,
					Message:      "Artefact files did not contain expected contents.",
					ReasonFailed: &testDataMismatch{TestInput: testInput},
				}
				sendResponse(w, res)
				omitOutputsCheck = true
			}
		}

		os.RemoveAll(tempDirPath)

		if omitOutputsCheck {
			return
		}

		if actualOutputs != test.ExpectedOutput {
			res = taskExecutionResponse{
				Success: false,
				Message: "Program did not output expected test data.",
				ReasonFailed: &testDataMismatch{
					TestInput:      testInput,
					Output:         actualOutputs,
					ExpectedOutput: test.ExpectedOutput,
				},
			}
			sendResponse(w, res)
			return
		}
	}

	res = taskExecutionResponse{Success: true, Message: "Test data matches output!", ReasonFailed: nil}
	sendResponse(w, res)
}

func handleTestExecutionFail(w http.ResponseWriter, tempDirPath string, err error) {
	api.InternalErrorHandlerGenericMsg(w, err)
	os.RemoveAll(tempDirPath)
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

func getRequestBody(r *http.Request, w http.ResponseWriter) (*taskExecutionRequest, bool) {
	if r.Body == nil {
		api.RequestErrorHandlerCustomMsg(w, "Body is missing task's data!")
		return nil, true
	}

	var requestBody taskExecutionRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil || requestBody.SolutionCode == "" {
		api.RequestErrorHandlerCustomMsg(w, "Body is not in correct format!")
		return nil, true
	}
	return &requestBody, false
}

// Stores temp CPP source code file and "stdin.txt" file which serves as stdin mock.
// If it fails, the function deletes whatever it created.
// Returns: new CPP file
func storeTempFiles(cppCode string, inputs string) (*os.File, error) {
	createdTempPath, _ := os.MkdirTemp("/var/temp_tasks/", "temp_cpp_solutions_*")

	createdTempCppFile, err := os.CreateTemp(createdTempPath, "solution_*.cpp")
	if err != nil {
		log.Error("Failed to create temp cpp file!", err)
		return nil, err
	}
	err = fillFileWithData(createdTempCppFile, cppCode)
	if err != nil {
		os.Remove(createdTempCppFile.Name())
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

func fillFileWithData(tempFile *os.File, inputs string) error {
	_, err := tempFile.Write([]byte(inputs))
	if err != nil {
		log.Errorf("Failed to insert data in the temp file (%s)\n%s", tempFile.Name(), err)
	}
	return err
}

func runFileInIsolatedDockerContainer(cppFile *os.File) error {
	err := runDockerRunnerImage(cppFile.Name())

	if err != nil {
		if err.Error() == "image not built" {
			log.Warn("Image wasn't built!")
			buildErr := buildRunnerImage()
			if buildErr != nil {
				log.Errorf("Failed to build Docker container: %s", err)
				return err
			}

			err = runDockerRunnerImage(cppFile.Name())
		}

		if err != nil {
			log.Errorf("Failed to run Docker container: %v", err)
			return err
		}
	}

	return nil
}

func buildRunnerImage() error {
	log.Info("Building task-runner image.")
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-c", dockerPath+" build -t task-runner:latest -f ./task-runner.Dockerfile .")
	return cmd.Run()
}

func runDockerRunnerImage(fullFilePath string) error {
	var sourceCodePath = filepath.Dir(fullFilePath)
	var parentFolderName = filepath.Base(sourceCodePath)
	var sourceFileName = filepath.Base(fullFilePath)

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	var volumeName = os.Getenv("TASKS_VOLUME_NAME")

	var dockerRunArguments = fmt.Sprintf(
		"%s run --rm "+
			"-v %s:/var/temp_tasks/ "+
			"--memory 30m --cpus 0.15 "+
			"--security-opt no-new-privileges --network none "+
			"-e SOURCE_CODE_FOLDER=%s "+
			"-e SOURCE_FILE_NAME=%s "+
			"task-runner:latest", dockerPath, volumeName, parentFolderName, sourceFileName,
	)

	cmd := exec.Command("bash", "-c", dockerRunArguments)

	output, err := cmd.CombinedOutput()
	var stringOutput = string(output)
	if stringOutput != "" {
		if strings.Contains(stringOutput, "Unable to find image 'task-runner:latest'") {
			return errors.New("image not built")
		} else {
			log.Warningf("Suspicious output from task-runner container: %s", stringOutput)
		}
	}

	return err
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

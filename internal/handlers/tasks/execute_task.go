package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	api "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/errors"
	log "github.com/sirupsen/logrus"
)

type taskExecutionRequest = struct {
	SolutionCode string `json:"solution_code"`
}

type taskExecutionResponse = struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func ExecuteTask(w http.ResponseWriter, r *http.Request) {
	requestBody, failedToGetRequestBody := getRequestBody(r, w)
	if failedToGetRequestBody {
		return
	}

	outputBytes, err := compileCppToJs(requestBody.SolutionCode)
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	readableOutput := string(outputBytes)

	res := taskExecutionResponse{Success: true, Message: string(readableOutput)}

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

func compileCppToJs(cppCode string) ([]byte, error) {
	tempFile, err := createTempCppFile(cppCode)
	if err != nil {
		return nil, err
	}

	jsCode, err := useEmscriptenConversion(tempFile)
	if err != nil {
		return nil, err
	}

	return jsCode, nil
}

func createTempCppFile(fileContents string) (*os.File, error) {
	createdTempPath, _ := os.MkdirTemp("", "temp_cpp_solutions")
	createdTempFile, err := os.CreateTemp(createdTempPath, "solution_*.cpp")
	if err != nil {
		log.Error("Failed to create temp cpp file!\n", err)
		return nil, err
	}

	_, err = createdTempFile.Write([]byte(fileContents))
	if err != nil {
		log.Error("Failed to insert data in the temp cpp file ("+fileContents+")\n", err)
		return nil, err
	}

	return createdTempFile, nil
}

func useEmscriptenConversion(tempCppFile *os.File) (javascript []byte, err error) {
	var cmd *exec.Cmd

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, err
	}
	idPath, err := exec.LookPath("id")
	if err != nil {
		return nil, err
	}

	cmd = exec.Command(idPath, "-u")
	idU, _ := cmd.Output()
	cmd = exec.Command(idPath, "-g")
	idG, _ := cmd.Output()

	var mappedUsers string = strings.Split(string(idU), "\n")[0] + ":" + strings.Split(string(idG), "\n")[0]

	pureFileName, _ := filepath.Abs(tempCppFile.Name())
	tempFilePath := filepath.Dir(tempCppFile.Name())

	var dockerEmscriptenArguments = dockerPath + " run --rm " +
		"-v " + tempFilePath + ":" + tempFilePath + " " +
		"-u " + mappedUsers + " " +
		"emscripten/emsdk:3.1.64 emcc " +
		pureFileName + " -o " + pureFileName + ".js"

	cmd = exec.Command("bash", "-c", dockerEmscriptenArguments)
	var stdError bytes.Buffer
	cmd.Stderr = &stdError
	_, err = cmd.Output()
	if err != nil {
		log.Error("Docker reported the following error: "+err.Error(), ", with standard output saying: "+stdError.String())
		return nil, err
	}

	outputFileName := tempCppFile.Name() + ".js"
	javascriptContents, err := os.ReadFile(outputFileName)
	if err != nil {
		log.Error("Couldn't read javascript file at: " + outputFileName)
		return nil, err
	}

	return javascriptContents, nil
}

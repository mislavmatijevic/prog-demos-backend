package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Outputs string `json:"outputs,omitempty"`
	Errors  string `json:"errors,omitempty"`
}

type CompileResponse struct {
	Outputs string `json:"outputs"`
	Errors  string `json:"errors"`
}

func ExecuteTask(w http.ResponseWriter, r *http.Request) {
	requestBody, failedToGetRequestBody := getRequestBody(r, w)
	if failedToGetRequestBody {
		return
	}

	tempFileName, err := compileCppToJs(requestBody.SolutionCode)
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	response, err := executeCompiledCodeInInsolate(tempFileName)
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	res := taskExecutionResponse{Success: true, Outputs: response.Outputs, Errors: response.Errors}

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

func compileCppToJs(cppCode string) (tempCppFileNameWithoutExt *os.File, err error) {
	tempCppSourceFile, err := createTempCppFile(cppCode)
	if err != nil {
		return nil, err
	}

	err = useEmscriptenConversion(tempCppSourceFile)
	if err != nil {
		return nil, err
	}

	return tempCppSourceFile, nil
}

func createTempCppFile(fileContents string) (*os.File, error) {
	createdTempPath, _ := os.MkdirTemp("/var/temp_solutions/", "temp_cpp_solutions")
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

func useEmscriptenConversion(tempCppFile *os.File) (err error) {
	var cmd *exec.Cmd

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return err
	}
	idPath, err := exec.LookPath("id")
	if err != nil {
		return err
	}

	cmd = exec.Command(idPath, "-u")
	idU, _ := cmd.Output()
	cmd = exec.Command(idPath, "-g")
	idG, _ := cmd.Output()

	var mappedUsers string = strings.Split(string(idU), "\n")[0] + ":" + strings.Split(string(idG), "\n")[0]

	pureFileName, _ := filepath.Abs(tempCppFile.Name())

	var dockerEmscriptenArguments = dockerPath + " run --rm --memory=100m " +
		"-v prog-demos-backend_solutions:/var/temp_solutions/ " +
		"-u " + mappedUsers + " " +
		"emscripten/emsdk:3.1.64 " +
		"emcc " + pureFileName + " -s ENVIRONMENT=shell -o " + pureFileName + ".js"

	cmd = exec.Command("bash", "-c", dockerEmscriptenArguments)
	var stdError bytes.Buffer
	cmd.Stderr = &stdError
	_, err = cmd.Output()
	if err != nil {
		log.Error("Docker reported the following error: "+err.Error(), ", with standard output saying: "+stdError.String())
		return err
	}

	return nil
}

func executeCompiledCodeInInsolate(tempFile *os.File) (*CompileResponse, error) {
	executionPayload := map[string]string{
		"jsFileName":   fmt.Sprintf("%s.js", tempFile.Name()),
		"wasmFileName": fmt.Sprintf("%s.wasm", tempFile.Name()),
		"tests":        "3 7-4 5",
	}

	payloadBytes, err := json.Marshal(executionPayload)
	if err != nil {
		log.Error("Failed to marshal execution payload", err)
		return nil, err
	}

	var taskRunnerUrl = os.Getenv("TASK-RUNNER-URL")
	resp, err := http.Post(taskRunnerUrl, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Error("Failed to send request to Node.js service", err)
		return nil, err
	}
	defer resp.Body.Close()

	var compileRes CompileResponse
	err = json.NewDecoder(resp.Body).Decode(&compileRes)
	if err != nil {
		log.Error("Failed to decode response from Node.js service", err)
		return nil, err
	}

	return &compileRes, nil
}

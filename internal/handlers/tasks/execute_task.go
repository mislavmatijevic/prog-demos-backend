package tasks

import (
	"encoding/json"
	"net/http"
	"os"

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
	err := createTempCppFile(cppCode)
	if err != nil {
		return nil, err
	}

	return []byte("ok"), nil
}

func createTempCppFile(fileContents string) error {
	path, _ := os.MkdirTemp("", "temp_cpp_solutions")
	f, err := os.CreateTemp(path, "solution_*.cpp")
	if err != nil {
		log.Error("Failed to create temp cpp file!\n", err)
		return err
	}

	_, err = f.Write([]byte(fileContents))
	if err != nil {
		log.Error("Failed to insert data in the temp cpp file ("+fileContents+")\n", err)
		return err
	}

	return nil
}

package tasks

import (
	"encoding/json"
	"net/http"

	api "github.com/mislavmatijevic/prog-demos-backend/internal/handlers/errors"
	"github.com/sirupsen/logrus"
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
	logrus.Info(readableOutput)

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
	var outputBytes []byte = []byte(cppCode)
	// cmd := exec.Command("docker")
	// outputBytes, err := cmd.Output()
	// if err != nil {
	// 	return nil, err
	// }
	return outputBytes, nil
}

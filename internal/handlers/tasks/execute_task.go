package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/lizard"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/taskexecution"
	log "github.com/sirupsen/logrus"
)

type ExecutionErrorCode int

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

	execution, err := taskexecution.PrepareTaskExecution(taskexecution.TaskExecutionRequestInfo{Code: solutionCode, UserId: userId, TaskId: taskId})
	if err != nil {
		switch err.Error() {
		case taskexecution.ErrUserHasRunningTasks.Error():
			api.TooEarlyErrorHandlerCustomMsg(w, err.Error())
		case taskexecution.ErrTaskExecutionStartErr.Error():
			api.InternalErrorHandlerGenericMsg(w, err)
		case taskexecution.ErrTempFileCreationErr.Error():
			execution.SetTaskExecutionStatusFailed()
			api.RequestErrorHandlerGenericMsg(w, err)
		case taskexecution.ErrNoTests.Error():
			execution.SetTaskExecutionStatusFailed()
			api.InternalErrorHandlerCustomMsg(w, err.Error())
		}
		return
	}

	reportedErrors, err := taskexecution.CheckSyntax(execution)
	if err != nil {
		execution.SetTaskExecutionStatusFailed()
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}
	if len(reportedErrors) != 0 {
		execution.SetTaskExecutionStatusFailed()
		sendTaskExecutionFailedResponse(w, EXEC_ERR_BAD_SYNTAX, reportedErrors)
		return
	}

	for _, test := range execution.Tests {
		err := execution.CreateInputFile(test)
		if err != nil {
			execution.SetTaskExecutionStatusFailed()
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Couldn't store temp files during testing!")
			handleTestExecutionInternalFail(w, err, execution)
			return
		}
	}

	var container = taskexecution.CreateTaskExecutionContainer(execution, r.Context())

	err = container.RunTests()
	if err != nil {
		execution.SetTaskExecutionStatusFailed()

		switch err.Error() {
		case taskexecution.ErrContainerTimeoutMark.Error():
			sendTaskExecutionFailedResponse(w, EXEC_ERR_TIMEOUT, nil)
		case taskexecution.ErrContainerForcefullyKilledMark.Error():
			sendTaskExecutionFailedResponse(w, EXEC_ERR_KILLED, nil)
		default:
			log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Task runner failed to run in Docker!")
			handleTestExecutionInternalFail(w, err, execution)
		}
		return
	}

	testDataMismatchReason, err := container.CheckOutputs()
	if err != nil {
		execution.SetTaskExecutionStatusFailed()
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Couldn't read output file!")
		handleTestExecutionInternalFail(w, err, execution)
		return
	}
	if testDataMismatchReason != nil {
		execution.SetTaskExecutionStatusFailed()
		sendTaskExecutionFailedResponse(w, EXEC_ERR_TEST_FAILED, testDataMismatchReason)
		return
	}

	artefactMismatchReason, err := container.CheckArtefacts()
	if err != nil {
		execution.SetTaskExecutionStatusFailed()
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "task_execution", "task_id": taskId}).Error("Artefact SHA256 comparison failed!")
		handleTestExecutionInternalFail(w, err, execution)
		return
	}
	if artefactMismatchReason != nil {
		execution.SetTaskExecutionStatusFailed()
		sendTaskExecutionFailedResponse(w, EXEC_ERR_ARTEFACT_CONTENT_MISMATCH, artefactMismatchReason)
		return
	}

	var solvedTask = database.GetSingleFullTask(taskId)
	numbericComplexity, err := strconv.Atoi(solvedTask.BasicTask.Complexity)
	if err != nil {
		numbericComplexity = 0
		log.WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Task complexity could not be converted to integer!")
	}

	score, err := lizard.CalculateScore(execution.File, numbericComplexity)
	if err != nil {
		execution.SetTaskExecutionStatusFailed()
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Could not calculate score after task execution.")
		api.InternalErrorHandlerCustomMsg(w, fmt.Sprintf("Could not calculate score: %v", err))
		return
	}

	execution.CleanupTempFolder()

	execution.SetTaskExecutionStatusSucceeded(solvedTask, *score)

	var successResponse = successfulTaskExecutionResponse{Success: true, Score: *score, Message: "Test data matches output!"}
	api.RespondOk(w, successResponse)
}

func sendTaskExecutionFailedResponse(w http.ResponseWriter, execErrCode ExecutionErrorCode, reasonFailed interface{}) {
	var errorResponse = taskExecutionFailedResponse{Success: false, ErrorCode: execErrCode.EnumIndex(), Message: execErrCode.String(), ReasonFailed: reasonFailed}
	api.RespondWithStatus(w, errorResponse, http.StatusUnprocessableEntity)
}

func handleTestExecutionInternalFail(w http.ResponseWriter, err error, data *taskexecution.TaskExecutionData) {
	api.InternalErrorHandlerGenericMsg(w, err)
	data.SetTaskExecutionStatusFailed()
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

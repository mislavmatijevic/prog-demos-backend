package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	log "github.com/sirupsen/logrus"
)

type helpStepResponse struct {
	Success  bool                  `json:"success"`
	HelpStep database.TaskHelpStep `json:"helpStep"`
}
type helpStepCountResponse struct {
	Success   bool  `json:"success"`
	HelpSteps int64 `json:"helpSteps"`
}
type availableHelpStepObject struct {
	Step              int       `json:"step"`
	DateMadeAvailable time.Time `json:"dateMadeAvailable"`
}
type availableResponse struct {
	Success   bool                      `json:"success"`
	HelpSteps []availableHelpStepObject `json:"availableHelpSteps"`
}

func getHelpStep(w http.ResponseWriter, r *http.Request) {
	var originalHelpStep = chi.URLParam(r, "helpStep")
	helpStep, err := strconv.Atoi(originalHelpStep)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	var originalTaskId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalTaskId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	if helpStep <= 0 || taskId <= 0 {
		api.RequestErrorHandlerCustomMsg(w, "Values given are not valid.")
		return
	}

	foundHelpStep := database.GetHelpStepByStepAndTaskId(helpStep, taskId)

	if foundHelpStep == nil {
		msg := fmt.Sprintf("Help step %s not found for task with id %s!", originalHelpStep, originalTaskId)
		api.NotFoundHandlerCustomMsg(w, msg)
		return
	}

	var res = helpStepResponse{
		Success:  true,
		HelpStep: *foundHelpStep,
	}

	api.RespondOk(w, res)
}

func getHelpStepCount(w http.ResponseWriter, r *http.Request) {
	var originalTaskId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalTaskId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	if taskId <= 0 {
		api.RequestErrorHandlerCustomMsg(w, "Task ID is not valid.")
		return
	}

	helpStepCount := database.GetHelpStepCountByTaskId(taskId)

	var res = helpStepCountResponse{
		Success:   true,
		HelpSteps: helpStepCount,
	}

	api.RespondOk(w, res)
}

func setHelpStepAvailable(w http.ResponseWriter, r *http.Request) {
	var originalHelpStep = chi.URLParam(r, "helpStep")
	helpStepId, err := strconv.Atoi(originalHelpStep)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	var originalTaskId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalTaskId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil || userId == 0 {
		api.InternalErrorHandlerCustomMsg(w, "Couldn't get user from JWT token.")
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't get user from JWT token during task execution.")
		return
	}

	err = database.MakeHelpStepAvailableForUser(userId, taskId, helpStepId)
	if err != nil {
		log.Info(err.Error())
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			api.RequestErrorHandlerCustomMsg(w, fmt.Sprintf("Help step %d was already made available for task %d.", helpStepId, taskId))
		} else if strings.Contains(err.Error(), "Help step not found") {
			api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Help step %d not found for task with ID %d.", helpStepId, taskId))
		} else {
			api.InternalErrorHandlerGenericMsg(w, err)
		}
		return
	}

	api.RespondOkWithDefaultBody(w)
}

func getAvailableHelpStepsPerTask(w http.ResponseWriter, r *http.Request) {
	var originalTaskId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalTaskId)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil || userId == 0 {
		api.InternalErrorHandlerCustomMsg(w, "Couldn't get user from JWT token.")
		log.WithError(err).WithFields(log.Fields{"priority": "medium", "context": "task_execution", "task_id": taskId}).Error("Couldn't get user from JWT token during task execution.")
		return
	}

	availableHelpSteps, err := database.GetAvailableHelpStepsForTaskByUser(userId, taskId)
	if err != nil || availableHelpSteps == nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}
	if availableHelpSteps == nil || len(availableHelpSteps) == 0 {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("No help steps available for task %d.", taskId))
		return
	}

	availableHelpStepsResponse := availableResponse{}
	for _, availableHelpStepEntity := range availableHelpSteps {
		availableHelpStepObject := availableHelpStepObject{}
		availableHelpStepObject.Step = availableHelpStepEntity.TaskHelpStep.Step
		availableHelpStepObject.DateMadeAvailable = availableHelpStepEntity.AvailableSince
		availableHelpStepsResponse.HelpSteps = append(availableHelpStepsResponse.HelpSteps, availableHelpStepObject)
	}
	api.RespondOk(w, availableHelpStepsResponse)
}

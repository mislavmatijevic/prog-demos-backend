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
type unlockedHelpStepObject struct {
	Step         int       `json:"step"`
	DateUnlocked time.Time `json:"dateUnlocked"`
}
type unlockedHelpStepsResponse struct {
	Success   bool                     `json:"success"`
	HelpSteps []unlockedHelpStepObject `json:"unlockedHelpSteps"`
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

func setHelpStepUnlocked(w http.ResponseWriter, r *http.Request) {
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

	err = database.SetUnlockedHelpStepsForTaskByUser(userId, taskId, helpStepId)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			api.RequestErrorHandlerCustomMsg(w, "This help step was already unlocked.")
		} else {
			api.InternalErrorHandlerGenericMsg(w, err)
		}
		return
	}

	api.RespondOkWithDefaultBody(w)
}

func getUnlockedHelpStepsPerTask(w http.ResponseWriter, r *http.Request) {
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

	unlockedHelpSteps, err := database.GetUnlockedHelpStepsForTaskByUser(userId, taskId)
	if err != nil || unlockedHelpSteps == nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}
	if unlockedHelpSteps == nil || len(unlockedHelpSteps) == 0 {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("No help steps unlocked for task %d.", taskId))
		return
	}

	unlockedHelpStepsResponse := unlockedHelpStepsResponse{}
	for _, unlockedHelpStepEntity := range unlockedHelpSteps {
		unlockedHelpStepObject := unlockedHelpStepObject{}
		unlockedHelpStepObject.Step = unlockedHelpStepEntity.TaskHelpStep.Step
		unlockedHelpStepObject.DateUnlocked = unlockedHelpStepEntity.DateUnlocked
		unlockedHelpStepsResponse.HelpSteps = append(unlockedHelpStepsResponse.HelpSteps, unlockedHelpStepObject)
	}
	api.RespondOk(w, unlockedHelpStepsResponse)
}

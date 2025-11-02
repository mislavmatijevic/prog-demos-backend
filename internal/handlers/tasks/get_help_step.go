package tasks

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type helpStepResponse struct {
	Success  bool                  `json:"success"`
	HelpStep database.TaskHelpStep `json:"helpStep"`
}
type helpStepCountResponse struct {
	Success   bool  `json:"success"`
	HelpSteps int64 `json:"helpSteps"`
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

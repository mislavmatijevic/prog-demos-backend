package tasks

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	log "github.com/sirupsen/logrus"
)

type taskResponse struct {
	Success bool              `json:"success"`
	Task    database.FullTask `json:"task"`
}

func getSingleTask(w http.ResponseWriter, r *http.Request) {
	var originalParamId = chi.URLParam(r, "taskId")
	taskId, err := strconv.Atoi(originalParamId)

	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	task := database.GetSingleFullTasks(taskId)

	if err := authentication.ValidateJwtTokenFromRequest(r); err == nil {
		fillTaskWithPersonalizedInfo(r, task)
	}

	if task == nil {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Task with id %s not found!", originalParamId))
		return
	}

	var res = taskResponse{
		Success: true,
		Task:    *task,
	}

	api.RespondOk(w, res)
}

func fillTaskWithPersonalizedInfo(r *http.Request, task *database.FullTask) {
	var userId, err = authentication.GetUserIdFromRequest(r)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "tasks"}).Error("Token validated for filling a task with user info, but couldn't extract user id!")
		return
	}

	fillInfoOnCompletedTask(userId, task)
}

func fillInfoOnCompletedTask(userId int, task *database.FullTask) {
	task.BasicInfo.BestExecutionForUser = database.GetBestScoreExecutionForUserAndTask(userId, task.BasicInfo.ID)
	if task.BasicInfo.BestExecutionForUser != nil {
		task.BasicInfo.BestExecutionForUser.SubmittedCode = ""
	}
}

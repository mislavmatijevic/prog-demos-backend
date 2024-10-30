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
	var originalParamId = chi.URLParam(r, "taskIdentifier")
	taskIdentifier, err := strconv.Atoi(originalParamId)

	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	task := database.GetSingleFullTaskByIdentifier(taskIdentifier)

	if task == nil {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Task with identifier %s not found!", originalParamId))
		return
	}

	if err := authentication.CheckJwtTokenSignature(r); err == nil {
		fillTaskWithPersonalizedInfo(r, task)
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
	task.BasicTask.BestExecutionForUser = database.GetBestScoreExecutionForUserAndTask(userId, task.BasicTask.ID)
}

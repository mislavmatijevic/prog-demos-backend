package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type taskResponse = struct {
	Success bool              `json:"success"`
	Task    database.FullTask `json:"task"`
}

func getSingleTask(w http.ResponseWriter, r *http.Request) {
	var originalParamId = chi.URLParam(r, "taskId")
	TaskId, err := strconv.Atoi(originalParamId)

	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	Task := database.GetSingleFullTasks(TaskId)

	if Task == nil {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Task with id %s not found!", originalParamId))
		return
	}

	res := taskResponse{
		Success: true,
		Task:    *Task,
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

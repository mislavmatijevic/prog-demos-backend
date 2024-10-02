package tasks

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type tasksResponse struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getAllTasksPerTopics(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllTasksPerTopic()

	var res = tasksResponse{
		Success: true,
		Topics:  topics,
	}

	api.RespondOk(w, res)
}

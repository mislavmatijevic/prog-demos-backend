package topics

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type topicsResponse struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getAllTopics(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllTopicsWithSubtopics()

	var res = topicsResponse{
		Success: true,
		Topics:  topics,
	}

	api.RespondOk(w, res)
}

package topics

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

type topicsResponse = struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getAllTopics(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllTopicsWithSubtopics()

	res := topicsResponse{
		Success: true,
		Topics:  topics,
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

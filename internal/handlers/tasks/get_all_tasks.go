package tasks

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

type tasksResponse = struct {
	Topics []database.Topic `json:"topics"`
}

func GetAllTasksPerTopics(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllTasksPerTopic()

	res := tasksResponse{Topics: topics}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

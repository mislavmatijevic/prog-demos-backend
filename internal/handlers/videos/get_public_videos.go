package videos

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

type videosResponse = struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getPublicVideos(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllVideosPerTopics()

	res := videosResponse{
		Success: true,
		Topics:  topics,
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

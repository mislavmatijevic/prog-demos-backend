package videos

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type videosResponse struct {
	Success bool             `json:"success"`
	Topics  []database.Topic `json:"topics"`
}

func getPublicVideos(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllVideosPerTopics()

	var res = videosResponse{
		Success: true,
		Topics:  topics,
	}

	api.RespondOk(w, res)
}

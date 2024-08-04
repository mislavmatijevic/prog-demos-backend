package videos

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

type videosResponse = struct {
	Topics []database.Topic `json:"topics"`
}

func GetPublicVideos(w http.ResponseWriter, r *http.Request) {
	topics := database.GetAllVideosPerTopics()

	res := videosResponse{Topics: topics}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

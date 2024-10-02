package videos

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type videoResponse struct {
	Success bool           `json:"success"`
	Video   database.Video `json:"video"`
}

func getSingleVideo(w http.ResponseWriter, r *http.Request) {
	var originalParamId = chi.URLParam(r, "videoId")
	videoId, err := strconv.Atoi(originalParamId)

	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	video := database.GetSingleVideo(videoId)

	if video == nil {
		api.NotFoundHandlerCustomMsg(w, fmt.Sprintf("Video with id %s not found!", originalParamId))
		return
	}

	var res = videoResponse{
		Success: true,
		Video:   *video,
	}

	api.RespondOk(w, res)
}

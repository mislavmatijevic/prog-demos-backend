package misc

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type newsResponse struct {
	Success bool            `json:"success"`
	News    []database.News `json:"news"`
}

func getNews(w http.ResponseWriter, r *http.Request) {
	news := database.GetNews()

	var res = newsResponse{
		Success: true,
		News:    news,
	}

	api.RespondOk(w, res)
}

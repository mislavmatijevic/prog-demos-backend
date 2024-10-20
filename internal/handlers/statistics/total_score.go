package statistics

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type totalScoreBody struct {
	TotalScore int `json:"totalScore"`
}

func GetTotalScore(w http.ResponseWriter, r *http.Request) {
	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	user, err := database.GetUserById(userId)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	api.RespondOk(w, &totalScoreBody{TotalScore: user.TotalScore})
}

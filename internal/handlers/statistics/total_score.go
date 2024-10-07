package statistics

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

type totalScoreBody struct {
	TotalScore float32 `json:"totalScore"`
}

func GetTotalScore(w http.ResponseWriter, r *http.Request) {
	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	totalScore, err := database.GetTotalScore(userId)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	totalScore = utils.RoundNumberDownToTwoDecimals(totalScore)

	api.RespondOk(w, &totalScoreBody{TotalScore: totalScore})
}

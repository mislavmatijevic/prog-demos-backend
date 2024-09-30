package statistics

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

func GetTotalCountOfSolutionAttempts(w http.ResponseWriter, r *http.Request) {
	userId, err := authentication.GetUserIdFromToken(r)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	solutionAttemptsDto, err := database.GetAllSolutionAttempts(userId)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(solutionAttemptsDto)
}

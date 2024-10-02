package auth

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type logoutBody struct {
	RefreshToken string `json:"refreshToken"`
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	var logoutBody logoutBody
	err := json.NewDecoder(r.Body).Decode(&logoutBody)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	wasSuccessful := authentication.RemoveRefreshToken(logoutBody.RefreshToken)
	if wasSuccessful {
		api.RespondOk(w, struct{ message string }{message: "Logged out."})
	} else {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}
}

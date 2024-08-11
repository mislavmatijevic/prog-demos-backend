package auth

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type LogoutBody struct {
	RefreshToken string `json:"refreshToken"`
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	var logoutBody LogoutBody
	err := json.NewDecoder(r.Body).Decode(&logoutBody)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	wasSuccessful := authentication.RemoveRefreshToken(logoutBody.RefreshToken)
	if wasSuccessful {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("content-type", "application/json")
	} else {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}
}

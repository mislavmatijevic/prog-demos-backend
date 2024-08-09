package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type ActivationBody struct {
	ActivationToken string `json:"activation_token"`
}

type ActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func ActivateUser(w http.ResponseWriter, r *http.Request) {
	var loginBody ActivationBody
	if err := json.NewDecoder(r.Body).Decode(&loginBody); err != nil && strings.Trim(loginBody.ActivationToken, " ") != "" {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	var res ActivationResponse

	user, err := database.SetUserActivated(loginBody.ActivationToken)
	if err != nil {
		res.Success = false
		res.Message = fmt.Sprintf("Failed to activate the user: %s", err)
		w.WriteHeader(http.StatusForbidden)
	} else {
		res.Success = true
		res.Message = fmt.Sprintf("User %s activated", user.Username)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

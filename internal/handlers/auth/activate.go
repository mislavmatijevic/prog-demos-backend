package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

type activationBody struct {
	ActivationToken string `json:"activationToken"`
}

type activationResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Username string `json:"username,omitempty"`
}

func activateUser(w http.ResponseWriter, r *http.Request) {
	var activationBody activationBody
	if err := json.NewDecoder(r.Body).Decode(&activationBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	hasValue, trimmedToken := utils.GetTrimmedStringWithValue(activationBody.ActivationToken)
	if !hasValue {
		api.RequestErrorHandlerCustomMsg(w, "Activation token not procured.")
		return
	}

	user, err := database.SetUserActivated(trimmedToken)

	var res activationResponse
	var status int

	if err != nil {
		res = activationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to activate the user: %s", err),
		}
		status = http.StatusForbidden
	} else {
		res = activationResponse{
			Success:  true,
			Message:  fmt.Sprintf("User %s activated", user.Username),
			Username: user.Username,
		}
		status = http.StatusOK
	}

	api.RespondWithStatus(w, res, status)
}

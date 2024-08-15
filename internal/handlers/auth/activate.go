package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

type ActivationBody struct {
	ActivationToken string `json:"activationToken"`
}

type ActivationResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Username string `json:"username,omitempty"`
}

func ActivateUser(w http.ResponseWriter, r *http.Request) {
	var activationBody ActivationBody
	if err := json.NewDecoder(r.Body).Decode(&activationBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	hasValue, trimmedToken := utils.GetTrimmedStringWithValue(activationBody.ActivationToken)
	if !hasValue {
		api.RequestErrorHandlerCustomMsg(w, "Activation token not procured.")
		return
	}

	var res ActivationResponse

	user, err := database.SetUserActivated(trimmedToken)
	if err != nil {
		res.Success = false
		res.Message = fmt.Sprintf("Failed to activate the user: %s", err)
		w.WriteHeader(http.StatusForbidden)
	} else {
		res.Success = true
		res.Message = fmt.Sprintf("User %s activated", user.Username)
		res.Username = user.Username
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

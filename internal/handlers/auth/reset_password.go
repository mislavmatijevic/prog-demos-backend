package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security"
)

type resetPasswordBody struct {
	NewPassword    string `json:"newPassword"`
	ResetToken     string `json:"resetToken"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type passwordResetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func resetPassword(w http.ResponseWriter, r *http.Request) {
	var resetPasswordBody resetPasswordBody
	err := json.NewDecoder(r.Body).Decode(&resetPasswordBody)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err = security.VerifyRecaptcha("reset_password", resetPasswordBody.RecaptchaToken, r.RemoteAddr)
	if err != nil {
		handleRecaptchaError(w, err)
		return
	}

	isPasswordValid := len(resetPasswordBody.NewPassword) >= 8 && len(resetPasswordBody.NewPassword) < 72
	if !isPasswordValid {
		api.RequestErrorHandlerCustomMsg(w, "New password is invalid!")
		return
	}

	isTokenSet, trimmedToken := utils.GetTrimmedStringWithValue(resetPasswordBody.ResetToken)
	if !isTokenSet || len(trimmedToken) != 128 {
		api.RequestErrorHandlerCustomMsg(w, "Reset token not set or invalid!")
		return
	}

	hashedPassword, err := utils.CreateSecureHash(resetPasswordBody.NewPassword)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	user, err := database.ChangeUserPassword(trimmedToken, hashedPassword)

	var res passwordResetResponse
	var status int

	if err != nil {
		res = passwordResetResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to fulfill password reset request: %s", err),
		}
		status = http.StatusForbidden
	} else {
		res = passwordResetResponse{
			Success: true,
			Message: fmt.Sprintf("Changed password for user: %s", user.Username),
		}
		status = http.StatusOK
	}

	api.RespondWithStatus(w, res, status)
}

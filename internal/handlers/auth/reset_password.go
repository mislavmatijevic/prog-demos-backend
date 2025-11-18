package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/utils_errors"
	log "github.com/sirupsen/logrus"
)

type checkPasswordResetTokenBody struct {
	ResetToken   string `json:"resetToken"`
	CaptchaToken string `json:"captchaToken"`
}

type resetPasswordBody struct {
	NewPassword  string `json:"newPassword"`
	ResetToken   string `json:"resetToken"`
	CaptchaToken string `json:"captchaToken"`
}

type passwordResetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func checkPasswordResetToken(w http.ResponseWriter, r *http.Request) {
	var checkPasswordResetTokenBody checkPasswordResetTokenBody
	err := json.NewDecoder(r.Body).Decode(&checkPasswordResetTokenBody)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err = captcha.Verify("password-reset", checkPasswordResetTokenBody.CaptchaToken, r.RemoteAddr)
	if err != nil {
		utils_errors.HandleCaptchaError(w, err)
		return
	}

	user, errCode := checkIfTokenValid(checkPasswordResetTokenBody.ResetToken)
	if errCode != utils_errors.NO_ERROR {
		utils_errors.RespondForErrorCode(w, errCode)
		return
	}

	res := passwordResetResponse{Success: true, Message: fmt.Sprintf("Found token for user: %s", user.Username)}
	api.RespondOk(w, res)
}

func resetPassword(w http.ResponseWriter, r *http.Request) {
	var resetPasswordBody resetPasswordBody
	err := json.NewDecoder(r.Body).Decode(&resetPasswordBody)
	if err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	user, errCode := checkIfTokenValid(resetPasswordBody.ResetToken)
	if errCode != utils_errors.NO_ERROR {
		utils_errors.RespondForErrorCode(w, errCode)
		return
	}

	err = captcha.Verify("password-reset", resetPasswordBody.CaptchaToken, r.RemoteAddr)
	if err != nil {
		utils_errors.HandleCaptchaError(w, err)
		return
	}

	isPasswordValid := len(resetPasswordBody.NewPassword) >= 8 && len(resetPasswordBody.NewPassword) < 72
	if !isPasswordValid {
		api.RequestErrorHandlerCustomMsg(w, "New password is invalid!")
		return
	}

	hashedPassword, err := utils.CreateSecureHash(resetPasswordBody.NewPassword)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	err = database.ChangeUserPassword(user, hashedPassword)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "reset_password"}).Error("Failed to change user password")
		api.InternalErrorHandlerCustomMsg(w, "Unknown captcha error.")
		return
	}

	var res = passwordResetResponse{
		Success: true,
		Message: fmt.Sprintf("Changed password for user: %s", user.Username),
	}
	api.RespondWithStatus(w, res, http.StatusOK)
}

func checkIfTokenValid(resetToken string) (*database.User, utils_errors.ErrorCode) {
	isTokenSet, trimmedToken := utils.GetTrimmedStringWithValue(resetToken)
	if !isTokenSet || len(trimmedToken) != 128 {
		return nil, utils_errors.ERR_TOKEN_NOT_VALID
	}

	user, err := database.GetUserByPasswordResetToken(trimmedToken)
	if err != nil {
		return nil, utils_errors.ERR_TOKEN_NOT_FOUND
	}

	if !user.PasswordResetExpiry.Valid || time.Now().After(user.PasswordResetExpiry.Time) {
		database.RemovePasswordReset(user)
		return nil, utils_errors.ERR_TOKEN_EXPIRED
	}

	return user, utils_errors.NO_ERROR
}

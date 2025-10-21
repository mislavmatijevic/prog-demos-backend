package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
)

type requestPasswordResetBody struct {
	Email        string `json:"email,omitempty"`
	CaptchaToken string `json:"captchaToken"`
}

func requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var requestBody requestPasswordResetBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err := captcha.Verify("request-password-reset", requestBody.CaptchaToken, r.RemoteAddr)
	if err != nil {
		handleCaptchaError(w, err)
		return
	}

	if !utils.IsEmailValid(requestBody.Email) {
		api.NotFoundHandlerCustomMsg(w, "Valid email not procured.")
		return
	}

	user := getActivatedUserByEmail(requestBody.Email)
	if user == nil {
		api.NotFoundHandlerCustomMsg(w, "Valid email not procured.")
		return
	}
	if userHasPasswordResetInProgress(user) {
		api.TooEarlyErrorHandlerCustomMsg(w, "Another password request in progress, try again later.")
		return
	}

	setPasswordResetToken(user)
	err = mailing.SendPasswordRequestMail(*user)
	if err != nil {
		api.InternalErrorHandlerCustomMsg(w, "Could not send email for password reset!")
		return
	}

	api.RespondWithStatus(w, struct{ message string }{message: "Mail sent."}, http.StatusAccepted)
}

func getActivatedUserByEmail(email string) *database.User {
	foundUser := database.GetUserByEmail(email)
	if foundUser != nil && foundUser.IsActivated {
		return foundUser
	} else {
		return nil
	}
}

func userHasPasswordResetInProgress(user *database.User) bool {
	var hasPasswordResetInProgress = false

	if user.PasswordResetExpiry.Valid {
		hasPasswordResetInProgress = user.PasswordResetExpiry.Time.After(time.Now())
	}

	return hasPasswordResetInProgress
}

func setPasswordResetToken(user *database.User) {
	user.PasswordResetToken = database.WrappedNullString{NullString: sql.NullString{String: utils.RandomString(128), Valid: true}}
	user.PasswordResetExpiry = database.WrappedNullTime{NullTime: sql.NullTime{Time: time.Now().Add(10 * time.Minute), Valid: true}}
	database.SaveUser(*user)
}

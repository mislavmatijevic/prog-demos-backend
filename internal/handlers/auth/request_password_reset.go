package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

type requestPasswordResetBody struct {
	Email string `json:"email,omitempty"`
}

func requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var requestBody requestPasswordResetBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}
	if !utils.IsEmailValid(requestBody.Email) {
		api.RequestErrorHandlerCustomMsg(w, "Valid email not procured.")
		return
	}

	user := getActivatedUserByEmail(requestBody.Email)
	if user == nil {
		api.RequestErrorHandlerCustomMsg(w, fmt.Sprintf("User with email '%s' not found.", requestBody.Email))
		return
	}
	if userHasPasswordResetInProgress(user) {
		api.TooEarlyErrorHandlerCustomMsg(w, "Another password request in progress, try again later.")
		return
	}

	setPasswordResetToken(user)
	err := mailing.SendPasswordRequestMail(*user)
	if err != nil {
		api.InternalErrorHandlerCustomMsg(w, "Could not send email for password reset!")
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("content-type", "application/json")
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

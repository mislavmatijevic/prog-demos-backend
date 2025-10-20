package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/mailing"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	log "github.com/sirupsen/logrus"
)

type userRegisterBody struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captchaToken"`
}

type successResponse struct {
	Success bool `json:"success"`
	NewId   int  `json:"newId"`
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var userReqBody userRegisterBody
	if err := json.NewDecoder(r.Body).Decode(&userReqBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err := captcha.Verify("register", userReqBody.CaptchaToken, r.RemoteAddr)
	if err != nil {
		handleCaptchaError(w, err)
		return
	}

	username, email := strings.TrimSpace(userReqBody.Username), strings.TrimSpace(userReqBody.Email)
	var infoIsValid bool = checkIfUserInfoValid(username, email, userReqBody.Password)
	if !infoIsValid {
		respondForErrorCode(w, EXEC_ERR_INFO_INVALID)
		return
	}

	hashPassword, err := utils.CreateSecureHash(userReqBody.Password)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "registration", "username": username, "email": email}).Error("Couldn't create hashed password.")
		return
	}

	var user database.User = createUser(username, email, hashPassword)

	newUser, err := database.RegisterNewUser(user)
	if err != nil {
		if err.Error() == "user already exists" {
			respondForErrorCode(w, EXEC_ERR_USERNAME_TAKEN)
		} else {
			api.RequestErrorHandlerCustomMsg(w, err.Error())
		}
		return
	}

	err = mailing.SendRegistrationMail(*newUser)
	if err != nil {
		database.DeleteUser(newUser)
		api.InternalErrorHandlerCustomMsg(w, "Failed to send registration mail, rollbacked registration.")
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "registration", "mail_address": newUser.Email, "username": newUser.Username}).Error("Couldn't send mail to user.")
		return
	}

	var res = successResponse{
		Success: true,
		NewId:   newUser.ID,
	}

	api.RespondWithStatus(w, res, http.StatusCreated)
}

func checkIfUserInfoValid(username, email, password string) bool {
	var usernameAtLeast2Characters = len(username) >= 2
	var usernameAtMost20Characters = len(username) <= 20
	var emailAtLeast4Characters = len(email) >= 4
	var isEmailValid = utils.IsEmailValid(email)
	var usernameDoesNotContainAt = !strings.Contains(username, "@")
	var passwordAtLeast8Chars = len(password) >= 8
	var passwordNotLongerThan72Chars = len(password) < 72

	return usernameAtLeast2Characters && usernameAtMost20Characters &&
		emailAtLeast4Characters && isEmailValid &&
		usernameDoesNotContainAt &&
		passwordAtLeast8Chars && passwordNotLongerThan72Chars
}

func createUser(username, email, hashPassword string) database.User {
	return database.User{
		Username:        username,
		Email:           email,
		Password:        hashPassword,
		IsActivated:     false,
		ActivationToken: database.WrappedNullString{NullString: sql.NullString{String: utils.RandomString(128), Valid: true}},
		DateRegistered:  time.Now(),
		UserType:        "basic",
	}
}

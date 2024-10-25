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
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security"
	log "github.com/sirupsen/logrus"
)

type registrationErrorCode int

const (
	EXEC_ERR_INFO_INVALID registrationErrorCode = iota + 1
	EXEC_ERR_USERNAME_TAKEN
	EXEC_ERR_RECAPTCHA_REQUIRES_CHALLENGE
)

func (execErrCode registrationErrorCode) String() string {
	return [...]string{
		"Given information is not valid for registration.",
		"Username or email already taken.",
		"Login did not score well at ReCaptcha, challenge user.",
	}[execErrCode-1]
}

func (execErrCode registrationErrorCode) EnumIndex() int {
	return int(execErrCode)
}

type userRegisterBody struct {
	Email          string `json:"email"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type successResponse struct {
	Success bool `json:"success"`
	NewId   int  `json:"newId"`
}

type errorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var userReqBody userRegisterBody
	if err := json.NewDecoder(r.Body).Decode(&userReqBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	username, email, recaptchaToken := strings.Trim(userReqBody.Username, " "), strings.Trim(userReqBody.Email, " "), strings.Trim(userReqBody.RecaptchaToken, " ")
	var infoIsValid bool = checkIfUserInfoValid(username, email, userReqBody.Password, recaptchaToken)
	if !infoIsValid {
		respondForErrorCode(w, EXEC_ERR_INFO_INVALID)
		return
	}

	err := security.VerifyRecaptcha("register", recaptchaToken, r.RemoteAddr)
	if err != nil {
		handleRecaptchaError(w, err)
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

func handleRecaptchaError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case security.ErrFailedToProcess.Error():
		api.InternalErrorHandlerCustomMsg(w, "Recaptcha is not available.")
	case security.ErrMustChallenge.Error():
		respondForErrorCode(w, EXEC_ERR_RECAPTCHA_REQUIRES_CHALLENGE)
	case security.ErrInvalid.Error():
		fallthrough
	default:
		api.RequestErrorHandlerCustomMsg(w, "Failed recaptcha.")
	}
}

func respondForErrorCode(w http.ResponseWriter, errorCode registrationErrorCode) {
	var res = errorResponse{
		Success:   false,
		Message:   errorCode.String(),
		ErrorCode: errorCode.EnumIndex(),
	}

	api.RespondWithStatus(w, res, http.StatusBadRequest)
}

func checkIfUserInfoValid(username, email, password string, recaptchaToken string) bool {
	var usernameAtLeast2Characters = len(username) >= 2
	var usernameAtMost15Characters = len(username) <= 20
	var emailAtLeast4Characters = len(email) >= 4
	var isEmailValid = utils.IsEmailValid(email)
	var usernameDoesNotContainAt = !strings.Contains(username, "@")
	var passwordAtLeast8Chars = len(password) >= 8
	var passwordNotLongerThan72Chars = len(password) < 72
	var recaptchaTokenNotEmpty = len(recaptchaToken) > 0

	return usernameAtLeast2Characters && usernameAtMost15Characters &&
		emailAtLeast4Characters && isEmailValid &&
		usernameDoesNotContainAt &&
		passwordAtLeast8Chars && passwordNotLongerThan72Chars &&
		recaptchaTokenNotEmpty
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

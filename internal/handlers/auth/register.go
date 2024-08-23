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
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type registrationErrorCode int

const (
	EXEC_ERR_INFO_INVALID registrationErrorCode = iota + 1
	EXEC_ERR_USERNAME_TAKEN
)

func (execErrCode registrationErrorCode) String() string {
	return [...]string{
		"Given information is not valid for registration.",
		"Username or email already taken.",
	}[execErrCode-1]
}

func (execErrCode registrationErrorCode) EnumIndex() int {
	return int(execErrCode)
}

type userRegisterBody struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type successResponse struct {
	Success bool `json:"success"`
	NewId   int  `json:"newId"`
}

type errorResponse = struct {
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

	username, email := strings.Trim(userReqBody.Username, " "), strings.Trim(userReqBody.Email, " ")
	var infoIsValid bool = checkIfUserInfoValid(username, email, userReqBody.Password)
	if !infoIsValid {
		respondForErrorCode(w, EXEC_ERR_INFO_INVALID)
		return
	}

	hashPassword, err := getHashPassword(userReqBody.Password)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
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
		log.Errorf("Error sending mail: %v", err)
		database.DeleteUser(newUser)
		api.InternalErrorHandlerCustomMsg(w, "Failed to send registration mail, rollbacked registration.")
		return
	}

	res := successResponse{
		Success: true,
		NewId:   newUser.ID,
	}

	w.WriteHeader(http.StatusCreated)
	writeRequest(w, res)
}

func respondForErrorCode(w http.ResponseWriter, errorCode registrationErrorCode) {
	var res = errorResponse{
		Success:   false,
		Message:   errorCode.String(),
		ErrorCode: errorCode.EnumIndex(),
	}
	w.WriteHeader(http.StatusBadRequest)
	writeRequest(w, res)
}

func getHashPassword(password string) (string, error) {
	bytePassword := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkIfUserInfoValid(username, email, password string) bool {
	var usernameAtLeast2Characters = len(username) >= 2
	var emailAtLeast4Characters = len(email) >= 4
	var isEmailValid = utils.IsEmailValid(email)
	var usernameDoesNotContainAt = !strings.Contains(username, "@")
	var passwordAtLeast8Chars = len(password) >= 8
	return usernameAtLeast2Characters && emailAtLeast4Characters && isEmailValid && usernameDoesNotContainAt && passwordAtLeast8Chars
}

func createUser(username, email, hashPassword string) database.User {
	return database.User{
		Username:        username,
		Email:           email,
		Password:        hashPassword,
		IsActivated:     false,
		ActivationToken: utils.RandomString(128),
		DateRegistered:  time.Now(),
		UserType:        "basic",
	}
}

func writeRequest(w http.ResponseWriter, res any) {
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

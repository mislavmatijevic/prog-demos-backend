package auth

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserRegisterBody struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	NewId int `json:"new_id"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userReqBody UserRegisterBody
	if err := json.NewDecoder(r.Body).Decode(&userReqBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	username, email := userReqBody.Username, userReqBody.Email
	var infoIsValid bool = checkIfUserInfoValid(username, email, userReqBody.Password)
	if !infoIsValid {
		api.RequestErrorHandlerCustomMsg(w, "User information is not valid for registration.")
		return
	}

	hashPassword, err := getHashPassword(userReqBody.Password)
	if err != nil {
		api.InternalErrorHandler(w, err)
		return
	}

	var user database.User = createUser(username, email, hashPassword)

	newUserId, err := database.RegisterNewUser(user)
	if err != nil {
		api.RequestErrorHandlerCustomMsg(w, err.Error())
		return
	}

	res := Response{
		NewId: newUserId,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
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
	var nonEmptyStrings = strings.Trim(username, " ") != "" && strings.Trim(email, " ") != ""
	var isEmailValid = isEmailValid(email)
	var usernameDoesNotContainAt = !strings.Contains(username, "@")
	var passwordAtLeast8Chars = len(password) >= 8
	return nonEmptyStrings && isEmailValid && usernameDoesNotContainAt && passwordAtLeast8Chars
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
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

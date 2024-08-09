package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginBody UserLoginBody
	err := json.NewDecoder(r.Body).Decode(&loginBody)
	if err != nil || !isValidLoginBody(loginBody) {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	loginAllowed, user := getValidUser(loginBody)
	if !loginAllowed {
		api.RequestErrorHandlerCustomMsg(w, "Couldn't log in.")
		return
	}

	res, err := authentication.GenerateNewTokenPair(user)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func isValidLoginBody(loginBody UserLoginBody) bool {
	return strings.Trim(loginBody.Identifier, " ") != "" && strings.Trim(loginBody.Password, " ") != ""
}

func getValidUser(loginBody UserLoginBody) (isUserOk bool, foundUser *database.User) {
	if strings.Contains(loginBody.Identifier, "@") {
		foundUser = database.GetUserByEmail(loginBody.Identifier)
	} else {
		foundUser = database.GetUserByUsername(loginBody.Identifier)
	}

	isUserOk = foundUser != nil && isPasswordCorrect(foundUser.Password, loginBody.Password) && foundUser.IsActivated
	return isUserOk, foundUser
}

func isPasswordCorrect(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

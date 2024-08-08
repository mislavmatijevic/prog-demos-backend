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
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginSuccessResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

	accessTokenString, err := authentication.GenerateAccessToken(user)
	if err != nil {
		api.InternalErrorHandler(w, err)
		return
	}

	refreshTokenString, err := authentication.GenerateRefreshToken(user)
	if err != nil {
		api.InternalErrorHandler(w, err)
		return
	}

	var res = LoginSuccessResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func isValidLoginBody(loginBody UserLoginBody) bool {
	return strings.Trim(loginBody.Email, " ") != "" && strings.Trim(loginBody.Password, " ") != ""
}

func getValidUser(loginBody UserLoginBody) (isUserOk bool, foundUser *database.User) {
	foundUser = database.GetUserByEmail(loginBody.Email)
	isUserOk = foundUser != nil && isPasswordCorrect(foundUser.Password, loginBody.Password) && foundUser.IsActivated
	return isUserOk, foundUser
}

func isPasswordCorrect(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/captcha"
	"golang.org/x/crypto/bcrypt"
)

type loginBody struct {
	Identifier   string `json:"identifier"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captchaToken"`
}

type loginResponse struct {
	Success  bool                         `json:"success"`
	UserInfo database.User                `json:"user"`
	Tokens   authentication.AuthTokenPair `json:"tokens"`
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var loginBody loginBody
	if err := json.NewDecoder(r.Body).Decode(&loginBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err := captcha.Verify("login", loginBody.CaptchaToken, r.RemoteAddr)
	if err != nil {
		handleCaptchaError(w, err)
		return
	}

	isIdentifierSet, trimmedIdentifier := utils.GetTrimmedStringWithValue(loginBody.Identifier)
	if !isIdentifierSet || len(loginBody.Password) == 0 {
		api.RequestErrorHandlerCustomMsg(w, "Login attributes not correctly set!")
		return
	}

	loginAllowed, user := getValidUser(trimmedIdentifier, loginBody.Password)
	if !loginAllowed {
		api.RequestErrorHandlerCustomMsg(w, "Couldn't log in.")
		return
	}

	newTokenPair, err := authentication.GenerateNewTokenPair(user)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	var res = loginResponse{
		Success:  true,
		UserInfo: *user,
		Tokens:   *newTokenPair,
	}

	api.RespondOk(w, res)
}

func getValidUser(identifier string, password string) (isUserOk bool, foundUser *database.User) {
	if strings.Contains(identifier, "@") {
		foundUser = database.GetUserByEmail(identifier)
	} else {
		foundUser = database.GetUserByUsername(identifier)
	}

	isUserOk = foundUser != nil && isPasswordCorrect(foundUser.Password, password) && foundUser.IsActivated
	return isUserOk, foundUser
}

func isPasswordCorrect(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

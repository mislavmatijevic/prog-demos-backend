package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/sirupsen/logrus"
)

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func RefreshAccess(w http.ResponseWriter, r *http.Request) {
	var refreshBody RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	var refreshTokenValue string = refreshBody.RefreshToken

	isOk := authentication.ValidateRefreshTokenFormat(refreshTokenValue)
	if !isOk {
		api.AuthorizationExpiredGenericMsg(w)
		return
	}

	refreshToken, isValid := findValidRefreshToken(refreshTokenValue)
	if !isValid {
		api.AuthorizationExpiredGenericMsg(w)
		return
	}

	newTokenPair, err := authentication.GenerateNewTokenPair(refreshToken.Owner)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	var res authentication.AuthTokenPair = *newTokenPair
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func findValidRefreshToken(refreshTokenValue string) (user *database.RefreshToken, isValid bool) {
	refreshToken := database.GetRefreshTokenWithUser(refreshTokenValue)
	logrus.Info(refreshToken)

	if refreshToken == nil || refreshToken.Owner == nil {
		return nil, false
	}

	if time.Now().After(refreshToken.Expiration) {
		return nil, false
	}

	return refreshToken, true
}

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
)

type refreshRequest struct {
	AccessTokenValue  string `json:"accessToken"`
	RefreshTokenValue string `json:"refreshToken"`
}

type refreshResponse struct {
	Success bool `json:"success"`
	*authentication.AuthTokenPair
}

func refreshAccess(w http.ResponseWriter, r *http.Request) {
	var refreshBody refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	previousAccessTokenValue, refreshTokenValue, isValid := validateRequestFormat(refreshBody)
	if !isValid {
		api.RequestErrorHandlerCustomMsg(w, "Valid token pair could not be extracted from your request.")
		return
	}

	refreshToken, isExpired := findValidRefreshToken(refreshTokenValue)
	if refreshToken == nil {
		api.RequestErrorHandlerCustomMsg(w, "Refresh token does not exist.")
		return
	}
	if isExpired {
		api.AuthorizationExpiredGenericMsg(w)
		return
	}

	err := authentication.ValidateTokenPair(previousAccessTokenValue, refreshToken)
	if err != nil {
		api.AuthorizationInvalidCustomMsg(w, fmt.Sprintf("Couldn't generate new refresh token: %v", err))
		return
	}

	newTokenPair, err := authentication.GenerateNewTokenPair(refreshToken.Owner)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	var res refreshResponse = refreshResponse{
		Success:       true,
		AuthTokenPair: newTokenPair,
	}
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func validateRequestFormat(refreshBody refreshRequest) (previousAccessTokenValue string, refreshTokenValue string, isValid bool) {
	isValid = authentication.ValidateRefreshTokenFormat(refreshBody.RefreshTokenValue) && len(refreshBody.AccessTokenValue) > 20
	if isValid {
		previousAccessTokenValue = refreshBody.AccessTokenValue
		refreshTokenValue = refreshBody.RefreshTokenValue
	}
	return
}

func findValidRefreshToken(refreshTokenValue string) (refreshToken *database.RefreshToken, isExpired bool) {
	refreshToken = database.GetRefreshTokenWithUser(refreshTokenValue)

	if refreshToken == nil || refreshToken.Owner == nil {
		return nil, false
	}

	if time.Now().After(refreshToken.Expiration) {
		return nil, true
	}

	return refreshToken, false
}

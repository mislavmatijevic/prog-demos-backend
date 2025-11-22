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
	Success   bool                          `json:"success"`
	NewTokens *authentication.AuthTokenPair `json:"tokens"`
}

func refreshAccess(w http.ResponseWriter, r *http.Request) {
	var refreshBody refreshRequest
	var err error

	if err = json.NewDecoder(r.Body).Decode(&refreshBody); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	previousAccessTokenValue, refreshTokenValue, isValid := validateRequestFormat(refreshBody)
	if !isValid {
		api.AuthorizationInvalidCustomMsg(w, "Valid token pair could not be extracted from your request.")
		return
	}

	refreshToken, isRefreshTokenExpired := findCurrentRefreshToken(refreshTokenValue)
	if refreshToken == nil {
		api.AuthorizationInvalidCustomMsg(w, "Refresh token does not exist.")
		return
	}
	if isRefreshTokenExpired {
		api.RefreshTokenExpiredGenericMsg(w)
		return
	}

	err = authentication.ValidateTokenPairByUsers(previousAccessTokenValue, refreshToken)
	if err != nil {
		api.AuthorizationInvalidCustomMsg(w, err.Error())
		return
	}

	err = authentication.ValidateTokenPairByCreationTimes(previousAccessTokenValue, refreshToken)
	if err != nil {
		api.AuthorizationInvalidCustomMsg(w, fmt.Sprintf("Couldn't generate new refresh token: %v", err))
		return
	}

	newTokenPair, err := authentication.GenerateNewTokenPair(refreshToken.Owner)
	if err != nil {
		api.InternalErrorHandlerGenericMsg(w, err)
		return
	}

	database.DeleteRefreshTokenWithValue(refreshToken.Value)

	var res refreshResponse = refreshResponse{
		Success:   true,
		NewTokens: newTokenPair,
	}

	api.RespondOk(w, res)
}

func validateRequestFormat(refreshBody refreshRequest) (previousAccessTokenValue string, refreshTokenValue string, isValid bool) {
	isValid = authentication.ValidateRefreshTokenFormat(refreshBody.RefreshTokenValue) && len(refreshBody.AccessTokenValue) > 20
	if isValid {
		previousAccessTokenValue = refreshBody.AccessTokenValue
		refreshTokenValue = refreshBody.RefreshTokenValue
	}
	return
}

func findCurrentRefreshToken(refreshTokenValue string) (refreshToken *database.RefreshToken, isExpired bool) {
	refreshToken = database.GetRefreshTokenWithUser(refreshTokenValue)

	if refreshToken == nil || refreshToken.Owner == nil {
		return nil, false
	}

	if time.Now().After(refreshToken.Expiration) {
		return nil, true
	}

	return refreshToken, false
}

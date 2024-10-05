package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	"github.com/sirupsen/logrus"
)

var authToken *jwtauth.JWTAuth
var accessTokenDuration time.Duration
var refreshTokenDuration time.Duration

var specialTypes = []string{"creator", "admin"}

const REFRESH_TOKEN_SIZE int = 512

type AuthTokenPair struct {
	AccessToken  string                `json:"accessToken"`
	RefreshToken database.RefreshToken `json:"refreshToken"`
}

func Initialize() {
	var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	var JWT_ACCESS_DURATION_MINUTES, _ = strconv.Atoi(os.Getenv("JWT_ACCESS_DURATION_MINUTES"))
	var REFRESH_TOKEN_DURATION_DAYS, _ = strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))

	accessTokenDuration = time.Duration(JWT_ACCESS_DURATION_MINUTES) * time.Minute
	refreshTokenDuration = time.Duration(REFRESH_TOKEN_DURATION_DAYS) * 24 * time.Hour

	authToken = jwtauth.New("HS256", []byte(JWT_SECRET_KEY), nil)
}

func RequireAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jwtauth.VerifyRequest(authToken, r, jwtauth.TokenFromHeader)

		if err != nil {
			switch err {
			case jwtauth.ErrNoTokenFound:
				api.AuthorizationMissingGenericMsg(w)
				return
			case jwtauth.ErrUnauthorized:
				api.AuthorizationInvalidGenericMsg(w)
				return
			case jwtauth.ErrExpired:
				api.AuthorizationExpiredGenericMsg(w)
				return
			}
		}

		var contextWithToken = context.WithValue(r.Context(), jwtauth.TokenCtxKey, token)
		next.ServeHTTP(w, r.WithContext(contextWithToken))
	})
}

func RequireSpecialType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userType, err := GetUserTypeFromToken(r)

		if err != nil || !slices.Contains(specialTypes, userType) {
			api.AuthorizationInvalidGenericMsg(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GenerateNewTokenPair(user *database.User) (*AuthTokenPair, error) {
	accessToken, err := generateNewAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := persistNewRefreshTokenForUser(user)
	if err != nil {
		return nil, err
	}

	var tokenPair = AuthTokenPair{
		AccessToken:  accessToken,
		RefreshToken: *refreshToken,
	}

	return &tokenPair, err
}

func ValidateRefreshTokenFormat(refreshTokenValue string) bool {
	return utils.IsValidRandomString(refreshTokenValue, REFRESH_TOKEN_SIZE)
}

func ValidateTokenPair(previousAccessTokenValue string, refreshTokenValue *database.RefreshToken) error {
	var parsedToken, err = jwtauth.VerifyToken(authToken, previousAccessTokenValue)
	if err != nil && err.Error() != "token is expired" {
		return err
	}

	var tokenIssuedAt = parsedToken.IssuedAt()
	tokenIssuedAt = tokenIssuedAt.UTC().Truncate(time.Second)
	var refreshTokenCreatedAt = refreshTokenValue.Expiration.Add(-refreshTokenDuration)
	refreshTokenCreatedAt = refreshTokenCreatedAt.UTC().Truncate(time.Second)

	if tokenIssuedAt.Compare(refreshTokenCreatedAt) != 0 {
		return errors.New("given tokens are not a pair")
	}

	return nil
}

func RemoveRefreshToken(refreshTokenValue string) bool {
	return database.DeleteRefreshTokenWithValue(refreshTokenValue)
}

func GetUserIdFromToken(r *http.Request) (int, error) {
	userIdClaim, err := getClaimFromToken("user_id", r)
	if err != nil {
		return 0, err
	}

	logrus.Tracef("User: %v", userIdClaim)

	userId, err := strconv.Atoi(userIdClaim)
	return userId, err
}

func GetUserTypeFromToken(r *http.Request) (string, error) {
	return getClaimFromToken("type", r)
}

func getClaimFromToken(claimKey string, r *http.Request) (string, error) {
	var tokenClaim string

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", err
	}

	tokenClaim = fmt.Sprintf("%v", claims[claimKey])
	return tokenClaim, err
}

func generateNewAccessToken(user *database.User) (string, error) {
	claims := map[string]interface{}{"user_id": user.ID, "email": user.Email, "username": user.Username, "type": user.UserType}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiryIn(claims, accessTokenDuration)
	_, accessToken, err := authToken.Encode(claims)
	return accessToken, err
}

func persistNewRefreshTokenForUser(user *database.User) (*database.RefreshToken, error) {
	var refreshTokenValue string = utils.RandomString(REFRESH_TOKEN_SIZE)
	refreshTokenExpiresAt := time.Now().Add(refreshTokenDuration)
	refreshToken, err := database.CreateRefreshTokenForUser(refreshTokenValue, user, refreshTokenExpiresAt)
	return refreshToken, err
}

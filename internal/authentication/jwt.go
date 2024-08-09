package authentication

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

var authToken *jwtauth.JWTAuth
var accessTokenDuration time.Duration
var refreshTokenDuration time.Duration

const REFRESH_TOKEN_SIZE int = 512

type AuthTokenPair struct {
	AccessToken  string                `json:"access_token"`
	RefreshToken database.RefreshToken `json:"refresh_token"`
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
		_, err := jwtauth.VerifyRequest(authToken, r, jwtauth.TokenFromHeader)

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
	refreshToken, err := database.UpdateRefreshTokenForUser(refreshTokenValue, user, refreshTokenExpiresAt)
	return refreshToken, err
}

package authentication

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
)

var authToken *jwtauth.JWTAuth
var accessTokenDuration time.Duration
var RefreshTokenDuration time.Duration

func Initialize() {
	var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	var JWT_ACCESS_DURATION_MINUTES, _ = strconv.Atoi(os.Getenv("JWT_ACCESS_DURATION_MINUTES"))
	var REFRESH_TOKEN_DURATION_DAYS, _ = strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))

	accessTokenDuration = time.Duration(JWT_ACCESS_DURATION_MINUTES) * time.Minute
	RefreshTokenDuration = time.Duration(REFRESH_TOKEN_DURATION_DAYS) * time.Hour * 24

	authToken = jwtauth.New("HS256", []byte(JWT_SECRET_KEY), nil)
}

func UseAuthenticator() func(http.Handler) http.Handler {
	return jwtauth.Authenticator
}

func UseVerifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(authToken)
}

func GenerateAccessToken(user *database.User) (string, error) {
	claims := map[string]interface{}{"user_id": user.ID, "email": user.Email, "username": user.Username}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiryIn(claims, accessTokenDuration)
	_, tokenString, err := authToken.Encode(claims)
	return tokenString, err
}

func GenerateRefreshToken(user *database.User) (string, error) {
	user.RefreshToken = utils.RandomString(512)
	err := database.SaveUser(*user)
	return user.RefreshToken, err
}

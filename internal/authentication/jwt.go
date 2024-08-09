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
var RefreshTokenDuration time.Duration

func Initialize() {
	var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	var JWT_ACCESS_DURATION_MINUTES, _ = strconv.Atoi(os.Getenv("JWT_ACCESS_DURATION_MINUTES"))
	var REFRESH_TOKEN_DURATION_DAYS, _ = strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))

	accessTokenDuration = time.Duration(JWT_ACCESS_DURATION_MINUTES) * time.Minute
	RefreshTokenDuration = time.Duration(REFRESH_TOKEN_DURATION_DAYS) * time.Hour * 24

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

func GenerateAccessToken(user *database.User) (string, error) {
	claims := map[string]interface{}{"user_id": user.ID, "email": user.Email, "username": user.Username, "type": user.UserType}
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

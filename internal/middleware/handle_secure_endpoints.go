package middleware

import (
	"context"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
)

var (
	CensoredBodyCtxKey = &specialRequestContextKey{"censoredBody"}
)

var secureRequestContextHandlers = map[string]func(*http.Request) context.Context{
	"/auth/login":          getBodyForRequestWithPassword,
	"/auth/register":       getBodyForRequestWithPassword,
	"/auth/password/reset": getBodyForPasswordResetRequest,
	"/auth/refresh":        getBodyForTokenRefreshRequest,
}

func HandleSecureEndpoints(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestContext context.Context = r.Context()

		if handler, canHandle := secureRequestContextHandlers[r.URL.String()]; canHandle {
			requestContext = handler(r)
		}

		next.ServeHTTP(w, r.WithContext(requestContext))
	})
}

func getBodyForRequestWithPassword(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosingWithCustomLimit(r, 1024<<2)
	jsonReqBody := logging_json_bodies.GetJsonBodyWithFieldsRemoved(reqBody, 0, "password", "recaptchaToken")

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &specialRequestContextValue{jsonReqBody})
}

func getBodyForPasswordResetRequest(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.LimitRequestBodySize(r)
	jsonReqBody := logging_json_bodies.GetJsonBodyWithFieldsRemoved(reqBody, 0, "newPassword")
	return context.WithValue(r.Context(), CensoredBodyCtxKey, &specialRequestContextValue{jsonReqBody})
}

func getBodyForTokenRefreshRequest(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.LimitRequestBodySize(r)
	jsonReqBody := logging_json_bodies.GetJsonBodyWithFieldsRemoved(reqBody, 10, "accessToken", "refreshToken")
	return context.WithValue(r.Context(), CensoredBodyCtxKey, &specialRequestContextValue{jsonReqBody})
}

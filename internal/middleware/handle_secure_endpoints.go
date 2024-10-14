package middleware

import (
	"context"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
)

type ContextKey struct {
	Name string
}

type ContextValue struct {
	Name string
}

var (
	CensoredBodyCtxKey = &ContextKey{"censoredBody"}
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
	var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosing(r)
	jsonReqBody := logging_json_bodies.HideFieldsFromJsonBody(reqBody, false, "password")
	return context.WithValue(r.Context(), CensoredBodyCtxKey, &ContextValue{jsonReqBody})
}

func getBodyForPasswordResetRequest(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosing(r)
	jsonReqBody := logging_json_bodies.HideFieldsFromJsonBody(reqBody, false, "newPassword")
	return context.WithValue(r.Context(), CensoredBodyCtxKey, &ContextValue{jsonReqBody})
}

func getBodyForTokenRefreshRequest(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosing(r)
	jsonReqBody := logging_json_bodies.HideFieldsFromJsonBody(reqBody, true, "accessToken", "refreshToken")
	return context.WithValue(r.Context(), CensoredBodyCtxKey, &ContextValue{jsonReqBody})
}

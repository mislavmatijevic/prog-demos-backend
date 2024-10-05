package logging

import (
	"context"
	"encoding/json"
	"net/http"
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
	"/auth/login":          getBodyForLoginRequest,
	"/auth/register":       getBodyForRegisterRequest,
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

func getBodyForLoginRequest(r *http.Request) context.Context {
	return removeFieldFromBody(r, "password")
}

func getBodyForRegisterRequest(r *http.Request) context.Context {
	return removeFieldFromBody(r, "password")
}

func getBodyForPasswordResetRequest(r *http.Request) context.Context {
	return removeFieldFromBody(r, "newPassword")
}

func getBodyForTokenRefreshRequest(r *http.Request) context.Context {
	return hideFieldFromBody(r, "accessToken", "refreshToken")
}

func removeFieldFromBody(r *http.Request, fieldName string) context.Context {
	var reqBody = ReadRequestBodyWithoutClosing(r)
	var jsonReqBody string
	var data map[string]interface{}

	if reqBody != nil {
		json.Unmarshal(reqBody, &data)
		delete(data, fieldName)
		reqBody, _ = json.Marshal(&data)
		jsonReqBody = string(reqBody)
	}

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &ContextValue{jsonReqBody})
}

func hideFieldFromBody(r *http.Request, fields ...string) context.Context {
	var reqBody = ReadRequestBodyWithoutClosing(r)
	var jsonReqBody string
	var data map[string]string

	if reqBody != nil {
		json.Unmarshal(reqBody, &data)
		for _, fieldName := range fields {
			data[fieldName] = data[fieldName][0:10] + "... [HIDDEN]"
		}
		reqBody, _ = json.Marshal(&data)
		jsonReqBody = string(reqBody)
	}

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &ContextValue{jsonReqBody})
}

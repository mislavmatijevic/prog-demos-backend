package logging

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey struct {
	Name string
}

type contextValue struct {
	Name string
}

var (
	CensoredBodyCtxKey = &contextKey{"censoredBody"}
)

var secureRequestContextHandlers = map[string]func(*http.Request) context.Context{
	"/auth/login":          getBodyForLoginRequest,
	"/auth/register":       getBodyForRegisterRequest,
	"/auth/password/reset": getBodyForPasswordResetRequest,
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

func removeFieldFromBody(r *http.Request, fieldName string) context.Context {
	var reqBody any
	var jsonReqBody string
	var data map[string]interface{}
	json.NewDecoder(r.Body).Decode(&reqBody)

	if reqBody != nil {
		jsonRawBody, _ := json.Marshal(reqBody)
		json.Unmarshal(jsonRawBody, &data)
		delete(data, fieldName)
		jsonRawBody, _ = json.Marshal(&data)
		jsonReqBody = string(jsonRawBody)
	}

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &contextValue{jsonReqBody})
}

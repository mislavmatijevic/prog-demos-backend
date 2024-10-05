package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	var reqBody = readRequestBodyWithoutClosing(r)
	var jsonReqBody string
	var data map[string]interface{}

	if reqBody != nil {
		json.Unmarshal(reqBody, &data)
		delete(data, fieldName)
		reqBody, _ = json.Marshal(&data)
		jsonReqBody = string(reqBody)
	}

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &contextValue{jsonReqBody})
}

func hideFieldFromBody(r *http.Request, fields ...string) context.Context {
	var reqBody = readRequestBodyWithoutClosing(r)
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

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &contextValue{jsonReqBody})
}

func readRequestBodyWithoutClosing(r *http.Request) []byte {
	var bodyBuffer bytes.Buffer
	io.Copy(&bodyBuffer, r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBuffer.Bytes()), 1024))
	return bodyBuffer.Bytes()
}

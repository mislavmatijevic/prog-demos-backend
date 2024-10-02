package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
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
	var reqBody = ReadRequestBodyWithoutClosing(r)
	var jsonReqBody string
	var data map[string]interface{}

	if reqBody != nil {
		json.Unmarshal(reqBody, &data)
		delete(data, fieldName)
		reqBody, _ = json.Marshal(&data)
		jsonReqBody = string(reqBody)
	}

	logrus.Info(jsonReqBody)

	return context.WithValue(r.Context(), CensoredBodyCtxKey, &contextValue{jsonReqBody})
}

func ReadRequestBodyWithoutClosing(r *http.Request) []byte {
	var bodyBuffer bytes.Buffer
	io.Copy(&bodyBuffer, r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBuffer.Bytes()), 1024))
	return bodyBuffer.Bytes()
}

package api

import (
	"net/http"
)

type defaultResponseBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (e defaultResponseBody) Error() string {
	return e.Message
}

func writeError(w http.ResponseWriter, message string, code int) {
	res := defaultResponseBody{
		Success: false,
		Message: message,
	}

	RespondWithStatus(w, res, code)
}

var (
	RequestErrorHandlerGenericMsg = func(w http.ResponseWriter, err error) {
		writeError(w, "Request invalid.", http.StatusBadRequest)
	}
	RequestErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writeError(w, errorMessage, http.StatusBadRequest)
	}
	InternalErrorHandlerGenericMsg = func(w http.ResponseWriter, err error) {
		writeError(w, "An Unexpected Error Occurred.", http.StatusInternalServerError)
	}
	InternalErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writeError(w, errorMessage, http.StatusInternalServerError)
	}
	AuthorizationMissingGenericMsg = func(w http.ResponseWriter) {
		writeError(w, "Valid authorization header missing.", http.StatusUnauthorized)
	}
	AuthorizationExpiredGenericMsg = func(w http.ResponseWriter) {
		writeError(w, "Authorization token expired.", http.StatusForbidden)
	}
	AuthorizationInvalidGenericMsg = func(w http.ResponseWriter) {
		writeError(w, "Authorization token is invalid.", http.StatusForbidden)
	}
	AuthorizationInvalidCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writeError(w, errorMessage, http.StatusForbidden)
	}
	RefreshTokenExpiredGenericMsg = func(w http.ResponseWriter) {
		writeError(w, "Refresh token expired.", http.StatusForbidden)
	}
	TooEarlyErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writeError(w, errorMessage, http.StatusTooEarly)
	}
	NotFoundHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writeError(w, errorMessage, http.StatusNotFound)
	}
)

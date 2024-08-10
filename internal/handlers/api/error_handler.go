package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Error struct {
	Status  string
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func writerError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Status:  "failed",
		Message: message,
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandlerGenericMsg = func(w http.ResponseWriter, err error) {
		if err != nil {
			log.Error(err)
		}
		writerError(w, "Request invalid.", http.StatusBadRequest)
	}
	RequestErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writerError(w, errorMessage, http.StatusBadRequest)
	}
	InternalErrorHandlerGenericMsg = func(w http.ResponseWriter, err error) {
		log.Error(err)
		writerError(w, "An Unexpected Error Occurred.", http.StatusInternalServerError)
	}
	InternalErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writerError(w, errorMessage, http.StatusInternalServerError)
	}
	AuthorizationMissingGenericMsg = func(w http.ResponseWriter) {
		writerError(w, "Valid authorization header missing.", http.StatusUnauthorized)
	}
	AuthorizationExpiredGenericMsg = func(w http.ResponseWriter) {
		writerError(w, "Authorization token expired.", http.StatusForbidden)
	}
	AuthorizationInvalidGenericMsg = func(w http.ResponseWriter) {
		writerError(w, "Authorization token is invalid.", http.StatusForbidden)
	}
	TooEarlyErrorHandlerCustomMsg = func(w http.ResponseWriter, errorMessage string) {
		writerError(w, errorMessage, http.StatusTooEarly)
	}
)

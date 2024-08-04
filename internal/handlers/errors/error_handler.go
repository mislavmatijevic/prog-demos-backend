package errors

import (
	"encoding/json"
	"net/http"
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
		writerError(w, err.Error(), http.StatusBadRequest)
	}
	RequestErrorHandlerCustomMsg = func(w http.ResponseWriter, err string) {
		writerError(w, err, http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writerError(w, "An Unexpected Error Occurred.", http.StatusInternalServerError)
	}
)

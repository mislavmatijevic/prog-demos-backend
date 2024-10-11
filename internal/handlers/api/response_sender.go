package api

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/logging/logging_responses"
)

func RespondOk(w http.ResponseWriter, res interface{}) {
	RespondWithStatus(w, res, http.StatusOK)
}

func RespondWithStatus(w http.ResponseWriter, res interface{}, status int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	jsonRes, _ := json.Marshal(res)
	w.Write(jsonRes)

	logging_responses.LogResponse(status, w.Header(), jsonRes)
}

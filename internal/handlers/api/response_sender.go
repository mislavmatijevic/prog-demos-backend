package api

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/logging"
)

func RespondOk(w http.ResponseWriter, res interface{}) {
	RespondWithStatus(w, res, http.StatusOK)
}

func RespondWithStatus(w http.ResponseWriter, res interface{}, status int) {
	jsonRes, _ := json.Marshal(res)
	w.WriteHeader(status)
	w.Header().Add("content-type", "application/json")
	w.Write(jsonRes)

	logging.LogResponse(status, jsonRes)
}

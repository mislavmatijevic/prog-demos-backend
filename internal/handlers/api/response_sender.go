package api

import (
	"encoding/json"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/monitoring"
)

func RespondOkWithDefaultBody(w http.ResponseWriter) {
	res := defaultResponseBody{
		Success: true,
		Message: "Request succeeded",
	}

	RespondOk(w, res)
}

func RespondOk(w http.ResponseWriter, res interface{}) {
	RespondWithStatus(w, res, http.StatusOK)
}

func RespondWithStatus(w http.ResponseWriter, res interface{}, status int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	jsonRes, _ := json.Marshal(res)
	w.Write(jsonRes)

	monitoring.LogResponse(status, w.Header(), jsonRes)
}

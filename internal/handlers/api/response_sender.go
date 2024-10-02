package api

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func RespondOk(w http.ResponseWriter, res interface{}) {
	RespondWithStatus(w, res, http.StatusOK)
}

func RespondWithStatus(w http.ResponseWriter, res interface{}, status int) {
	jsonRes, _ := json.Marshal(res)
	w.WriteHeader(status)
	w.Header().Add("content-type", "application/json")
	w.Write(jsonRes)

	logResponse(status, jsonRes)
}

func logResponse(status int, jsonRes []byte) {
	var bodyOutputLimit = 50
	if status >= 400 {
		bodyOutputLimit = 500
	}
	resBodyLength := len(jsonRes)
	if bodyOutputLimit > resBodyLength {
		bodyOutputLimit = resBodyLength
	}
	log.Trace("HTTP ", status, string(jsonRes)[0:bodyOutputLimit])
}

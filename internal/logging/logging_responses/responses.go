package logging_responses

import (
	"net/http"

	chimiddle "github.com/go-chi/chi/v5/middleware"
	"github.com/mislavmatijevic/prog-demos-backend/internal/logging"
	log "github.com/sirupsen/logrus"
)

func LogResponse(status int, header http.Header, jsonRes []byte) {
	requestId := header.Get(chimiddle.RequestIDHeader)

	var level = log.InfoLevel
	if status >= 400 {
		level = log.WarnLevel
	}

	resBody := logging.HideFieldsFromJsonBody(jsonRes, true, "tokens.accessToken", "tokens.refreshToken.value")

	log.WithFields(
		log.Fields{
			"context":    "response",
			"request_id": requestId,
			"status":     status,
			"body":       resBody,
		},
	).Log(level, http.StatusText(status))
}

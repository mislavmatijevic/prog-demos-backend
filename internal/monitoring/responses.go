package monitoring

import (
	"net/http"

	chimiddle "github.com/go-chi/chi/v5/middleware"
	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
	log "github.com/sirupsen/logrus"
)

func LogResponse(status int, header http.Header, jsonRes []byte) {
	requestId := header.Get(chimiddle.RequestIDHeader)

	var level = log.InfoLevel
	if status >= 400 {
		level = log.WarnLevel
	}

	resBody := logging_json_bodies.HideFieldsFromJsonBody(jsonRes, 10, "tokens.accessToken", "tokens.refreshToken.value")

	log.WithFields(
		log.Fields{
			"context":    "response",
			"request_id": requestId,
			"status":     status,
			"body":       resBody,
		},
	).Log(level, http.StatusText(status))
}

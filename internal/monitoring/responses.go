package monitoring

import (
	"net/http"

	chimiddle "github.com/go-chi/chi/v5/middleware"
	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
	log "github.com/sirupsen/logrus"
)

func LogResponse(status int, header http.Header, jsonRes []byte) {
	requestId := header.Get(chimiddle.RequestIDHeader)

	var level log.Level

	if status >= 500 {
		level = log.ErrorLevel
	} else if status >= 400 {
		level = log.WarnLevel
	} else {
		level = log.InfoLevel
	}

	resBodyForLogging := hideUnnecessaryPropertiesFromBody(jsonRes)

	log.WithFields(
		log.Fields{
			"context":    "response",
			"request_id": requestId,
			"status":     status,
			"body":       resBodyForLogging,
		},
	).Log(level, http.StatusText(status))
}

func hideUnnecessaryPropertiesFromBody(jsonRes []byte) string {
	var optimizedResBody = string(jsonRes)
	optimizedResBody = logging_json_bodies.GetJsonBodyWithFieldsRemoved([]byte(optimizedResBody), 0, "topics")
	optimizedResBody = logging_json_bodies.GetJsonBodyWithFieldsRemoved([]byte(optimizedResBody), 10, "tokens.accessToken", "tokens.refreshToken.value")
	optimizedResBody = logging_json_bodies.GetJsonBodyWithFieldsRemoved([]byte(optimizedResBody), 0, "helpStep.helperCode", "helpStep.helperText")
	return optimizedResBody
}

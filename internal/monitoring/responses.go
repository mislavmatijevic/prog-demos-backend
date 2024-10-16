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

	var resBody string
	resBody = logging_json_bodies.GetJsonBodyWithFieldsRemoved(jsonRes, 10, "tokens.accessToken", "tokens.refreshToken.value")
	resBody = logging_json_bodies.GetJsonBodyWithFieldsRemoved([]byte(resBody), 0, "helpStep.helperCode", "helpStep.helperText")

	log.WithFields(
		log.Fields{
			"context":    "response",
			"request_id": requestId,
			"status":     status,
			"body":       resBody,
		},
	).Log(level, http.StatusText(status))
}

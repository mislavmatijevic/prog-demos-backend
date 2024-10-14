package middleware

import (
	"net/http"
	"strconv"
	"strings"

	chimiddle "github.com/go-chi/chi/v5/middleware"
	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
	log "github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonReqBody string = ""

		if censoredBody := r.Context().Value(CensoredBodyCtxKey); censoredBody != nil {
			jsonReqBody = censoredBody.(*specialRequestContextValue).Name
		} else if largePayloadBody := r.Context().Value(LargePayloadBodyCtxKey); largePayloadBody != nil {
			jsonReqBody = largePayloadBody.(*specialRequestContextValue).Name
		} else {
			var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosing(r)
			if reqBody != nil {
				jsonReqBody = string(reqBody)
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "\n", "")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "  ", " ")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, " \"", "\"")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "\" ", "\"")
				jsonReqBody = strings.Trim(jsonReqBody, " ")
			}
		}

		var authedUserId string = "NO_AUTH"
		userId, _ := authentication.GetUserIdFromRequest(r)
		if userId != 0 {
			authedUserId = strconv.Itoa(userId)
		}

		requestId := chimiddle.GetReqID(r.Context())
		w.Header().Add(chimiddle.RequestIDHeader, requestId)

		maxLogBodySize := len(jsonReqBody)
		if maxLogBodySize > 1024 {
			maxLogBodySize = 1024
		}

		log.WithFields(
			log.Fields{
				"context":    "request",
				"request_id": requestId,
				"method":     r.Method,
				"url":        r.URL,
				"user":       authedUserId,
				"body":       jsonReqBody,
				"request_ip": r.RemoteAddr,
			},
		).Infof("%s %s", r.Method, r.URL)

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

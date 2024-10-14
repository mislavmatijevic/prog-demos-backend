package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
)

var (
	RequestBodyForLoggingCtxKey = &specialRequestContextKey{"censoredBody"}
)

func LimitNonHandledRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonReqBody = ""

		if censoredBody := r.Context().Value(CensoredBodyCtxKey); censoredBody != nil {
			jsonReqBody = censoredBody.(*specialRequestContextValue).Name
		} else if largePayloadBody := r.Context().Value(LargePayloadBodyCtxKey); largePayloadBody != nil {
			jsonReqBody = largePayloadBody.(*specialRequestContextValue).Name
		} else {
			var reqBody = logging_json_bodies.LimitRequestBodySize(r)
			if reqBody != nil {
				jsonReqBody = string(reqBody)
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "\n", "")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "  ", " ")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, " \"", "\"")
				jsonReqBody = strings.ReplaceAll(jsonReqBody, "\" ", "\"")
				jsonReqBody = strings.Trim(jsonReqBody, " ")
			}
		}

		var requestContext = context.WithValue(r.Context(), RequestBodyForLoggingCtxKey, &specialRequestContextValue{jsonReqBody})
		next.ServeHTTP(w, r.WithContext(requestContext))
	})
}

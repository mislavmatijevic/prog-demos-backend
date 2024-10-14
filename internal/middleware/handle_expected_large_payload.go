package middleware

import (
	"context"
	"net/http"
	"regexp"

	"github.com/mislavmatijevic/prog-demos-backend/internal/middleware/logging/logging_json_bodies"
)

var (
	LargePayloadBodyCtxKey = &specialRequestContextKey{"largeBody"}
)

var largePayloadRouteHandlers = map[*regexp.Regexp]func(*http.Request) context.Context{
	regexp.MustCompile(`^\/tasks[\/a-z0-9]*$`): getBodyForTaskExecutionRequest,
}

func HandleExpectedLargePayload(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			next.ServeHTTP(w, r)
			return
		}

		var requestContext context.Context = r.Context()

		for regexp, handler := range largePayloadRouteHandlers {
			hitRoute := r.URL.String()
			if regexp.Match([]byte(hitRoute)) {
				requestContext = handler(r)
			}
		}

		next.ServeHTTP(w, r.WithContext(requestContext))
	})
}

func getBodyForTaskExecutionRequest(r *http.Request) context.Context {
	var reqBody = logging_json_bodies.ReadRequestBodyWithoutClosingWithCustomLimit(r, 1024<<4)
	jsonReqBody := logging_json_bodies.GetJsonBodyWithFieldsRemoved(reqBody, 128, "solutionCode")
	return context.WithValue(r.Context(), LargePayloadBodyCtxKey, &specialRequestContextValue{jsonReqBody})
}

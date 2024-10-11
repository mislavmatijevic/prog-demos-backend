package logging_requests

import (
	"net/http"
	"strconv"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/logging"
	log "github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonReqBody string = ""
		var bodyOutputLimit = 500

		if censoredBody := r.Context().Value(logging.CensoredBodyCtxKey); censoredBody != nil {
			jsonReqBody = censoredBody.(*logging.ContextValue).Name
		} else {
			var reqBody = logging.ReadRequestBodyWithoutClosing(r)
			if reqBody != nil {
				reqBodyLength := len(reqBody)
				if bodyOutputLimit > reqBodyLength {
					bodyOutputLimit = reqBodyLength - 1
				}

				jsonReqBody = string(reqBody)[0:bodyOutputLimit]
			}
		}

		var authedUser string = "NO_AUTH"
		userId, _ := authentication.GetUserIdFromRequest(r)
		if userId != 0 {
			authedUser = strconv.Itoa(userId)
		}

		log.WithFields(log.Fields{"context": "request", "method": r.Method, "url": r.URL, "user": authedUser, "body": jsonReqBody}).Trace()
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

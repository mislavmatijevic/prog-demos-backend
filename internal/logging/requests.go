package logging

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonReqBody string = ""
		var bodyOutputLimit = 500

		if censoredBody := r.Context().Value(CensoredBodyCtxKey); censoredBody != nil {
			jsonReqBody = censoredBody.(*contextValue).Name
		} else {
			var reqBody = readRequestBodyWithoutClosing(r)
			if reqBody != nil {
				reqBodyLength := len(reqBody)
				if bodyOutputLimit > reqBodyLength {
					bodyOutputLimit = reqBodyLength - 1
				}

				jsonReqBody = string(reqBody)[0:bodyOutputLimit]
			}
		}

		log.Tracef("%s %s %s", r.Method, r.URL, jsonReqBody)
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

package logging

import (
	"encoding/json"
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
			var reqBody any
			json.NewDecoder(r.Body).Decode(&reqBody)
			if reqBody != nil {
				jsonRawBody, _ := json.Marshal(reqBody)
				jsonReqBody = string(jsonRawBody)[0:bodyOutputLimit]
			}
		}

		log.Tracef("%s %s %s", r.Method, r.URL, jsonReqBody)
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

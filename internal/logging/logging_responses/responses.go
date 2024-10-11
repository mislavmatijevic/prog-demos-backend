package logging_responses

import log "github.com/sirupsen/logrus"

func LogResponse(status int, jsonRes []byte) {
	var bodyOutputLimit = 150
	if status >= 400 {
		bodyOutputLimit = 500
	}
	resBodyLength := len(jsonRes)
	if bodyOutputLimit > resBodyLength {
		bodyOutputLimit = resBodyLength - 1
	}
	log.WithFields(log.Fields{"context": "response", "status": status, "body": string(jsonRes[0:bodyOutputLimit])}).Trace()
}

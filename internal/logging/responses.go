package logging

import log "github.com/sirupsen/logrus"

func LogResponse(status int, jsonRes []byte) {
	var bodyOutputLimit = 150
	if status >= 400 {
		bodyOutputLimit = 500
	}
	resBodyLength := len(jsonRes)
	if bodyOutputLimit > resBodyLength {
		bodyOutputLimit = resBodyLength
	}
	log.Tracef("HTTP %d %s", status, string(jsonRes)[0:bodyOutputLimit])
}

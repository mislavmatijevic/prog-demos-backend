package integrations

import (
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

func Initialize() {
	if !utils.IsProd() {
		log.Info("Not PROD, skipping integrations...")
		return
	}

	initializeGithubIntegration()

	return
}

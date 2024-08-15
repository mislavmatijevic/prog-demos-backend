package mailing

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

var client *mail.Client
var isMailingInitialized = false

func Initialize() {
	var PROD = os.Getenv("PROD")
	if PROD != "1" {
		log.Info("Did not detect PROD flag, skipping mailing...")
		return
	}

	var SMTP_URL = os.Getenv("SMTP_URL")
	var SMTP_PORT, _ = strconv.Atoi(os.Getenv("SMTP_PORT"))
	var SMTP_USERNAME = os.Getenv("SMTP_USERNAME")
	var SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")

	var err error
	client, err = mail.NewClient(
		SMTP_URL,
		mail.WithPort(SMTP_PORT),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(SMTP_USERNAME),
		mail.WithPassword(SMTP_PASSWORD),
	)

	if err == nil {
		isMailingInitialized = true
	} else {
		log.Fatalf("Failed to initialize mail service: %v", err)
	}
}

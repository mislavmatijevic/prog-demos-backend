package mailing

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

var client *mail.Client
var isMailingInitialized = false

func Initialize() {
	if !utils.IsProd() {
		log.Info("Not PROD, skipping mailing...")
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

func SendMailToUser(user database.User, templateHtmlFilename string, subject string) error {
	if !isMailingInitialized {
		log.Infof("Skipping mail with template %s for user %s.", templateHtmlFilename, user.Username)
		return nil
	}

	tmpl, err := template.ParseFiles("./internal/mailing/templates/" + templateHtmlFilename)
	if err != nil {
		return fmt.Errorf("failed to parse template file '%s' due to error: %v", templateHtmlFilename, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, user); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	m := mail.NewMsg()
	m.From("no-reply@prog_demos.com")
	m.To(user.Email)
	m.Subject(subject)

	m.SetBodyString(mail.TypeTextHTML, body.String())

	if err := client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

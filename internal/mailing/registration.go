package mailing

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	log "github.com/sirupsen/logrus"
	"github.com/wneessen/go-mail"
)

func SendRegistrationMail(user database.User) error {
	if !isMailingInitialized {
		log.Infof("Skipping registration mail for user %s.", user.Username)
		return nil
	}

	tmpl, err := template.ParseFiles("./internal/mailing/templates/welcome-mail.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, user); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	m := mail.NewMsg()
	m.From("no-reply@prog_demos.com")
	m.To(user.Email)
	m.Subject("[prog_demos] Dovrši registraciju!")

	m.SetBodyString(mail.TypeTextHTML, body.String())

	if err := client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

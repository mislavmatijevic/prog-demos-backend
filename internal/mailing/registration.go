package mailing

import (
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

func SendRegistrationMail(user database.User) error {
	return SendMailToUser(user, "welcome-mail.html", "[prog_demos] Dovrši registraciju!")
}

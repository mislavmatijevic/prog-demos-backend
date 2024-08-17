package mailing

import (
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
)

func SendPasswordRequestMail(user database.User) error {
	return SendMailToUser(user, "password-reset.html", "[prog_demos] Promijeni lozinku!")
}

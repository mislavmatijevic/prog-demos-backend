package database

import (
	"database/sql"
	"errors"
)

func RegisterNewUser(userInfo User) (int, error) {
	var alreadyExistingUser User

	Instance.db.Where("username = @Username OR email = @Email",
		sql.Named("Username", userInfo.Username),
		sql.Named("Email", userInfo.Email),
	).Find(&alreadyExistingUser)

	if alreadyExistingUser.ID != 0 {
		return -1, errors.New("User already exists")
	}

	Instance.db.Create(&userInfo)
	return userInfo.ID, nil
}

package database

import (
	"database/sql"
	"errors"

	log "github.com/sirupsen/logrus"
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

func GetUserByEmail(email string) *User {
	var foundUser User

	var result = Instance.db.Where("email = @Email", sql.Named("Email", email)).Find(&foundUser)

	if result.Error != nil {
		log.Error("Error fetching user: ", result.Error)
		return nil
	}

	return &foundUser
}

func AssignRefreshTokenToUser(user User) error {
	result := Instance.db.Save(user)
	return result.Error
}

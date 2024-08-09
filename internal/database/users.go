package database

import (
	"database/sql"
	"errors"

	log "github.com/sirupsen/logrus"
)

func RegisterNewUser(userInfo User) (*User, error) {
	var alreadyExistingUser User

	Instance.db.Where("username = @Username OR email = @Email",
		sql.Named("Username", userInfo.Username),
		sql.Named("Email", userInfo.Email),
	).Find(&alreadyExistingUser)

	if alreadyExistingUser.ID != 0 {
		return nil, errors.New("User already exists")
	}

	Instance.db.Create(&userInfo)
	return &userInfo, nil
}

func GetUserByEmail(email string) *User {
	return getUserByCondition("email = ?", email)
}

func GetUserByUsername(username string) *User {
	return getUserByCondition("username = ?", username)
}

func getUserByCondition(query interface{}, args ...interface{}) *User {
	var foundUser User

	var result = Instance.db.Where(query, args).First(&foundUser)

	if result.Error != nil {
		log.Error("Error fetching user: ", result.Error)
		return nil
	}

	return &foundUser
}

func SaveUser(user User) error {
	result := Instance.db.Save(user)
	return result.Error
}

func SetUserActivated(activationToken string) (*User, error) {
	user := getUserByCondition("activation_token = ?", activationToken)
	if user == nil {
		return nil, errors.New("token does not exist")
	}

	user.IsActivated = true
	user.ActivationToken = ""

	err := SaveUser(*user)
	return user, err
}

func DeleteUser(user *User) error {
	result := Instance.db.Delete(user)
	return result.Error
}

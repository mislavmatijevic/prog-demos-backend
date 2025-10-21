package database

import (
	"database/sql"
	"errors"
)

func RegisterNewUser(userInfo User) (*User, error) {
	var alreadyExistingUser User

	Instance.db.Where("username = @Username OR email = @Email",
		sql.Named("Username", userInfo.Username),
		sql.Named("Email", userInfo.Email),
	).Find(&alreadyExistingUser)

	if alreadyExistingUser.ID != 0 {
		return nil, errors.New("user already exists")
	}

	Instance.db.Create(&userInfo)
	return &userInfo, nil
}

func GetUserById(id int) (*User, error) {
	var user User
	result := Instance.db.First(&user, id)
	return &user, result.Error
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
	user.ActivationToken = WrappedNullString{}

	err := SaveUser(*user)
	return user, err
}

func DeleteUser(user *User) error {
	result := Instance.db.Delete(user)
	return result.Error
}

func ChangeUserPassword(user *User, newPasswordHash string) error {
	user.Password = newPasswordHash
	RemovePasswordReset(user)
	return SaveUser(*user)
}

func RemovePasswordReset(user *User) {
	user.PasswordResetToken = WrappedNullString{NullString: sql.NullString{Valid: false}}
	user.PasswordResetExpiry = WrappedNullTime{NullTime: sql.NullTime{Valid: false}}
	SaveUser(*user)
}

func GetUserByPasswordResetToken(questionableToken string) (*User, error) {
	user := getUserByCondition("password_reset_token = ?", questionableToken)
	if user == nil {
		return nil, errors.New("password reset token does not exist")
	}
	return user, nil
}

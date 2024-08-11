package database

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func UpdateRefreshTokenForUser(newRefreshTokenValue string, user *User, expiresAt time.Time) (*RefreshToken, error) {
	var refreshToken RefreshToken
	var result = Instance.db.Preload("Owner").Where("id_user = ?", user.ID).First(&refreshToken)

	if result.Error == nil {
		Instance.db.Delete(refreshToken)
	}

	refreshToken = RefreshToken{
		Value:      newRefreshTokenValue,
		Owner:      user,
		Expiration: expiresAt,
	}

	tokenCreationResult := Instance.db.Save(&refreshToken)
	return &refreshToken, tokenCreationResult.Error
}

func GetRefreshTokenWithUser(refreshTokenValue string) *RefreshToken {
	var refreshToken RefreshToken

	var result = Instance.db.Preload("Owner").Where("value = ?", refreshTokenValue).First(&refreshToken)

	if result.Error != nil {
		log.Error("Error fetching refresh token: ", result.Error)
		return nil
	}

	return &refreshToken
}

func DeleteRefreshTokenWithValue(refreshTokenValue string) bool {
	var refreshToken RefreshToken
	var result = Instance.db.Where("value = ?", refreshTokenValue).Unscoped().Delete(&refreshToken)
	return result.RowsAffected == 1
}

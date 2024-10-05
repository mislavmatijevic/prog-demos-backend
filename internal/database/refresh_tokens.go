package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func CreateRefreshTokenForUser(newRefreshTokenValue string, user *User, expiresAt time.Time) (*RefreshToken, error) {
	deleteExpiredRefreshTokensForUser(user)

	var refreshToken = RefreshToken{
		Value:      newRefreshTokenValue,
		Owner:      user,
		Expiration: expiresAt,
	}

	tokenCreationResult := Instance.db.Save(&refreshToken)
	return &refreshToken, tokenCreationResult.Error
}

func deleteExpiredRefreshTokensForUser(user *User) {
	var expiredRefreshTokens []RefreshToken = make([]RefreshToken, 0)
	Instance.db.Preload("Owner").Where("id_user = ?", user.ID).Where("expiration < ?", time.Now()).Find(&expiredRefreshTokens)

	expiredRefreshTokensCount := len(expiredRefreshTokens)
	if expiredRefreshTokensCount > 0 {
		for _, token := range expiredRefreshTokens {
			Instance.db.Delete(token)
		}
		log.Trace(fmt.Sprintf("Deleted %d expired refresh tokens belonging to user %s.", expiredRefreshTokensCount, user.Username))
	}
}

func GetRefreshTokenWithUser(refreshTokenValue string) *RefreshToken {
	var refreshToken RefreshToken

	var result = Instance.db.Preload("Owner").Where("value = ?", refreshTokenValue).Find(&refreshToken)

	if result.Error != nil {
		log.Error("Error fetching refresh token: ", result.Error)
		return nil
	}

	return &refreshToken
}

func DeleteRefreshTokenWithValue(refreshTokenValue string) (success bool) {
	var refreshToken RefreshToken
	var result = Instance.db.Where("value = ?", refreshTokenValue).Unscoped().Delete(&refreshToken)
	return result.RowsAffected == 1
}

package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"math/big"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var validRefreshTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		result[i] = letters[num.Int64()]
	}
	return string(result)
}

func IsValidRandomString(value string, n int) bool {
	return validRefreshTokenRegex.MatchString(value) && len(value) == n
}

func CreateSecureHash(text string) (string, error) {
	bytePassword := []byte(text)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CreateShortHash(text string) string {
	sha := sha1.New()
	sha.Write([]byte(text))
	return fmt.Sprintf("%x", sha.Sum(nil))[0:10]
}

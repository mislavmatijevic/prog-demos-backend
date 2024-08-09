package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
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

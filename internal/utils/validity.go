package utils

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func IsEmailValid(e string) bool {
	if len(e) > 100 {
		return false
	}
	return emailRegex.MatchString(e)
}

func GetTrimmedStringWithValue(s string) (containsChars bool, trimmedString string) {
	trimmedString = strings.Trim(s, " ")
	containsChars = len(trimmedString) != 0
	return containsChars, trimmedString
}

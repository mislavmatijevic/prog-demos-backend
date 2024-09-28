package utils

import "os"

var isProd = os.Getenv("PROD")

func IsProd() bool {
	return isProd == "1"
}

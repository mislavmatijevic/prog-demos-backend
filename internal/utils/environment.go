package utils

import "os"

var isProd = os.Getenv("PROD")
var useLoki = os.Getenv("USE_LOKI")

func IsProd() bool {
	return isProd == "1"
}

func UseLoki() bool {
	return useLoki == "1"
}

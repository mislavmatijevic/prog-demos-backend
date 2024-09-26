package utils

import "math"

func RoundNumberDownToTwoDecimals[T float32 | float64](numberWithManyDecimals T) float32 {
	return float32(math.Round(float64(numberWithManyDecimals)*100) / 100)
}

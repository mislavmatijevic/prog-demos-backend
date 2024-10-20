package utils

import "math"

func FloatToInt[T float32 | float64](numberWithManyDecimals T) int {
	return int(math.Round(float64(numberWithManyDecimals)))
}

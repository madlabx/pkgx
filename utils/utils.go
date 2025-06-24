package utils

import (
	"math"
	"runtime"
	"strconv"
	"strings"
)

/*
var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString returns a random string with a fixed length

	func RandomString(n int, allowedChars ...[]rune) string {
		var letters []rune

		if len(allowedChars) == 0 {
			letters = defaultLetters
		} else {
			letters = allowedChars[0]
		}

		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		return string(b)
	}
*/

// Round 对 float64 进行四舍五入，保留指定的小数位数
// digits 表示要保留的小数位数
func Round(num float64, digits int) float64 {
	shift := math.Pow(10, float64(digits))
	return math.Round(num*shift) / shift
}

func InArray[T comparable](x T, ss []T) bool {
	for _, s := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func InRange[T comparable](x T, ss ...T) bool {
	for _, s := range ss {
		if x == s {
			return true
		}
	}
	return false
}


package util

import (
	"math/rand"
	"strconv"
)

func GenerateOTP() string {
	otp := rand.Intn(900000) + 100000 // 100000 ile 999999 arasında sayı üret
	return strconv.Itoa(otp)          // Sayıyı string'e çevir
}

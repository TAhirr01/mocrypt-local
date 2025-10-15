package util

import (
	"strconv"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP(t *testing.T) {
	otp := GenerateOTP()

	// Check that OTP is not empty
	assert.NotEmpty(t, otp)

	// Check that OTP length is 6
	assert.Len(t, otp, 6)

	// Check that all characters are digits
	for _, c := range otp {
		assert.True(t, unicode.IsDigit(c), "OTP contains non-digit character: %c", c)
	}

	// Convert to integer and check range
	num, err := strconv.Atoi(otp)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, num, 100000)
	assert.LessOrEqual(t, num, 999999)
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	// Optional: check if multiple OTPs are usually different
	otp1 := GenerateOTP()
	otp2 := GenerateOTP()
	assert.NotEqual(t, otp1, otp2, "OTPs should generally differ (though rare collisions are possible)")
}

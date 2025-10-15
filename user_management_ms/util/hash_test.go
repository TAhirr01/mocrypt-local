package util

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"testing"
)

func HashPINMock(pin string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func TestHashTable(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		hash     string
		expected bool
	}{
		{"hash matches correct PIN", "016016", HashPINMock("016016"), true},
		{"hash does not match incorrect PIN", "016016", HashPINMock("016017"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := VerifyPIN(tt.pin, tt.hash)
			assert.Equal(t, tt.expected, match)
		})
	}
}

package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP creates a 6-digit numeric OTP
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	// Format as 6-digit string with leading zeros
	return fmt.Sprintf("%06d", n.Int64()), nil
}

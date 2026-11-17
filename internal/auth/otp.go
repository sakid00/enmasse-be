package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const OTPLength = 6

func RandomOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

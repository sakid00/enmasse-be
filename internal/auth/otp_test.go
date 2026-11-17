package auth_test

import (
	"testing"

	"github.com/sakid00/enmasse-be/internal/auth"
)

func TestRandomOTPIsSixDigits(t *testing.T) {
	t.Parallel()
	seen := map[string]struct{}{}
	for range 20 {
		code, err := auth.RandomOTP()
		if err != nil {
			t.Fatal(err)
		}
		if len(code) != auth.OTPLength {
			t.Fatalf("len = %d code=%q", len(code), code)
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				t.Fatalf("non-digit %q", code)
			}
		}
		seen[code] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatal("expected more than one distinct OTP across 20 draws")
	}
}

package domain

import "testing"

func TestParseOTPPurpose(t *testing.T) {
	t.Parallel()
	for _, purpose := range []string{OTPSignup, OTPClaim, OTPReset, OTPVerify} {
		got, err := ParseOTPPurpose(purpose)
		if err != nil || got != purpose {
			t.Fatalf("%s: got %q err=%v", purpose, got, err)
		}
	}
	if _, err := ParseOTPPurpose("login"); err != ErrOTPPurpose {
		t.Fatalf("got %v", err)
	}
}

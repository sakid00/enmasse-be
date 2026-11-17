package domain

const (
	OTPSignup = "signup"
	OTPClaim  = "claim"
	OTPReset  = "reset"
	OTPVerify = "verify"
)

func ParseOTPPurpose(raw string) (string, error) {
	switch raw {
	case OTPSignup, OTPClaim, OTPReset, OTPVerify:
		return raw, nil
	default:
		return "", ErrOTPPurpose
	}
}

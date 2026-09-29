package config

import "testing"

func TestProductionRejectsDummySecrets(t *testing.T) {
	t.Parallel()
	c := Config{
		Env:              "production",
		JWTAccessSecret:  "abcdefghijklmnopqrstuvwxyz012345",
		JWTServiceSecret: "abcdefghijklmnopqrstuvwxyz678901",
		TurnstileSecret:  DummyTurnstileSecret,
	}
	if err := c.validate(); err == nil {
		t.Fatal("dummy turnstile must fail in production")
	}

	c.TurnstileSecret = "real-turnstile-secret-from-cloudflare"
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}

	dev := c
	dev.Env = "development"
	dev.JWTAccessSecret = "change-me"
	if err := dev.validate(); err != nil {
		t.Fatal("development should skip production checks")
	}
}

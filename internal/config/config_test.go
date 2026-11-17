package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	c.TurnstileSecret = "0x4AAAAAAshort-site-key"
	c.MailDriver = "resend"
	c.ResendAPIKey = "re_test"
	c.ResendOTPTemplateID = "tmpl_otp"
	c.PasswordWrapPrivateKeyFile = "secrets/password-wrap.pem"
	if err := c.validate(); err == nil {
		t.Fatal("widget site key used as secret must fail in production")
	}

	c.TurnstileSecret = "real-turnstile-secret-from-cloudflare"
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}

	c.PasswordWrapPrivateKeyFile = ""
	if err := c.validate(); err == nil {
		t.Fatal("production must require a wrap key file or PEM")
	}
	c.PasswordWrapPrivateKeyFile = "secrets/password-wrap.pem"
	c.MailDriver = "log"
	if err := c.validate(); err == nil {
		t.Fatal("production must require resend")
	}

	dev := c
	dev.Env = "development"
	dev.JWTAccessSecret = "change-me"
	if err := dev.validate(); err != nil {
		t.Fatal("development should skip production checks")
	}
}

func TestWrapPrivateKeyPEMFromFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "wrap.pem")
	if err := os.WriteFile(path, []byte("-----BEGIN PRIVATE KEY-----\nlocal\n-----END PRIVATE KEY-----\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := Config{PasswordWrapPrivateKeyFile: path, PasswordWrapPrivateKey: "ignored"}
	got, err := c.WrapPrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "BEGIN PRIVATE KEY") {
		t.Fatalf("got %q", got)
	}
}

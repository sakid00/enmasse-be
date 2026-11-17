package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const DummyTurnstileSecret = "1x0000000000000000000000000000000AA"

type Config struct {
	Port string `env:"PORT" envDefault:"8081"`
	Env  string `env:"APP_ENV" envDefault:"development"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTAccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
	JWTIssuer        string        `env:"JWT_ISSUER" envDefault:"https://api.enmasse.id"`
	JWTAccessExpiry  time.Duration `env:"JWT_ACCESS_EXPIRY" envDefault:"8h"`
	JWTServiceSecret string        `env:"JWT_SERVICE_SECRET,required"`

	TurnstileSecret         string `env:"TURNSTILE_SECRET_KEY" envDefault:"1x0000000000000000000000000000000AA"`
	TurnstileExpectedAction string `env:"TURNSTILE_EXPECTED_ACTION" envDefault:"login"`

	PasswordWrapPrivateKey     string `env:"PASSWORD_WRAP_PRIVATE_KEY"`
	PasswordWrapPrivateKeyFile string `env:"PASSWORD_WRAP_PRIVATE_KEY_FILE"`

	PublicAppURL  string        `env:"PUBLIC_APP_URL" envDefault:"http://localhost:3000"`
	ClaimTokenTTL time.Duration `env:"CLAIM_TOKEN_TTL" envDefault:"48h"`

	MailDriver          string        `env:"MAIL_DRIVER" envDefault:"log"`
	ResendAPIKey        string        `env:"RESEND_API_KEY"`
	MailFrom            string        `env:"MAIL_FROM" envDefault:"MASSmaker <hello@enmasse.id>"`
	ResendOTPTemplateID string        `env:"RESEND_OTP_TEMPLATE_ID"`
	OTPTTL              time.Duration `env:"OTP_TTL" envDefault:"10m"`
	OTPMaxAttempts      int           `env:"OTP_MAX_ATTEMPTS" envDefault:"5"`
	OTPResendAfter      time.Duration `env:"OTP_RESEND_AFTER" envDefault:"60s"`

	SMTPHost string `env:"SMTP_HOST"`
	SMTPPort int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser string `env:"SMTP_USER"`
	SMTPPass string `env:"SMTP_PASS"`
	SMTPFrom string `env:"SMTP_FROM" envDefault:"hello@enmasse.id"`

	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000"`
}

func Load() (*Config, error) {
	loadDotEnv()
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w (copy .env.example to .env in the repo root, then run from that directory)", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if !c.IsProduction() {
		return nil
	}
	if len(c.JWTAccessSecret) < 32 || len(c.JWTServiceSecret) < 32 {
		return fmt.Errorf("config: JWT_ACCESS_SECRET and JWT_SERVICE_SECRET must be at least 32 characters in production")
	}
	if strings.Contains(c.JWTAccessSecret, "change-me") || strings.Contains(c.JWTServiceSecret, "change-me") {
		return fmt.Errorf("config: replace example JWT secrets before production")
	}
	if c.TurnstileSecret == DummyTurnstileSecret {
		return fmt.Errorf("config: replace dummy TURNSTILE_SECRET_KEY before production")
	}
	if c.TurnstileSecretLooksLikeSiteKey() {
		return fmt.Errorf("config: TURNSTILE_SECRET_KEY looks like a widget site key; use the Siteverify secret from the Cloudflare dashboard")
	}
	if strings.TrimSpace(c.PasswordWrapPrivateKey) == "" && strings.TrimSpace(c.PasswordWrapPrivateKeyFile) == "" {
		return fmt.Errorf("config: PASSWORD_WRAP_PRIVATE_KEY_FILE or PASSWORD_WRAP_PRIVATE_KEY is required in production")
	}
	if strings.ToLower(strings.TrimSpace(c.MailDriver)) != "resend" {
		return fmt.Errorf("config: MAIL_DRIVER=resend is required in production")
	}
	if strings.TrimSpace(c.ResendAPIKey) == "" {
		return fmt.Errorf("config: RESEND_API_KEY is required in production")
	}
	if strings.TrimSpace(c.ResendOTPTemplateID) == "" {
		return fmt.Errorf("config: RESEND_OTP_TEMPLATE_ID is required in production")
	}
	return nil
}

func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDummyTurnstile() bool {
	return c.TurnstileSecret == DummyTurnstileSecret
}

func (c *Config) WrapPrivateKeyPEM() (string, error) {
	file := strings.TrimSpace(c.PasswordWrapPrivateKeyFile)
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("config: PASSWORD_WRAP_PRIVATE_KEY_FILE: %w", err)
		}
		return string(b), nil
	}
	return c.PasswordWrapPrivateKey, nil
}

// TurnstileSecretLooksLikeSiteKey is true when the configured secret is too
// short to be a Siteverify key (widget site keys are typically ~26 chars).
func (c *Config) TurnstileSecretLooksLikeSiteKey() bool {
	s := strings.TrimSpace(c.TurnstileSecret)
	if s == "" || c.IsDummyTurnstile() {
		return false
	}
	return len(s) < 32
}

package mail

import (
	"context"
	"fmt"
	"strings"

	"github.com/sakid00/enmasse-be/internal/config"
)

type Mailer interface {
	SendClaim(ctx context.Context, to, setPasswordURL string) error
	SendOTP(ctx context.Context, to, code string) error
}

func New(cfg *config.Config) (Mailer, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.MailDriver))
	if driver == "" && strings.TrimSpace(cfg.SMTPHost) != "" {
		driver = "smtp"
	}
	switch driver {
	case "", "log":
		return Log{}, nil
	case "smtp":
		from := strings.TrimSpace(cfg.SMTPFrom)
		if from == "" {
			from = cfg.MailFrom
		}
		return SMTP{
			host: cfg.SMTPHost,
			port: cfg.SMTPPort,
			user: cfg.SMTPUser,
			pass: cfg.SMTPPass,
			from: from,
			ttl:  cfg.OTPTTL,
		}, nil
	case "resend":
		if strings.TrimSpace(cfg.ResendAPIKey) == "" {
			return nil, fmt.Errorf("mail: RESEND_API_KEY is required")
		}
		from := strings.TrimSpace(cfg.MailFrom)
		if from == "" {
			from = "MASSmaker <hello@enmasse.id>"
		}
		return NewResendMailer(cfg.ResendAPIKey, from, cfg.ResendOTPTemplateID, cfg.OTPTTL), nil
	default:
		return nil, fmt.Errorf("mail: unknown MAIL_DRIVER %q", driver)
	}
}

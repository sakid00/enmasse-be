package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/sakid00/enmasse-be/internal/config"
)

type Sender interface {
	SendClaim(ctx context.Context, to, setPasswordURL string) error
}

type SMTP struct {
	host string
	port int
	user string
	pass string
	from string
}

func New(cfg *config.Config) Sender {
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return Log{}
	}
	return SMTP{
		host: cfg.SMTPHost,
		port: cfg.SMTPPort,
		user: cfg.SMTPUser,
		pass: cfg.SMTPPass,
		from: cfg.SMTPFrom,
	}
}

func (s SMTP) SendClaim(_ context.Context, to, setPasswordURL string) error {
	subject := "Set your MASSmaker password"
	body := "Set your password (link expires soon):\r\n\r\n" + setPasswordURL + "\r\n"
	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("mail: send claim: %w", err)
	}
	return nil
}

type Log struct{}

func (Log) SendClaim(_ context.Context, to, setPasswordURL string) error {
	slog.Info("claim mail (smtp unset — logged only)", "to", to, "url", setPasswordURL)
	return nil
}

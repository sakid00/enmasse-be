package mail

import (
	"context"
	"log/slog"
)

type Log struct{}

func (Log) SendClaim(_ context.Context, to, setPasswordURL string) error {
	slog.Info("claim mail (MAIL_DRIVER=log)", "to", to, "url", setPasswordURL)
	return nil
}

func (Log) SendOTP(_ context.Context, to, code string) error {
	slog.Info("otp not emailed (MAIL_DRIVER=log)", "to", to, "code", code)
	return nil
}

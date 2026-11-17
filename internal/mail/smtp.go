package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"time"
)

type SMTP struct {
	host string
	port int
	user string
	pass string
	from string
	ttl  time.Duration
}

func (s SMTP) SendClaim(_ context.Context, to, setPasswordURL string) error {
	subject := "Set your MASSmaker password"
	body := "Set your password (link expires soon):\r\n\r\n" + setPasswordURL + "\r\n"
	return s.send(to, subject, body)
}

func (s SMTP) SendOTP(_ context.Context, to, code string) error {
	minutes := int(s.ttl.Minutes())
	if minutes < 1 {
		minutes = 10
	}
	subject := "Your MASSmaker verification code"
	body := fmt.Sprintf("Your verification code is %s.\r\nIt expires in %d minutes.\r\n", code, minutes)
	return s.send(to, subject, body)
}

func (s SMTP) send(to, subject, body string) error {
	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("mail: smtp send: %w", err)
	}
	return nil
}

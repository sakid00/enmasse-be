package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const resendEmailsURL = "https://api.resend.com/emails"

type ResendMailer struct {
	apiKey     string
	from       string
	templateID string
	ttl        time.Duration
	client     *http.Client
	endpoint   string
}

func NewResendMailer(apiKey, from, templateID string, ttl time.Duration) *ResendMailer {
	return &ResendMailer{
		apiKey:     apiKey,
		from:       from,
		templateID: templateID,
		ttl:        ttl,
		client:     &http.Client{Timeout: 15 * time.Second},
		endpoint:   resendEmailsURL,
	}
}

type resendEmailRequest struct {
	From     string          `json:"from"`
	To       []string        `json:"to"`
	Subject  string          `json:"subject,omitempty"`
	Text     string          `json:"text,omitempty"`
	Template *resendTemplate `json:"template,omitempty"`
}

type resendTemplate struct {
	ID        string         `json:"id"`
	Variables map[string]any `json:"variables"`
}

func (m *ResendMailer) SendOTP(ctx context.Context, to, code string) error {
	if strings.TrimSpace(m.templateID) == "" {
		return fmt.Errorf("mail: RESEND_OTP_TEMPLATE_ID is required")
	}
	err := m.send(ctx, resendEmailRequest{
		From: m.from,
		To:   []string{to},
		Template: &resendTemplate{
			ID: m.templateID,
			Variables: map[string]any{
				"code":       code,
				"TTLMinutes": ttlMinutes(m.ttl),
			},
		},
	})
	if err != nil {
		slog.Error("resend: send otp failed", "to", to, "from", m.from, "err", err)
		return err
	}
	return nil
}

func (m *ResendMailer) SendClaim(ctx context.Context, to, setPasswordURL string) error {
	err := m.send(ctx, resendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: "Set your MASSmaker password",
		Text:    "Set your password (link expires soon):\n\n" + setPasswordURL + "\n",
	})
	if err != nil {
		slog.Error("resend: send claim failed", "to", to, "from", m.from, "err", err)
		return err
	}
	return nil
}

func (m *ResendMailer) send(ctx context.Context, payload resendEmailRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("resend: encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: send: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if resp.StatusCode >= 300 {
		detail := strings.TrimSpace(string(raw))
		if detail == "" {
			return fmt.Errorf("resend: status %d", resp.StatusCode)
		}
		return fmt.Errorf("resend: status %d: %s", resp.StatusCode, detail)
	}
	return nil
}

func ttlMinutes(ttl time.Duration) int {
	minutes := int(ttl.Minutes())
	if minutes < 1 {
		return 10
	}
	return minutes
}

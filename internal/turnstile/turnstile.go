package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
)

const SiteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type Verifier struct {
	secret         string
	expectedAction string
	dummy          bool
	client         *http.Client
	url            string
}

func New(cfg *config.Config) *Verifier {
	return &Verifier{
		secret:         cfg.TurnstileSecret,
		expectedAction: cfg.TurnstileExpectedAction,
		dummy:          cfg.IsDummyTurnstile(),
		client:         &http.Client{Timeout: 5 * time.Second},
		url:            SiteverifyURL,
	}
}

func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrTurnstileRequired
	}
	if v.dummy {
		return nil
	}

	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)
	if ip := sanitizeRemoteIP(remoteIP); ip != "" {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.url, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("turnstile: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("turnstile: %w", err)
	}
	defer res.Body.Close()

	var body struct {
		Success bool     `json:"success"`
		Action  string   `json:"action"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return fmt.Errorf("turnstile: decode: %w", err)
	}
	if !body.Success {
		slog.Warn("turnstile siteverify failed", "error-codes", body.Errors, "action", body.Action)
		return domain.ErrTurnstileFailed
	}
	if v.expectedAction != "" && body.Action != "" && body.Action != v.expectedAction {
		return domain.ErrTurnstileFailed
	}
	return nil
}

func NewForTest(cfg *config.Config, siteverifyURL string) *Verifier {
	v := New(cfg)
	v.url = siteverifyURL
	return v
}

func sanitizeRemoteIP(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return ""
	}
	return ip.String()
}

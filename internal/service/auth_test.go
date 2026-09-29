package service_test

import (
	"context"
	"testing"

	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/mail"
	"github.com/sakid00/enmasse-be/internal/service"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

func TestLoginRequiresTurnstile(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{TurnstileSecret: config.DummyTurnstileSecret}
	svc := service.NewAuthService(nil, cfg, turnstile.New(cfg), mail.Log{})
	_, err := svc.Login(context.Background(), "artist@example.com", "password1", "", "")
	if err != domain.ErrTurnstileRequired {
		t.Fatalf("got %v", err)
	}
}

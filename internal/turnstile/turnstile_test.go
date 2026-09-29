package turnstile_test

import (
	"context"
	"testing"

	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

func TestDummyAcceptsAnyToken(t *testing.T) {
	t.Parallel()
	v := turnstile.New(&config.Config{TurnstileSecret: config.DummyTurnstileSecret})
	if err := v.Verify(context.Background(), "any-token", ""); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyTokenRejected(t *testing.T) {
	t.Parallel()
	v := turnstile.New(&config.Config{TurnstileSecret: config.DummyTurnstileSecret})
	err := v.Verify(context.Background(), "  ", "")
	if err != domain.ErrTurnstileRequired {
		t.Fatalf("got %v", err)
	}
}

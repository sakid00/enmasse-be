package turnstile_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

func TestDummyAcceptsAnyToken(t *testing.T) {
	t.Parallel()
	v := turnstile.New(&config.Config{TurnstileSecret: config.DummyTurnstileSecret})
	if err := v.Verify(context.Background(), "any-token", "127.0.0.1:54321"); err != nil {
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

func TestSiteverifyStripsPortAndReportsFailure(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form := string(body)
		if !strings.Contains(form, "remoteip=127.0.0.1") || strings.Contains(form, "54321") {
			t.Errorf("remoteip not sanitized: %s", form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":     false,
			"error-codes": []string{"invalid-input-secret"},
		})
	}))
	defer srv.Close()

	v := turnstile.NewForTest(&config.Config{TurnstileSecret: "real-secret-not-dummy"}, srv.URL)
	err := v.Verify(context.Background(), "token", "127.0.0.1:54321")
	if err != domain.ErrTurnstileFailed {
		t.Fatalf("got %v", err)
	}
}

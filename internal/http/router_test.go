package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAlwaysMounted(t *testing.T) {
	t.Parallel()
	h := NewRouter(nil, RouterOptions{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health: got %d", rec.Code)
	}
}

func TestInternalRequiresAuth(t *testing.T) {
	t.Parallel()
	h := NewRouter(nil, RouterOptions{JWTServiceSecret: "test-service-secret-32-bytes-min"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/internal/profile-status/00000000-0000-0000-0000-000000000001", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("internal without jwt: got %d", rec.Code)
	}
}

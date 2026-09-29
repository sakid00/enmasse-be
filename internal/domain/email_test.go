package domain_test

import (
	"testing"

	"github.com/sakid00/enmasse-be/internal/domain"
)

func TestCanonicalEmail(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"  User@Gmail.com ", "user@gmail.com"},
		{"j.o.h.n+vote2@gmail.com", "john@gmail.com"},
		{"john+tag@googlemail.com", "john@gmail.com"},
		{"john.doe@example.com", "john.doe@example.com"},
		{"a.b+c@Example.COM", "a.b+c@example.com"},
	}
	for _, tc := range cases {
		got := domain.CanonicalEmail(tc.in)
		if got != tc.want {
			t.Errorf("CanonicalEmail(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLooksLikeEmail(t *testing.T) {
	t.Parallel()
	if !domain.LooksLikeEmail("a@b.co") {
		t.Fatal("expected valid")
	}
	if domain.LooksLikeEmail("not-an-email") {
		t.Fatal("expected invalid")
	}
}

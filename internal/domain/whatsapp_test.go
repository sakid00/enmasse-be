package domain_test

import (
	"testing"

	"github.com/sakid00/enmasse-be/internal/domain"
)

func TestNormalizeWhatsApp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
		err      bool
	}{
		{"", "", false},
		{"081234567890", "+6281234567890", false},
		{"81234567890", "+6281234567890", false},
		{"+62 812-3456-7890", "+6281234567890", false},
		{"6281234567890", "+6281234567890", false},
		{"+15551234567", "", true},
		{"123", "", true},
	}
	for _, tc := range cases {
		got, err := domain.NormalizeWhatsApp(tc.in)
		if tc.err {
			if err == nil {
				t.Fatalf("%q: expected error", tc.in)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q %v, want %q", tc.in, got, err, tc.want)
		}
	}
}

package auth_test

import (
	"testing"

	"github.com/sakid00/enmasse-be/internal/auth"
)

func TestHashVerify(t *testing.T) {
	t.Parallel()
	hash, err := auth.Hash("secret-pass")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.Verify(hash, "secret-pass") {
		t.Fatal("expected match")
	}
	if auth.Verify(hash, "wrong") {
		t.Fatal("expected mismatch")
	}
}

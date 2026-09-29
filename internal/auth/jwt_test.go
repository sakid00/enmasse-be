package auth_test

import (
	"testing"
	"time"

	"github.com/sakid00/enmasse-be/internal/auth"
)

func TestAccessRoundTrip(t *testing.T) {
	t.Parallel()
	const secret = "test-access-secret-32-bytes-min!!"
	const issuer = "https://api.enmasse.id"
	tok, err := auth.MintAccess(secret, issuer, "user-1", "artist", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := auth.ParseAccess(secret, issuer, tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "user-1" || got.Role != "artist" || got.Typ != auth.TypAccess {
		t.Fatalf("claims %+v", got)
	}
}

func TestServiceRejectsAccess(t *testing.T) {
	t.Parallel()
	const secret = "test-service-secret-32-bytes-min"
	const issuer = "https://api.enmasse.id"
	tok, err := auth.MintAccess(secret, issuer, "user-1", "artist", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ParseService(secret, tok); err == nil {
		t.Fatal("human access token must not pass service middleware")
	}
}

func TestServiceRoundTrip(t *testing.T) {
	t.Parallel()
	const secret = "test-service-secret-32-bytes-min"
	tok, err := auth.MintService(secret, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	got, err := auth.ParseService(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.Typ != auth.TypService || got.Issuer != auth.IssMassmaker {
		t.Fatalf("claims %+v", got)
	}
}

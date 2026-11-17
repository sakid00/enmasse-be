package passwrap_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/sakid00/enmasse-be/internal/passwrap"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	box, err := passwrap.New("", true)
	if err != nil {
		t.Fatal(err)
	}
	if !box.Generated() {
		t.Fatal("expected generated key")
	}
	jwk := box.PublicJWK()
	if jwk.Kty != "RSA" || jwk.Alg != "RSA-OAEP-256" || jwk.N == "" || jwk.E == "" {
		t.Fatalf("jwk %+v", jwk)
	}

	plain := "secret-pass"
	ct, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicFromJWK(t, jwk), []byte(plain), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := box.Unwrap(base64.StdEncoding.EncodeToString(ct))
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("got %q", got)
	}
}

func TestRejectsGarbage(t *testing.T) {
	t.Parallel()
	box, err := passwrap.New("", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := box.Unwrap("not-base64!!"); err == nil {
		t.Fatal("expected error")
	}
}

func TestProductionRequiresPEM(t *testing.T) {
	t.Parallel()
	if _, err := passwrap.New("", false); err == nil {
		t.Fatal("expected error")
	}
}

func publicFromJWK(t *testing.T, jwk passwrap.PublicJWK) *rsa.PublicKey {
	t.Helper()
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		t.Fatal(err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		t.Fatal(err)
	}
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}
}

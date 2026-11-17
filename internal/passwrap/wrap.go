package passwrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
)

const rsaBits = 2048

type Box struct {
	priv      *rsa.PrivateKey
	generated bool
}

type PublicJWK struct {
	Kty    string   `json:"kty"`
	N      string   `json:"n"`
	E      string   `json:"e"`
	Alg    string   `json:"alg"`
	Ext    bool     `json:"ext"`
	KeyOps []string `json:"key_ops"`
}

func New(pemKey string, allowGenerate bool) (*Box, error) {
	pemKey = strings.TrimSpace(pemKey)
	if pemKey != "" {
		priv, err := parsePrivateKey(pemKey)
		if err != nil {
			return nil, err
		}
		return &Box{priv: priv}, nil
	}
	if !allowGenerate {
		return nil, fmt.Errorf("passwrap: wrap private key is required")
	}
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return nil, fmt.Errorf("passwrap: generate: %w", err)
	}
	return &Box{priv: priv, generated: true}, nil
}

func (b *Box) Generated() bool {
	return b != nil && b.generated
}

func (b *Box) PublicJWK() PublicJWK {
	n := b.priv.N.Bytes()
	e := big.NewInt(int64(b.priv.E)).Bytes()
	return PublicJWK{
		Kty:    "RSA",
		N:      base64.RawURLEncoding.EncodeToString(n),
		E:      base64.RawURLEncoding.EncodeToString(e),
		Alg:    "RSA-OAEP-256",
		Ext:    true,
		KeyOps: []string{"encrypt"},
	}
}

func (b *Box) Unwrap(cipher string) (string, error) {
	raw, err := decodeCipher(strings.TrimSpace(cipher))
	if err != nil {
		return "", err
	}
	pt, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, b.priv, raw, nil)
	if err != nil {
		return "", fmt.Errorf("passwrap: decrypt: %w", err)
	}
	return string(pt), nil
}

func parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.ReplaceAll(pemKey, `\n`, "\n")))
	if block == nil {
		return nil, fmt.Errorf("passwrap: invalid PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("passwrap: not an RSA key")
		}
		return rsaKey, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("passwrap: unsupported private key")
}

func decodeCipher(s string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var last error
	for _, enc := range encodings {
		raw, err := enc.DecodeString(s)
		if err == nil {
			return raw, nil
		}
		last = err
	}
	return nil, fmt.Errorf("passwrap: cipher: %w", last)
}

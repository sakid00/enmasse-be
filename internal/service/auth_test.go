package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/sakid00/enmasse-be/internal/auth"
	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/mail"
	"github.com/sakid00/enmasse-be/internal/service"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

func TestLoginRequiresTurnstile(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{TurnstileSecret: config.DummyTurnstileSecret}
	svc := service.NewAuthService(nil, cfg, turnstile.New(cfg), mail.Log{}, nil)
	_, err := svc.Login(context.Background(), "artist@example.com", "password1", "", "")
	if err != domain.ErrTurnstileRequired {
		t.Fatalf("got %v", err)
	}
}

func TestForgotPasswordInvalidEmailIsOK(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, turnstile.New(&config.Config{}), mail.Log{}, nil)
	if err := svc.ForgotPassword(context.Background(), "not-an-email"); err != nil {
		t.Fatal(err)
	}
}

func TestResendOTPRejectsVerifyPurpose(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.ResendOTP(context.Background(), "artist@example.com", domain.OTPVerify)
	if err != domain.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestResendOTPClaimAlways200ForBadEmail(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	out, err := svc.ResendOTP(context.Background(), "not-an-email", domain.OTPClaim)
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatal("expected ok")
	}
}

func TestClaimPasswordRequiresOTPCode(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.ClaimPassword(context.Background(), "not-an-email", "password1", "")
	if err != domain.ErrInvalidEmail {
		t.Fatalf("got %v", err)
	}
}

func TestUnwrapPasswordPlaintextPassthrough(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	got, err := svc.UnwrapPassword("secret-pass", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret-pass" {
		t.Fatalf("got %q", got)
	}
	if _, err := svc.UnwrapPassword("", "AAAA"); err != domain.ErrPasswordPayload {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterArtistRequiresSignupToken(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.RegisterArtist(context.Background(), service.RegisterArtistInput{
		Password:   "password1",
		Handle:     "artist1",
		FirstName:  "Ada",
		LastName:   "Lovelace",
		ArtistName: "Ada",
	})
	if err != domain.ErrInvalidToken {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterVendorRequiresSignupToken(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.RegisterVendor(context.Background(), service.RegisterVendorInput{
		Password: "password1",
		Handle:   "vendor1",
		Name:     "Studio",
	})
	if err != domain.ErrInvalidToken {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterArtistSignupTokenThenWeakPassword(t *testing.T) {
	t.Parallel()
	const secret = "test-access-secret-32-bytes-min!!"
	const issuer = "https://api.enmasse.id"
	cfg := &config.Config{JWTAccessSecret: secret, JWTIssuer: issuer}
	svc := service.NewAuthService(nil, cfg, nil, mail.Log{}, nil)
	tok, err := auth.MintSignup(
		secret,
		issuer,
		"artist@example.com",
		30*time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RegisterArtist(context.Background(), service.RegisterArtistInput{
		SignupToken: tok,
		Password:    "short",
		Handle:      "artist1",
		FirstName:   "Ada",
		LastName:    "Lovelace",
		ArtistName:  "Ada",
	})
	if err != domain.ErrWeakPassword {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterArtistRejectsAccessToken(t *testing.T) {
	t.Parallel()
	const secret = "test-access-secret-32-bytes-min!!"
	const issuer = "https://api.enmasse.id"
	cfg := &config.Config{JWTAccessSecret: secret, JWTIssuer: issuer}
	svc := service.NewAuthService(nil, cfg, nil, mail.Log{}, nil)
	tok, err := auth.MintAccess(
		secret,
		issuer,
		"user-1",
		"artist",
		time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RegisterArtist(context.Background(), service.RegisterArtistInput{
		SignupToken: tok,
		Password:    "password1",
		Handle:      "artist1",
		FirstName:   "Ada",
		LastName:    "Lovelace",
		ArtistName:  "Ada",
	})
	if err != domain.ErrInvalidToken {
		t.Fatalf("got %v", err)
	}
}

func TestVerifySignupOTPInvalidEmail(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.VerifySignupOTP(context.Background(), "not-an-email", "123456")
	if err != domain.ErrInvalidEmail {
		t.Fatalf("got %v", err)
	}
}

func TestResendOTPSignupInvalidEmail(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(nil, &config.Config{}, nil, mail.Log{}, nil)
	_, err := svc.ResendOTP(context.Background(), "not-an-email", domain.OTPSignup)
	if err != domain.ErrInvalidEmail {
		t.Fatalf("got %v", err)
	}
}

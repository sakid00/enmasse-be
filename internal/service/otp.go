package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/sakid00/enmasse-be/internal/auth"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/store"
)

const (
	StatusOTPSent         = "otpSent"
	StatusAlreadyVerified = "alreadyVerified"
	defaultOTPTTL         = 10 * time.Minute
	defaultOTPAttempts    = 5
	defaultOTPResendAfter = 60 * time.Second
	defaultSignupTokenTTL = 30 * time.Minute
)

type OTPSent struct {
	Status string `json:"status"`
}

type OTPResendResult struct {
	OK     bool   `json:"ok"`
	Status string `json:"status,omitempty"`
}

func (s *AuthService) ResendOTP(ctx context.Context, rawEmail, purpose string) (*OTPResendResult, error) {
	parsed, err := domain.ParseOTPPurpose(purpose)
	if err != nil {
		return nil, err
	}
	if parsed == domain.OTPVerify {
		return nil, domain.ErrUnauthorized
	}
	if !domain.LooksLikeEmail(rawEmail) {
		if parsed == domain.OTPClaim || parsed == domain.OTPReset {
			return &OTPResendResult{OK: true}, nil
		}
		return nil, domain.ErrInvalidEmail
	}
	email := domain.CanonicalEmail(rawEmail)

	switch parsed {
	case domain.OTPClaim, domain.OTPReset:
		if err := s.sendPublicOTP(ctx, email, parsed); err != nil {
			slog.Error("otp resend", "purpose", parsed, "err", err)
		}
		return &OTPResendResult{OK: true}, nil
	case domain.OTPSignup:
		user, err := s.db.Queries.GetUserByEmail(ctx, email)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("otp resend: %w", err)
		}
		if err == nil {
			if user.PasswordHash == nil || isEmailVerified(user) {
				return nil, domain.ErrNotFound
			}
		}
		if err := s.issueOTP(ctx, email, domain.OTPSignup); err != nil {
			return nil, err
		}
		return &OTPResendResult{OK: true, Status: StatusOTPSent}, nil
	default:
		return nil, domain.ErrOTPPurpose
	}
}

type SignupProof struct {
	SignupToken string `json:"signupToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

type VerifySignupOTPResult struct {
	Session *Session
	Proof   *SignupProof
}

func (s *AuthService) VerifySignupOTP(ctx context.Context, rawEmail, code string) (*VerifySignupOTPResult, error) {
	if !domain.LooksLikeEmail(rawEmail) {
		return nil, domain.ErrInvalidEmail
	}
	email := domain.CanonicalEmail(rawEmail)
	user, err := s.db.Queries.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("verify signup otp: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		if err := s.consumeOTP(ctx, email, domain.OTPSignup, code); err != nil {
			return nil, err
		}
		proof, err := s.mintSignupProof(email)
		if err != nil {
			return nil, err
		}
		return &VerifySignupOTPResult{Proof: proof}, nil
	}
	if user.PasswordHash == nil {
		return nil, domain.ErrInvalidOTP
	}
	if err := s.consumeOTP(ctx, email, domain.OTPSignup, code); err != nil {
		return nil, err
	}
	if err := s.db.Queries.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("verify signup otp: mark: %w", err)
	}
	user.EmailVerifiedAt = verifiedNow()
	complete, err := s.profileComplete(ctx, user)
	if err != nil {
		return nil, err
	}
	sess, err := s.session(user, complete)
	if err != nil {
		return nil, err
	}
	return &VerifySignupOTPResult{Session: sess}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, rawEmail string) error {
	if !domain.LooksLikeEmail(rawEmail) {
		return nil
	}
	if err := s.sendPublicOTP(ctx, domain.CanonicalEmail(rawEmail), domain.OTPReset); err != nil {
		slog.Error("forgot password otp", "err", err)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, rawEmail, code, password string) error {
	if !domain.LooksLikeEmail(rawEmail) {
		return domain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	email := domain.CanonicalEmail(rawEmail)
	user, err := s.db.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInvalidOTP
		}
		return fmt.Errorf("reset password: %w", err)
	}
	if user.PasswordHash == nil {
		return domain.ErrInvalidOTP
	}
	if err := s.consumeOTP(ctx, email, domain.OTPReset, code); err != nil {
		return err
	}
	hash, err := auth.Hash(password)
	if err != nil {
		return fmt.Errorf("reset password: hash: %w", err)
	}
	if err := s.db.Queries.SetPassword(ctx, store.SetPasswordParams{
		ID:                 user.ID,
		PasswordHash:       &hash,
		MustChangePassword: false,
	}); err != nil {
		return fmt.Errorf("reset password: update: %w", err)
	}
	if err := s.db.Queries.MarkEmailVerified(ctx, user.ID); err != nil {
		return fmt.Errorf("reset password: verify: %w", err)
	}
	return nil
}

func (s *AuthService) SendVerifyEmail(ctx context.Context, userID uuid.UUID) (*OTPSent, error) {
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("send verify email: %w", err)
	}
	if isEmailVerified(user) {
		return &OTPSent{Status: StatusAlreadyVerified}, nil
	}
	if err := s.issueOTP(ctx, user.Email, domain.OTPVerify); err != nil {
		return nil, err
	}
	return &OTPSent{Status: StatusOTPSent}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("verify email: %w", err)
	}
	if isEmailVerified(user) {
		return nil
	}
	if err := s.consumeOTP(ctx, user.Email, domain.OTPVerify, code); err != nil {
		return err
	}
	if err := s.db.Queries.MarkEmailVerified(ctx, user.ID); err != nil {
		return fmt.Errorf("verify email: mark: %w", err)
	}
	return nil
}

func (s *AuthService) sendPublicOTP(ctx context.Context, email, purpose string) error {
	user, err := s.db.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	switch purpose {
	case domain.OTPClaim:
		if user.PasswordHash != nil {
			return nil
		}
	case domain.OTPReset:
		if user.PasswordHash == nil {
			return nil
		}
	default:
		return domain.ErrOTPPurpose
	}
	if err := s.issueOTP(ctx, email, purpose); err != nil {
		if errors.Is(err, domain.ErrOTPResendWait) {
			return nil
		}
		return err
	}
	return nil
}

func (s *AuthService) issueOTP(ctx context.Context, email, purpose string) error {
	if err := s.enforceOTPResendCooldown(ctx, email, purpose); err != nil {
		return err
	}
	code, err := auth.RandomOTP()
	if err != nil {
		return fmt.Errorf("issue otp: %w", err)
	}
	hash, err := auth.Hash(code)
	if err != nil {
		return fmt.Errorf("issue otp: hash: %w", err)
	}
	if _, err := s.db.Queries.CreateEmailOTP(ctx, store.CreateEmailOTPParams{
		EmailCanonical: email,
		Purpose:        purpose,
		CodeHash:       hash,
		ExpiresAt:      time.Now().Add(s.otpTTL()),
	}); err != nil {
		return fmt.Errorf("issue otp: store: %w", err)
	}
	if s.mail == nil {
		return fmt.Errorf("issue otp: mailer unset")
	}
	if err := s.mail.SendOTP(ctx, email, code); err != nil {
		return fmt.Errorf("issue otp: send: %w", err)
	}
	return nil
}

func (s *AuthService) consumeOTP(ctx context.Context, email, purpose, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.ErrOTPRequired
	}
	row, err := s.db.Queries.GetLatestEmailOTP(ctx, store.GetLatestEmailOTPParams{
		EmailCanonical: email,
		Purpose:        purpose,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInvalidOTP
		}
		return fmt.Errorf("consume otp: %w", err)
	}
	if row.ConsumedAt.Valid {
		return domain.ErrInvalidOTP
	}
	if !time.Now().Before(row.ExpiresAt) {
		return domain.ErrInvalidOTP
	}
	if int(row.AttemptCount) >= s.otpMaxAttempts() {
		return domain.ErrOTPRateLimited
	}
	if !auth.Verify(row.CodeHash, code) {
		if _, incErr := s.db.Queries.IncrementEmailOTPAttempts(ctx, row.ID); incErr != nil {
			return fmt.Errorf("consume otp: attempts: %w", incErr)
		}
		return domain.ErrInvalidOTP
	}
	if err := s.db.Queries.ConsumeEmailOTP(ctx, row.ID); err != nil {
		return fmt.Errorf("consume otp: mark: %w", err)
	}
	return nil
}

func (s *AuthService) enforceOTPResendCooldown(ctx context.Context, email, purpose string) error {
	row, err := s.db.Queries.GetLatestEmailOTP(ctx, store.GetLatestEmailOTPParams{
		EmailCanonical: email,
		Purpose:        purpose,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("otp cooldown: %w", err)
	}
	if time.Since(row.CreatedAt) < s.otpResendAfter() {
		return domain.ErrOTPResendWait
	}
	return nil
}

func (s *AuthService) requireVerified(user store.User) error {
	if !isEmailVerified(user) {
		return domain.ErrEmailUnverified
	}
	return nil
}

func (s *AuthService) mintSignupProof(email string) (*SignupProof, error) {
	if s.cfg == nil {
		return nil, domain.ErrUnavailable
	}
	ttl := defaultSignupTokenTTL
	tok, err := auth.MintSignup(
		s.cfg.JWTAccessSecret,
		s.cfg.JWTIssuer,
		email,
		ttl,
	)
	if err != nil {
		return nil, fmt.Errorf("mint signup: %w", err)
	}
	return &SignupProof{
		SignupToken: tok,
		ExpiresIn:   int(ttl.Seconds()),
	}, nil
}

func (s *AuthService) emailFromSignupToken(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", domain.ErrInvalidToken
	}
	if s.cfg == nil {
		return "", domain.ErrUnavailable
	}
	claims, err := auth.ParseSignup(s.cfg.JWTAccessSecret, s.cfg.JWTIssuer, token)
	if err != nil {
		return "", domain.ErrInvalidToken
	}
	email := domain.CanonicalEmail(claims.Subject)
	if !domain.LooksLikeEmail(email) {
		return "", domain.ErrInvalidToken
	}
	return email, nil
}

func (s *AuthService) otpTTL() time.Duration {
	if s.cfg != nil && s.cfg.OTPTTL > 0 {
		return s.cfg.OTPTTL
	}
	return defaultOTPTTL
}

func (s *AuthService) otpMaxAttempts() int {
	if s.cfg != nil && s.cfg.OTPMaxAttempts > 0 {
		return s.cfg.OTPMaxAttempts
	}
	return defaultOTPAttempts
}

func (s *AuthService) otpResendAfter() time.Duration {
	if s.cfg != nil && s.cfg.OTPResendAfter > 0 {
		return s.cfg.OTPResendAfter
	}
	return defaultOTPResendAfter
}

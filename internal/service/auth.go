package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/sakid00/enmasse-be/internal/auth"
	"github.com/sakid00/enmasse-be/internal/config"
	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/mail"
	"github.com/sakid00/enmasse-be/internal/passwrap"
	"github.com/sakid00/enmasse-be/internal/store"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

type AuthService struct {
	db        *store.DB
	cfg       *config.Config
	turnstile *turnstile.Verifier
	mail      mail.Mailer
	wrap      *passwrap.Box
}

func NewAuthService(db *store.DB, cfg *config.Config, ts *turnstile.Verifier, mailer mail.Mailer, wrap *passwrap.Box) *AuthService {
	return &AuthService{db: db, cfg: cfg, turnstile: ts, mail: mailer, wrap: wrap}
}

func (s *AuthService) PasswordPublicJWK() (passwrap.PublicJWK, error) {
	if s == nil || s.wrap == nil {
		return passwrap.PublicJWK{}, domain.ErrUnavailable
	}
	return s.wrap.PublicJWK(), nil
}

func (s *AuthService) UnwrapPassword(plain, cipher string) (string, error) {
	cipher = strings.TrimSpace(cipher)
	if cipher == "" {
		return plain, nil
	}
	if s == nil || s.wrap == nil {
		return "", domain.ErrPasswordPayload
	}
	out, err := s.wrap.Unwrap(cipher)
	if err != nil {
		return "", domain.ErrPasswordPayload
	}
	return out, nil
}

type RegisterArtistInput struct {
	SignupToken string
	Password    string
	FirstName   string
	LastName    string
	ArtistName  string
	Handle      string
}

type RegisterVendorInput struct {
	SignupToken string
	Password    string
	Handle      string
	Name        string
}

type Session struct {
	AccessToken     string   `json:"accessToken"`
	ExpiresIn       int      `json:"expiresIn"`
	ProfileComplete bool     `json:"profileComplete"`
	User            UserView `json:"user"`
}

type UserView struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Handle        string `json:"handle"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"emailVerified"`
}

func (s *AuthService) RegisterArtist(ctx context.Context, in RegisterArtistInput) (*Session, error) {
	email, err := s.emailFromSignupToken(in.SignupToken)
	if err != nil {
		return nil, err
	}
	_, handle, err := validateNewAccount(email, in.Handle, in.Password)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" || strings.TrimSpace(in.ArtistName) == "" {
		return nil, domain.NewAppError(400, "validation", "firstName, lastName, and artistName are required")
	}
	hash, err := auth.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register artist: hash: %w", err)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("register artist: begin: %w", err)
	}
	defer tx.Rollback(ctx)
	q := s.db.Queries.WithTx(tx)

	if err := ensureUnique(ctx, q, email, handle); err != nil {
		return nil, err
	}

	user, err := q.CreateUser(ctx, store.CreateUserParams{
		Email:              email,
		PasswordHash:       &hash,
		Role:               "artist",
		Handle:             handle,
		MustChangePassword: false,
	})
	if err != nil {
		return nil, fmt.Errorf("register artist: user: %w", err)
	}
	if _, err := q.CreateArtist(ctx, store.CreateArtistParams{
		UserID:     user.ID,
		FirstName:  strings.TrimSpace(in.FirstName),
		LastName:   strings.TrimSpace(in.LastName),
		ArtistName: strings.TrimSpace(in.ArtistName),
	}); err != nil {
		return nil, fmt.Errorf("register artist: artist: %w", err)
	}
	if err := q.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("register artist: verify: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("register artist: commit: %w", err)
	}
	user.EmailVerifiedAt = verifiedNow()
	complete, err := s.profileComplete(ctx, user)
	if err != nil {
		return nil, err
	}
	return s.session(user, complete)
}

func (s *AuthService) RegisterVendor(ctx context.Context, in RegisterVendorInput) (*Session, error) {
	email, err := s.emailFromSignupToken(in.SignupToken)
	if err != nil {
		return nil, err
	}
	_, handle, err := validateNewAccount(email, in.Handle, in.Password)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, domain.NewAppError(400, "validation", "name is required")
	}
	hash, err := auth.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register vendor: hash: %w", err)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("register vendor: begin: %w", err)
	}
	defer tx.Rollback(ctx)
	q := s.db.Queries.WithTx(tx)

	if err := ensureUnique(ctx, q, email, handle); err != nil {
		return nil, err
	}

	user, err := q.CreateUser(ctx, store.CreateUserParams{
		Email:              email,
		PasswordHash:       &hash,
		Role:               "vendor",
		Handle:             handle,
		MustChangePassword: false,
	})
	if err != nil {
		return nil, fmt.Errorf("register vendor: user: %w", err)
	}
	if _, err := q.CreateVendor(ctx, store.CreateVendorParams{
		UserID: user.ID,
		Name:   strings.TrimSpace(in.Name),
		City:   "",
	}); err != nil {
		return nil, fmt.Errorf("register vendor: vendor: %w", err)
	}
	if err := q.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("register vendor: verify: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("register vendor: commit: %w", err)
	}
	user.EmailVerifiedAt = verifiedNow()
	complete, err := s.profileComplete(ctx, user)
	if err != nil {
		return nil, err
	}
	return s.session(user, complete)
}

func (s *AuthService) Login(ctx context.Context, email, password, turnstileToken, remoteIP string) (*Session, error) {
	if err := s.turnstile.Verify(ctx, turnstileToken, remoteIP); err != nil {
		return nil, err
	}
	if !domain.LooksLikeEmail(email) {
		return nil, domain.ErrUnauthorized
	}
	user, err := s.db.Queries.GetUserByEmail(ctx, domain.CanonicalEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnauthorized
		}
		return nil, fmt.Errorf("login: %w", err)
	}
	if user.PasswordHash == nil || !auth.Verify(*user.PasswordHash, password) {
		return nil, domain.ErrUnauthorized
	}
	if user.MustChangePassword {
		return nil, domain.ErrMustChangePassword
	}
	complete, err := s.profileComplete(ctx, user)
	if err != nil {
		return nil, err
	}
	return s.session(user, complete)
}

const (
	EmailAvailable  = "available"
	EmailUnclaimed  = "unclaimed"
	EmailRegistered = "registered"
)

type EmailStatus struct {
	Status string `json:"status"`
}

func (s *AuthService) EmailStatus(ctx context.Context, raw string) (*EmailStatus, error) {
	if !domain.LooksLikeEmail(raw) {
		return nil, domain.ErrInvalidEmail
	}
	user, err := s.db.Queries.GetUserByEmail(ctx, domain.CanonicalEmail(raw))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &EmailStatus{Status: EmailAvailable}, nil
		}
		return nil, fmt.Errorf("email status: %w", err)
	}
	if user.PasswordHash == nil {
		return &EmailStatus{Status: EmailUnclaimed}, nil
	}
	return &EmailStatus{Status: EmailRegistered}, nil
}

func (s *AuthService) ClaimPassword(ctx context.Context, raw, password, code string) (*Session, error) {
	if !domain.LooksLikeEmail(raw) {
		return nil, domain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, domain.ErrWeakPassword
	}
	email := domain.CanonicalEmail(raw)
	user, err := s.db.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("claim password: %w", err)
	}
	if user.PasswordHash != nil {
		return nil, domain.ErrNotUnclaimed
	}
	if err := s.consumeOTP(ctx, email, domain.OTPClaim, code); err != nil {
		return nil, err
	}
	hash, err := auth.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("claim password: hash: %w", err)
	}
	if err := s.db.Queries.SetPassword(ctx, store.SetPasswordParams{
		ID:                 user.ID,
		PasswordHash:       &hash,
		MustChangePassword: false,
	}); err != nil {
		return nil, fmt.Errorf("claim password: update: %w", err)
	}
	if err := s.db.Queries.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("claim password: verify: %w", err)
	}
	user.PasswordHash = &hash
	user.EmailVerifiedAt = verifiedNow()
	complete, err := s.profileComplete(ctx, user)
	if err != nil {
		return nil, err
	}
	return s.session(user, complete)
}

func (s *AuthService) Claim(ctx context.Context, email string) error {
	// Uniform 200. Do not leak whether the email exists or is already claimed.
	if !domain.LooksLikeEmail(email) {
		return nil
	}
	user, err := s.db.Queries.GetUserByEmail(ctx, domain.CanonicalEmail(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("claim: %w", err)
	}
	if user.PasswordHash != nil {
		return nil
	}

	plain, err := auth.NewURLToken()
	if err != nil {
		return fmt.Errorf("claim: token: %w", err)
	}
	if err := s.db.Queries.InvalidateUnusedTokens(ctx, store.InvalidateUnusedTokensParams{
		UserID:  user.ID,
		Purpose: "claim",
	}); err != nil {
		return fmt.Errorf("claim: invalidate: %w", err)
	}
	if _, err := s.db.Queries.InsertPasswordToken(ctx, store.InsertPasswordTokenParams{
		UserID:    user.ID,
		TokenHash: auth.HashToken(plain),
		Purpose:   "claim",
		ExpiresAt: time.Now().Add(s.cfg.ClaimTokenTTL),
	}); err != nil {
		return fmt.Errorf("claim: insert: %w", err)
	}

	setURL, err := url.JoinPath(strings.TrimRight(s.cfg.PublicAppURL, "/"), "set-password")
	if err != nil {
		setURL = s.cfg.PublicAppURL + "/set-password"
	}
	setURL = setURL + "?token=" + url.QueryEscape(plain)
	if err := s.mail.SendClaim(ctx, user.Email, setURL); err != nil {
		return fmt.Errorf("claim: mail: %w", err)
	}
	return nil
}

func (s *AuthService) SetPassword(ctx context.Context, token, password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	row, err := s.db.Queries.GetUnusedTokenByHash(ctx, auth.HashToken(strings.TrimSpace(token)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInvalidToken
		}
		return fmt.Errorf("set password: %w", err)
	}
	hash, err := auth.Hash(password)
	if err != nil {
		return fmt.Errorf("set password: hash: %w", err)
	}
	if err := s.db.Queries.SetPassword(ctx, store.SetPasswordParams{
		ID:                 row.UserID,
		PasswordHash:       &hash,
		MustChangePassword: false,
	}); err != nil {
		return fmt.Errorf("set password: update: %w", err)
	}
	if err := s.db.Queries.MarkTokenUsed(ctx, row.ID); err != nil {
		return fmt.Errorf("set password: mark: %w", err)
	}
	if err := s.db.Queries.MarkEmailVerified(ctx, row.UserID); err != nil {
		return fmt.Errorf("set password: verify: %w", err)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, current, next string) error {
	if len(next) < 8 {
		return domain.ErrWeakPassword
	}
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if err := s.requireVerified(user); err != nil {
		return err
	}
	if user.PasswordHash == nil || !auth.Verify(*user.PasswordHash, current) {
		return domain.ErrUnauthorized
	}
	hash, err := auth.Hash(next)
	if err != nil {
		return fmt.Errorf("change password: hash: %w", err)
	}
	return s.db.Queries.SetPassword(ctx, store.SetPasswordParams{
		ID:                 user.ID,
		PasswordHash:       &hash,
		MustChangePassword: false,
	})
}

func (s *AuthService) session(user store.User, complete bool) (*Session, error) {
	tok, err := auth.MintAccess(s.cfg.JWTAccessSecret, s.cfg.JWTIssuer, user.ID.String(), user.Role, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("mint access: %w", err)
	}
	return &Session{
		AccessToken:     tok,
		ExpiresIn:       int(s.cfg.JWTAccessExpiry.Seconds()),
		ProfileComplete: complete,
		User:            userView(user),
	}, nil
}

func validateNewAccount(email, handle, password string) (string, string, error) {
	if !domain.LooksLikeEmail(email) {
		return "", "", domain.ErrInvalidEmail
	}
	if len(password) < 8 {
		return "", "", domain.ErrWeakPassword
	}
	h := domain.CanonicalHandle(handle)
	if !domain.ValidHandle(h) {
		return "", "", domain.ErrInvalidHandle
	}
	return domain.CanonicalEmail(email), h, nil
}

func ensureUnique(ctx context.Context, q *store.Queries, email, handle string) error {
	if _, err := q.GetUserByEmail(ctx, email); err == nil {
		return domain.ErrEmailTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("unique email: %w", err)
	}
	if _, err := q.GetUserByHandle(ctx, handle); err == nil {
		return domain.ErrHandleTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("unique handle: %w", err)
	}
	return nil
}

func (s *AuthService) profileComplete(ctx context.Context, user store.User) (bool, error) {
	complete, _, err := s.completeness(ctx, user)
	return complete, err
}

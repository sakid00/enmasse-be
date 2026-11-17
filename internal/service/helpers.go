package service

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sakid00/enmasse-be/internal/store"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func isEmailVerified(user store.User) bool {
	return user.EmailVerifiedAt.Valid
}

func verifiedNow() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Now(), Valid: true}
}

func userView(user store.User) UserView {
	return UserView{
		ID:            user.ID.String(),
		Email:         user.Email,
		Handle:        user.Handle,
		Role:          user.Role,
		EmailVerified: isEmailVerified(user),
	}
}

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/sakid00/enmasse-be/internal/auth"
	"github.com/sakid00/enmasse-be/internal/domain"
)

type contextKey string

const contextUserID contextKey = "userID"

func HumanAuth(secret, issuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				writeAuthError(w, domain.ErrUnauthorized)
				return
			}
			claims, err := auth.ParseAccess(secret, issuer, token)
			if err != nil {
				writeAuthError(w, domain.ErrUnauthorized)
				return
			}
			id, err := uuid.Parse(claims.Subject)
			if err != nil {
				writeAuthError(w, domain.ErrUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), contextUserID, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ServiceAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				writeAuthError(w, domain.ErrUnauthorized)
				return
			}
			if _, err := auth.ParseService(secret, token); err != nil {
				writeAuthError(w, domain.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(contextUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return id, nil
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func writeAuthError(w http.ResponseWriter, err *domain.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err})
}

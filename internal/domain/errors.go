package domain

import (
	"errors"
	"net/http"
)

type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func NewAppError(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

var (
	ErrNotFound           = NewAppError(http.StatusNotFound, "not_found", "not found")
	ErrUnauthorized       = NewAppError(http.StatusUnauthorized, "unauthorized", "unauthorized")
	ErrForbidden          = NewAppError(http.StatusForbidden, "forbidden", "forbidden")
	ErrConflict           = NewAppError(http.StatusConflict, "conflict", "conflict")
	ErrBadRequest         = NewAppError(http.StatusBadRequest, "validation", "bad request")
	ErrTurnstileRequired  = NewAppError(http.StatusBadRequest, "validation", "turnstile token required")
	ErrTurnstileFailed    = NewAppError(http.StatusBadRequest, "validation", "turnstile verification failed")
	ErrMustChangePassword = NewAppError(http.StatusForbidden, "must_change_password", "password must be changed")
	ErrProfileIncomplete  = NewAppError(http.StatusForbidden, "profile_incomplete", "profile incomplete")
	ErrEmailTaken         = NewAppError(http.StatusConflict, "conflict", "email already registered")
	ErrHandleTaken        = NewAppError(http.StatusConflict, "conflict", "handle already taken")
	ErrWeakPassword       = NewAppError(http.StatusBadRequest, "validation", "password must be at least 8 characters")
	ErrInvalidEmail       = NewAppError(http.StatusBadRequest, "validation", "invalid email")
	ErrInvalidHandle      = NewAppError(http.StatusBadRequest, "validation", "handle must be 3-32 letters, numbers, or underscore")
	ErrInvalidToken       = NewAppError(http.StatusBadRequest, "validation", "invalid or expired token")
	ErrUnavailable        = NewAppError(http.StatusServiceUnavailable, "internal", "service unavailable")
	ErrNotUnclaimed       = NewAppError(http.StatusConflict, "conflict", "email already registered")
)

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

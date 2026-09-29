package handlers

import (
	"net/http"

	"github.com/sakid00/enmasse-be/internal/domain"
)

func newVal(msg string) error {
	return domain.NewAppError(http.StatusBadRequest, "validation", msg)
}

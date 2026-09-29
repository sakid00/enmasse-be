package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/service"
)

type InternalHandler struct {
	svc *service.AuthService
}

func NewInternalHandler(svc *service.AuthService) *InternalHandler {
	return &InternalHandler{svc: svc}
}

func (h *InternalHandler) ProfileStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	complete, missing, err := h.svc.ProfileStatus(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if missing == nil {
		missing = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"complete": complete,
		"missing":  missing,
	})
}

func (h *InternalHandler) Artist(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	a, err := h.svc.InternalArtist(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *InternalHandler) Vendor(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	v, err := h.svc.InternalVendor(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func pathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, domain.NewAppError(http.StatusBadRequest, "validation", "invalid id")
	}
	return id, nil
}

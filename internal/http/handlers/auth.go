package handlers

import (
	"net/http"

	"github.com/google/uuid"

	mw "github.com/sakid00/enmasse-be/internal/http/middleware"
	"github.com/sakid00/enmasse-be/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterArtist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		ArtistName string `json:"artistName"`
		Handle     string `json:"handle"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	sess, err := h.svc.RegisterArtist(r.Context(), service.RegisterArtistInput(body))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (h *AuthHandler) RegisterVendor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Handle   string `json:"handle"`
		Name     string `json:"name"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	sess, err := h.svc.RegisterVendor(r.Context(), service.RegisterVendorInput(body))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		TurnstileToken string `json:"turnstileToken"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	sess, err := h.svc.Login(r.Context(), body.Email, body.Password, body.TurnstileToken, r.RemoteAddr)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (h *AuthHandler) EmailStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	status, err := h.svc.EmailStatus(r.Context(), body.Email)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *AuthHandler) ClaimPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	sess, err := h.svc.ClaimPassword(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (h *AuthHandler) Claim(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	if err := h.svc.Claim(r.Context(), body.Email); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	if err := h.svc.SetPassword(r.Context(), body.Token, body.Password); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Current  string `json:"currentPassword"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, errBadJSON())
		return
	}
	id, err := userID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.ChangePassword(r.Context(), id, body.Current, body.Password); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func userID(r *http.Request) (uuid.UUID, error) {
	return mw.UserIDFromCtx(r.Context())
}

func errBadJSON() error {
	return badJSON
}

var badJSON = errVal("invalid json")

func errVal(msg string) error {
	return newVal(msg)
}

package handlers

import (
	"net/http"

	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/service"
)

type MeHandler struct {
	svc *service.AuthService
}

func NewMeHandler(svc *service.AuthService) *MeHandler {
	return &MeHandler{svc: svc}
}

func (h *MeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := userID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := h.svc.GetProfile(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *MeHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := userID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		FirstName       *string `json:"firstName"`
		LastName        *string `json:"lastName"`
		ArtistName      *string `json:"artistName"`
		PhotoURL        *string `json:"photoUrl"`
		Address         *string `json:"address"`
		Province        *string `json:"province"`
		ProvinceCode    *string `json:"provinceCode"`
		City            *string `json:"city"`
		CityCode        *string `json:"cityCode"`
		District        *string `json:"district"`
		DistrictCode    *string `json:"districtCode"`
		WebsiteURL      *string `json:"websiteUrl"`
		InstagramHandle *string `json:"instagramHandle"`
		TwitterHandle   *string `json:"twitterHandle"`
		WhatsApp        *string `json:"whatsapp"`
		Bio             *string `json:"bio"`
		Name            *string `json:"name"`
		PICName         *string `json:"picName"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, domain.NewAppError(http.StatusBadRequest, "validation", "invalid json"))
		return
	}
	p, err := h.svc.PatchProfile(r.Context(), id, service.PatchProfileInput{
		FirstName:       body.FirstName,
		LastName:        body.LastName,
		ArtistName:      body.ArtistName,
		PhotoURL:        body.PhotoURL,
		Address:         body.Address,
		Province:        body.Province,
		ProvinceCode:    body.ProvinceCode,
		City:            body.City,
		CityCode:        body.CityCode,
		District:        body.District,
		DistrictCode:    body.DistrictCode,
		WebsiteURL:      body.WebsiteURL,
		InstagramHandle: body.InstagramHandle,
		TwitterHandle:   body.TwitterHandle,
		WhatsApp:        body.WhatsApp,
		Bio:             body.Bio,
		Name:            body.Name,
		PICName:         body.PICName,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

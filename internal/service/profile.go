package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/sakid00/enmasse-be/internal/domain"
	"github.com/sakid00/enmasse-be/internal/store"
)

type Profile struct {
	User            UserView       `json:"user"`
	Role            string         `json:"role"`
	ProfileComplete bool           `json:"profileComplete"`
	Missing         []string       `json:"missing"`
	Artist          *ArtistProfile `json:"artist,omitempty"`
	Vendor          *VendorProfile `json:"vendor,omitempty"`
}

type ArtistProfile struct {
	ID              string `json:"id"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	ArtistName      string `json:"artistName"`
	PhotoURL        string `json:"photoUrl"`
	Address         string `json:"address"`
	Province        string `json:"province"`
	ProvinceCode    string `json:"provinceCode"`
	City            string `json:"city"`
	CityCode        string `json:"cityCode"`
	District        string `json:"district"`
	DistrictCode    string `json:"districtCode"`
	WebsiteURL      string `json:"websiteUrl"`
	InstagramHandle string `json:"instagramHandle"`
	TwitterHandle   string `json:"twitterHandle"`
	WhatsApp        string `json:"whatsapp"`
	Bio             string `json:"bio"`
	Nationality     string `json:"nationality"`
}

type VendorProfile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PhotoURL     string `json:"photoUrl"`
	Province     string `json:"province"`
	ProvinceCode string `json:"provinceCode"`
	City         string `json:"city"`
	CityCode     string `json:"cityCode"`
	District     string `json:"district"`
	DistrictCode string `json:"districtCode"`
	PICName      string `json:"picName"`
	WhatsApp     string `json:"whatsapp"`
	Address      string `json:"address"`
	Bio          string `json:"bio"`
}

type PatchProfileInput struct {
	FirstName       *string
	LastName        *string
	ArtistName      *string
	PhotoURL        *string
	Address         *string
	Province        *string
	ProvinceCode    *string
	City            *string
	CityCode        *string
	District        *string
	DistrictCode    *string
	WebsiteURL      *string
	InstagramHandle *string
	TwitterHandle   *string
	WhatsApp        *string
	Bio             *string
	Name            *string
	PICName         *string
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return s.buildProfile(ctx, user)
}

func (s *AuthService) PatchProfile(ctx context.Context, userID uuid.UUID, in PatchProfileInput) (*Profile, error) {
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("patch profile: %w", err)
	}

	switch user.Role {
	case "artist":
		cur, err := s.db.Queries.GetArtistByUserID(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("patch profile: artist: %w", err)
		}
		wa, err := coalesceWhatsApp(in.WhatsApp, cur.Whatsapp)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Queries.UpdateArtistProfile(ctx, store.UpdateArtistProfileParams{
			UserID:          user.ID,
			FirstName:       pick(in.FirstName, cur.FirstName),
			LastName:        pick(in.LastName, cur.LastName),
			ArtistName:      pick(in.ArtistName, cur.ArtistName),
			PhotoUrl:        coalescePtr(in.PhotoURL, cur.PhotoUrl),
			Address:         coalescePtr(in.Address, cur.Address),
			WebsiteUrl:      coalescePtr(in.WebsiteURL, cur.WebsiteUrl),
			InstagramHandle: coalescePtr(in.InstagramHandle, cur.InstagramHandle),
			TwitterHandle:   coalescePtr(in.TwitterHandle, cur.TwitterHandle),
			Whatsapp:        wa,
			Bio:             coalescePtr(in.Bio, cur.Bio),
			Province:        coalescePtr(in.Province, cur.Province),
			ProvinceCode:    coalescePtr(in.ProvinceCode, cur.ProvinceCode),
			City:            coalescePtr(in.City, cur.City),
			CityCode:        coalescePtr(in.CityCode, cur.CityCode),
			District:        coalescePtr(in.District, cur.District),
			DistrictCode:    coalescePtr(in.DistrictCode, cur.DistrictCode),
		})
		if err != nil {
			return nil, fmt.Errorf("patch profile: update artist: %w", err)
		}
	case "vendor":
		cur, err := s.db.Queries.GetVendorByUserID(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("patch profile: vendor: %w", err)
		}
		wa, err := coalesceWhatsApp(in.WhatsApp, cur.Whatsapp)
		if err != nil {
			return nil, err
		}
		_, err = s.db.Queries.UpdateVendor(ctx, store.UpdateVendorParams{
			UserID:       user.ID,
			Name:         pick(in.Name, cur.Name),
			City:         pick(in.City, cur.City),
			PicName:      coalescePtr(in.PICName, cur.PicName),
			Whatsapp:     wa,
			Address:      coalescePtr(in.Address, cur.Address),
			Province:     coalescePtr(in.Province, cur.Province),
			ProvinceCode: coalescePtr(in.ProvinceCode, cur.ProvinceCode),
			CityCode:     coalescePtr(in.CityCode, cur.CityCode),
			District:     coalescePtr(in.District, cur.District),
			DistrictCode: coalescePtr(in.DistrictCode, cur.DistrictCode),
			PhotoUrl:     coalescePtr(in.PhotoURL, cur.PhotoUrl),
			Bio:          coalescePtr(in.Bio, cur.Bio),
		})
		if err != nil {
			return nil, fmt.Errorf("patch profile: update vendor: %w", err)
		}
	}
	return s.buildProfile(ctx, user)
}

func (s *AuthService) ProfileStatus(ctx context.Context, userID uuid.UUID) (complete bool, missing []string, err error) {
	user, err := s.db.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, domain.ErrNotFound
		}
		return false, nil, fmt.Errorf("profile status: %w", err)
	}
	return s.completeness(ctx, user)
}

func (s *AuthService) InternalArtist(ctx context.Context, artistID uuid.UUID) (*ArtistProfile, error) {
	artist, err := s.db.Queries.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("internal artist: %w", err)
	}
	return artistView(artist), nil
}

func (s *AuthService) InternalVendor(ctx context.Context, vendorID uuid.UUID) (*VendorProfile, error) {
	vendor, err := s.db.Queries.GetVendorByID(ctx, vendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("internal vendor: %w", err)
	}
	return vendorView(vendor), nil
}

func (s *AuthService) buildProfile(ctx context.Context, user store.User) (*Profile, error) {
	out := &Profile{
		User: UserView{
			ID:     user.ID.String(),
			Email:  user.Email,
			Handle: user.Handle,
			Role:   user.Role,
		},
		Role:    user.Role,
		Missing: []string{},
	}
	complete, missing, err := s.completeness(ctx, user)
	if err != nil {
		return nil, err
	}
	out.ProfileComplete = complete
	if missing != nil {
		out.Missing = missing
	}
	switch user.Role {
	case "artist":
		artist, err := s.db.Queries.GetArtistByUserID(ctx, user.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("build profile: %w", err)
		}
		if err == nil {
			out.Artist = artistView(artist)
		}
	case "vendor":
		vendor, err := s.db.Queries.GetVendorByUserID(ctx, user.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("build profile: %w", err)
		}
		if err == nil {
			out.Vendor = vendorView(vendor)
		}
	}
	return out, nil
}

func (s *AuthService) completeness(ctx context.Context, user store.User) (bool, []string, error) {
	switch user.Role {
	case "artist":
		artist, err := s.db.Queries.GetArtistByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				miss := domain.ProfileFields{}.Missing()
				return false, miss, nil
			}
			return false, nil, fmt.Errorf("profile complete: %w", err)
		}
		fields := artistFields(artist)
		return fields.Complete(), fields.Missing(), nil
	case "vendor":
		vendor, err := s.db.Queries.GetVendorByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				miss := domain.VendorFields{}.Missing()
				return false, miss, nil
			}
			return false, nil, fmt.Errorf("profile complete: %w", err)
		}
		fields := vendorFields(vendor)
		return fields.Complete(), fields.Missing(), nil
	default:
		return false, []string{}, nil
	}
}

func artistFields(a store.Artist) domain.ProfileFields {
	return domain.ProfileFields{
		PhotoURL:        deref(a.PhotoUrl),
		Address:         deref(a.Address),
		Province:        deref(a.Province),
		City:            deref(a.City),
		District:        deref(a.District),
		WebsiteURL:      deref(a.WebsiteUrl),
		InstagramHandle: deref(a.InstagramHandle),
		TwitterHandle:   deref(a.TwitterHandle),
		WhatsApp:        deref(a.Whatsapp),
		Bio:             deref(a.Bio),
	}
}

func vendorFields(v store.Vendor) domain.VendorFields {
	return domain.VendorFields{
		Name:     v.Name,
		Province: deref(v.Province),
		City:     v.City,
		District: deref(v.District),
		PICName:  deref(v.PicName),
		Address:  deref(v.Address),
		WhatsApp: deref(v.Whatsapp),
		Bio:      deref(v.Bio),
	}
}

func vendorView(v store.Vendor) *VendorProfile {
	return &VendorProfile{
		ID:           v.ID.String(),
		Name:         v.Name,
		PhotoURL:     deref(v.PhotoUrl),
		Province:     deref(v.Province),
		ProvinceCode: deref(v.ProvinceCode),
		City:         v.City,
		CityCode:     deref(v.CityCode),
		District:     deref(v.District),
		DistrictCode: deref(v.DistrictCode),
		PICName:      deref(v.PicName),
		WhatsApp:     deref(v.Whatsapp),
		Address:      deref(v.Address),
		Bio:          deref(v.Bio),
	}
}

func artistView(a store.Artist) *ArtistProfile {
	return &ArtistProfile{
		ID:              a.ID.String(),
		FirstName:       a.FirstName,
		LastName:        a.LastName,
		ArtistName:      a.ArtistName,
		PhotoURL:        deref(a.PhotoUrl),
		Address:         deref(a.Address),
		Province:        deref(a.Province),
		ProvinceCode:    deref(a.ProvinceCode),
		City:            deref(a.City),
		CityCode:        deref(a.CityCode),
		District:        deref(a.District),
		DistrictCode:    deref(a.DistrictCode),
		WebsiteURL:      deref(a.WebsiteUrl),
		InstagramHandle: deref(a.InstagramHandle),
		TwitterHandle:   deref(a.TwitterHandle),
		WhatsApp:        deref(a.Whatsapp),
		Bio:             deref(a.Bio),
		Nationality:     deref(a.Nationality),
	}
}

func pick(override *string, current string) string {
	if override == nil {
		return current
	}
	return strings.TrimSpace(*override)
}

func coalescePtr(override *string, current *string) *string {
	if override == nil {
		return current
	}
	return strPtr(strings.TrimSpace(*override))
}

func coalesceWhatsApp(override *string, current *string) (*string, error) {
	if override == nil {
		return current, nil
	}
	norm, err := domain.NormalizeWhatsApp(*override)
	if err != nil {
		return nil, err
	}
	return strPtr(norm), nil
}

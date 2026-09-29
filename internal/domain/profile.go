package domain

import "strings"

const (
	MissingPhoto    = "photo_url"
	MissingAddress  = "address"
	MissingSocial   = "social"
	MissingWhatsApp = "whatsapp"
	MissingBio      = "bio"
	MissingPIC      = "pic_name"
	MissingName     = "name"
	MissingCity     = "city"
	MissingProvince = "province"
	MissingDistrict = "district"
)

type ProfileFields struct {
	PhotoURL        string
	Address         string
	Province        string
	City            string
	District        string
	WebsiteURL      string
	InstagramHandle string
	TwitterHandle   string
	WhatsApp        string
	Bio             string
}

func (p ProfileFields) Missing() []string {
	var missing []string
	if strings.TrimSpace(p.Address) == "" {
		missing = append(missing, MissingAddress)
	}
	if strings.TrimSpace(p.Province) == "" {
		missing = append(missing, MissingProvince)
	}
	if strings.TrimSpace(p.City) == "" {
		missing = append(missing, MissingCity)
	}
	if strings.TrimSpace(p.District) == "" {
		missing = append(missing, MissingDistrict)
	}
	if strings.TrimSpace(p.Bio) == "" {
		missing = append(missing, MissingBio)
	}
	hasSocial := strings.TrimSpace(p.InstagramHandle) != "" ||
		strings.TrimSpace(p.TwitterHandle) != "" ||
		strings.TrimSpace(p.WebsiteURL) != ""
	if !hasSocial {
		missing = append(missing, MissingSocial)
	}
	if !ValidWhatsApp(p.WhatsApp) {
		missing = append(missing, MissingWhatsApp)
	}
	return missing
}

func (p ProfileFields) Complete() bool {
	return len(p.Missing()) == 0
}

type VendorFields struct {
	Name     string
	Province string
	City     string
	District string
	PICName  string
	Address  string
	WhatsApp string
	Bio      string
}

func (p VendorFields) Missing() []string {
	var missing []string
	if strings.TrimSpace(p.Name) == "" {
		missing = append(missing, MissingName)
	}
	if strings.TrimSpace(p.Province) == "" {
		missing = append(missing, MissingProvince)
	}
	if strings.TrimSpace(p.City) == "" {
		missing = append(missing, MissingCity)
	}
	if strings.TrimSpace(p.District) == "" {
		missing = append(missing, MissingDistrict)
	}
	if strings.TrimSpace(p.PICName) == "" {
		missing = append(missing, MissingPIC)
	}
	if strings.TrimSpace(p.Address) == "" {
		missing = append(missing, MissingAddress)
	}
	if !ValidWhatsApp(p.WhatsApp) {
		missing = append(missing, MissingWhatsApp)
	}
	if strings.TrimSpace(p.Bio) == "" {
		missing = append(missing, MissingBio)
	}
	return missing
}

func (p VendorFields) Complete() bool {
	return len(p.Missing()) == 0
}

func ValidWhatsApp(raw string) bool {
	got, err := NormalizeWhatsApp(raw)
	return err == nil && got != ""
}

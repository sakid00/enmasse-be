package domain_test

import (
	"testing"

	"github.com/sakid00/enmasse-be/internal/domain"
)

func TestProfileComplete(t *testing.T) {
	t.Parallel()
	full := domain.ProfileFields{
		PhotoURL:        "https://cdn.example/p.jpg",
		Address:         "Jl. Contoh No. 1",
		Province:        "DKI Jakarta",
		City:            "Kota Administrasi Jakarta Selatan",
		District:        "Kebayoran Baru",
		InstagramHandle: "artist",
		WhatsApp:        "+628111111111",
		Bio:             "Illustrator",
	}
	if !full.Complete() {
		t.Fatalf("expected complete, missing %v", full.Missing())
	}

	empty := domain.ProfileFields{}
	got := empty.Missing()
	if len(got) != 7 {
		t.Fatalf("missing = %v, want 7 items", got)
	}
	hasBio := false
	for _, m := range got {
		if m == domain.MissingBio {
			hasBio = true
		}
	}
	if !hasBio {
		t.Fatalf("expected bio in missing: %v", got)
	}

	noPhoto := domain.ProfileFields{
		Address:         "Jl. Contoh No. 1",
		Province:        "DKI Jakarta",
		City:            "Kota Administrasi Jakarta Selatan",
		District:        "Kebayoran Baru",
		InstagramHandle: "artist",
		WhatsApp:        "+628111111111",
		Bio:             "Illustrator",
	}
	if !noPhoto.Complete() {
		t.Fatalf("photo is optional, missing %v", noPhoto.Missing())
	}

	webOnly := domain.ProfileFields{
		Address:    "Jl. Asia Afrika",
		Province:   "Jawa Barat",
		City:       "Kota Bandung",
		District:   "Sumur Bandung",
		WebsiteURL: "https://artist.example",
		WhatsApp:   "+6281234567890",
		Bio:        "Works on paper",
	}
	if !webOnly.Complete() {
		t.Fatalf("website should count as social, missing %v", webOnly.Missing())
	}

	localWA := domain.ProfileFields{
		Address:    "Jl. Sudirman",
		Province:   "DKI Jakarta",
		City:       "Kota Administrasi Jakarta Pusat",
		District:   "Tanah Abang",
		WebsiteURL: "https://a.example",
		WhatsApp:   "081234567890",
		Bio:        "ok",
	}
	if !localWA.Complete() {
		t.Fatalf("08 should normalize as valid whatsapp, missing %v", localWA.Missing())
	}
}

func TestVendorFieldsIncompleteAfterSignup(t *testing.T) {
	t.Parallel()
	got := domain.VendorFields{Name: "Studio"}.Missing()
	want := []string{
		domain.MissingProvince,
		domain.MissingCity,
		domain.MissingDistrict,
		domain.MissingPIC,
		domain.MissingAddress,
		domain.MissingWhatsApp,
		domain.MissingBio,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i, token := range want {
		if got[i] != token {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

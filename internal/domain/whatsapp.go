package domain

import (
	"strings"
	"unicode"
)

var ErrInvalidWhatsApp = NewAppError(400, "validation", "whatsapp must be a +62 number")

func NormalizeWhatsApp(raw string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		if unicode.IsDigit(r) || r == '+' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" {
		return "", nil
	}
	switch {
	case strings.HasPrefix(s, "+62"):
		s = "+62" + strings.TrimPrefix(s, "+62")
	case strings.HasPrefix(s, "62"):
		s = "+" + s
	case strings.HasPrefix(s, "08"):
		s = "+62" + s[1:]
	case strings.HasPrefix(s, "8"):
		s = "+62" + s
	default:
		return "", ErrInvalidWhatsApp
	}
	rest := strings.TrimPrefix(s, "+62")
	if rest == "" || rest[0] == '0' || len(rest) < 8 || len(rest) > 13 {
		return "", ErrInvalidWhatsApp
	}
	for _, r := range rest {
		if !unicode.IsDigit(r) {
			return "", ErrInvalidWhatsApp
		}
	}
	return "+62" + rest, nil
}

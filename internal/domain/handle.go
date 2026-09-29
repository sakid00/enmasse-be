package domain

import (
	"regexp"
	"strings"
	"unicode"
)

var handleRE = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

func CanonicalHandle(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidHandle(raw string) bool {
	return handleRE.MatchString(CanonicalHandle(raw))
}

func HandleFromEmail(email string) string {
	email = CanonicalEmail(email)
	at := strings.Index(email, "@")
	if at <= 0 {
		return ""
	}
	local := email[:at]
	var b strings.Builder
	for _, r := range local {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	h := b.String()
	if len(h) < 3 {
		h = h + "art"
	}
	if len(h) > 32 {
		h = h[:32]
	}
	return h
}

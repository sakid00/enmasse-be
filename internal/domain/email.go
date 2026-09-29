package domain

import (
	"strings"
	"unicode"
)

// CanonicalEmail is the uniqueness key.
// Always: trim, lowercase the whole string.
// For gmail.com / googlemail.com only: strip dots and +tags in the local-part.
func CanonicalEmail(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return s
	}
	local, domain := s[:at], s[at+1:]
	if domain == "googlemail.com" {
		domain = "gmail.com"
	}
	if domain == "gmail.com" {
		if plus := strings.Index(local, "+"); plus >= 0 {
			local = local[:plus]
		}
		local = strings.ReplaceAll(local, ".", "")
	}
	return local + "@" + domain
}

func LooksLikeEmail(s string) bool {
	s = strings.TrimSpace(s)
	at := strings.Index(s, "@")
	if at <= 0 || at != strings.LastIndex(s, "@") {
		return false
	}
	local, domain := s[:at], s[at+1:]
	if local == "" || !strings.Contains(domain, ".") {
		return false
	}
	for _, r := range s {
		if unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

package models

import "strings"

type Hreflang struct {
	URL  string
	Lang string
}

// NormalizeLang returns a normalized language code so hreflang values that
// differ only in case or surrounding whitespace are considered equal.
// Language codes are case-insensitive (e.g. "EN", "en" and "de-CH", "de-ch"
// identify the same language).
func NormalizeLang(lang string) string {
	return strings.ToLower(strings.TrimSpace(lang))
}

package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

// MakeSlug converts any string to a clean URL-friendly slug
func MakeSlug(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "item"
	}

	// Normalize unicode and remove accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, err := transform.String(t, value)
	if err != nil {
		normalized = value
	}

	lower := strings.ToLower(normalized)
	slug := nonAlphaNumRegex.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "item"
	}
	return slug
}

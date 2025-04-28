package utils

import (
	"net/url"
	"strings"
)

// NormalizeURL normalizes a URL by:
// - Converting to lowercase
// - Removing fragments (#)
// - Removing query parameters (?)
// - Adding scheme if missing
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(rawURL)
	}

	// Add scheme if missing
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	// Remove fragment and query
	u.Fragment = ""
	u.RawQuery = ""

	// Standardize host
	u.Host = strings.ToLower(u.Host)

	return u.String()
}
package urlutils

import (
	"net/url"
	"strings"
)

// NormalizeURL returns a normalized form of the given URL string so URLs that
// refer to the same resource but differ in scheme or host case, default port
// or fragment map to the same value. It is used to build and reconcile the
// link and hreflang graphs. URLs that cannot be parsed are returned trimmed
// but otherwise unchanged.
func NormalizeURL(s string) string {
	trimmed := strings.TrimSpace(s)

	u, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove default ports so http://example.com:80 and http://example.com
	// are considered the same URL.
	if port := u.Port(); (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		u.Host = u.Hostname()
	}

	u.Fragment = ""

	if u.Host != "" && u.Path == "" {
		u.Path = "/"
	}

	return u.String()
}

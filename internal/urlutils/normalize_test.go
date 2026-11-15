package urlutils_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/urlutils"
)

// Test NormalizeURL maps URLs that refer to the same resource to the same
// value: scheme and host case, default ports and fragments are normalized.
func TestNormalizeURL(t *testing.T) {
	table := []struct {
		url  string
		want string
	}{
		{"https://example.com", "https://example.com/"},
		{"https://example.com/", "https://example.com/"},
		{"HTTPS://EXAMPLE.COM/", "https://example.com/"},
		{"https://Example.COM/Path", "https://example.com/Path"},
		{"https://example.com:443/", "https://example.com/"},
		{"http://example.com:80/", "http://example.com/"},
		{"http://example.com:8080/", "http://example.com:8080/"},
		{"https://example.com/page#section", "https://example.com/page"},
		{"https://example.com/a?b=1&c=2", "https://example.com/a?b=1&c=2"},
		{"  https://example.com/spaced  ", "https://example.com/spaced"},
		{"/relative/path", "/relative/path"},
	}

	for _, tc := range table {
		if got := urlutils.NormalizeURL(tc.url); got != tc.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

// Test that equivalent URLs normalize to the same value.
func TestNormalizeURLEquivalence(t *testing.T) {
	equivalent := []string{
		"https://example.com/fr",
		"HTTPS://EXAMPLE.COM/fr",
		"https://example.com:443/fr",
		"https://example.com/fr#top",
	}

	want := urlutils.NormalizeURL(equivalent[0])
	for _, u := range equivalent[1:] {
		if got := urlutils.NormalizeURL(u); got != want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", u, got, want)
		}
	}
}

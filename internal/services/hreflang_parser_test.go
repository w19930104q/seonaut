package services_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

// hreflangSet returns a set of the page report's hreflangs for easy comparison.
func hreflangSet(hreflangs []models.Hreflang) map[models.Hreflang]int {
	set := make(map[models.Hreflang]int)
	for _, h := range hreflangs {
		set[h]++
	}

	return set
}

// Test that hreflangs from the HTML link tags and the HTTP Link headers are
// merged and that duplicates present in both sources are reported only once.
func TestHreflangMergeHTMLAndHeaders(t *testing.T) {
	u, err := url.Parse("https://example.com/en")
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`
		<html>
			<head>
				<link rel="alternate" href="https://example.com/en" hreflang="en" />
				<link rel="alternate" href="https://example.com/fr" hreflang="fr" />
			</head>
		</html>`)

	headers := http.Header{
		"Content-Type": []string{"text/html"},
		"Link": []string{`
			<https://example.com/fr>; rel="alternate"; hreflang="fr",
			<https://example.com/de>; rel="alternate"; hreflang="de"
		`},
	}

	pageReport, _, err := services.NewHTMLParser(u, 200, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	want := map[models.Hreflang]int{
		{URL: "https://example.com/en", Lang: "en"}: 1,
		{URL: "https://example.com/fr", Lang: "fr"}: 1,
		{URL: "https://example.com/de", Lang: "de"}: 1,
	}

	got := hreflangSet(pageReport.Hreflangs)
	if len(got) != len(want) {
		t.Errorf("HreflangMerge: got %v, want %v", got, want)
	}

	for h := range want {
		if got[h] != 1 {
			t.Errorf("HreflangMerge: hreflang %v found %d times, want 1", h, got[h])
		}
	}
}

// Test that rel="alternate" link tags without an hreflang attribute, or with
// an empty one, are not included in the hreflangs.
func TestHreflangSkipsEntriesWithoutLang(t *testing.T) {
	u, err := url.Parse("https://example.com/en")
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`
		<html>
			<head>
				<link rel="alternate" href="https://example.com/en" hreflang="en" />
				<link rel="alternate" href="https://example.com/no-lang" />
				<link rel="alternate" href="https://example.com/empty-lang" hreflang="" />
			</head>
		</html>`)

	headers := http.Header{
		"Content-Type": []string{"text/html"},
		"Link": []string{`
			<https://example.com/header-no-lang>; rel="alternate",
			<https://example.com/fr>; rel="alternate"; hreflang="fr"
		`},
	}

	pageReport, _, err := services.NewHTMLParser(u, 200, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	want := map[models.Hreflang]int{
		{URL: "https://example.com/en", Lang: "en"}: 1,
		{URL: "https://example.com/fr", Lang: "fr"}: 1,
	}

	got := hreflangSet(pageReport.Hreflangs)
	if len(got) != len(want) {
		t.Errorf("HreflangSkipsEntriesWithoutLang: got %v, want %v", got, want)
	}

	for h := range want {
		if got[h] != 1 {
			t.Errorf("HreflangSkipsEntriesWithoutLang: hreflang %v found %d times, want 1", h, got[h])
		}
	}
}

// Test that relative hreflang URLs in the HTML tags and the Link headers are
// resolved against the document URL.
func TestHreflangRelativeURLs(t *testing.T) {
	u, err := url.Parse("https://example.com/en/")
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`
		<html>
			<head>
				<link rel="alternate" href="/fr/" hreflang="fr" />
			</head>
		</html>`)

	headers := http.Header{
		"Content-Type": []string{"text/html"},
		"Link":         []string{`</de/>; rel="alternate"; hreflang="de"`},
	}

	pageReport, _, err := services.NewHTMLParser(u, 200, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	want := map[models.Hreflang]int{
		{URL: "https://example.com/fr/", Lang: "fr"}: 1,
		{URL: "https://example.com/de/", Lang: "de"}: 1,
	}

	got := hreflangSet(pageReport.Hreflangs)
	if len(got) != len(want) {
		t.Errorf("HreflangRelativeURLs: got %v, want %v", got, want)
	}

	for h := range want {
		if got[h] != 1 {
			t.Errorf("HreflangRelativeURLs: hreflang %v found %d times, want 1", h, got[h])
		}
	}
}

// Test that hreflang language codes and URLs are normalized so entries that
// differ only in case are deduplicated.
func TestHreflangCaseNormalization(t *testing.T) {
	u, err := url.Parse("https://example.com/en")
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`
		<html>
			<head>
				<link rel="alternate" href="https://example.com/en" hreflang="EN" />
				<link rel="alternate" href="HTTPS://EXAMPLE.COM/en" hreflang="en" />
				<link rel="alternate" href="https://example.com/de-ch" hreflang="de-CH" />
			</head>
		</html>`)

	headers := http.Header{
		"Content-Type": []string{"text/html"},
		"Link":         []string{`<https://EXAMPLE.COM/de-ch>; rel="alternate"; hreflang="DE-ch"`},
	}

	pageReport, _, err := services.NewHTMLParser(u, 200, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	want := map[models.Hreflang]int{
		{URL: "https://example.com/en", Lang: "en"}:       1,
		{URL: "https://example.com/de-ch", Lang: "de-ch"}: 1,
	}

	got := hreflangSet(pageReport.Hreflangs)
	if len(got) != len(want) {
		t.Errorf("HreflangCaseNormalization: got %v, want %v", got, want)
	}

	for h := range want {
		if got[h] != 1 {
			t.Errorf("HreflangCaseNormalization: hreflang %v found %d times, want 1", h, got[h])
		}
	}
}

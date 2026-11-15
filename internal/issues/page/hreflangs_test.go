package page_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/issues/page"
	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/net/html"
)

// Test the HreflangXDefaultMissing reporter with an pagereport that has
// hreflang tags with x-default value.
// Should return false.
func TestHreflangXDefaultMissingNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "x-default"},
			{URL: "http://example.com/en", Lang: "en"},
		},
	}

	reporter := page.NewHreflangXDefaultMissingReporter()
	if reporter.ErrorType != errors.ErrorHreflangMissingXDefault {
		t.Errorf("HreflangXDefaultMissing: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangXDefaultMissing: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("HreflangXDefaultMissing: reportsIssue should be false")
	}
}

// Test the HreflangXDefaultMissing reporter with an pagereport that has
// hreflang tags without x-default value.
// Should return true.
func TestHreflangXDefaultMissingIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "en"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangXDefaultMissingReporter()
	if reporter.ErrorType != errors.ErrorHreflangMissingXDefault {
		t.Errorf("HreflangXDefaultMissing: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangXDefaultMissing: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("HreflangXDefaultMissing: reportsIssue should be true")
	}
}

// Test the HreflangMissingSelfReference reporter with an pagereport that has
// hreflang tags with a self-referencing hreflang tag.
// Should return false.
func TestHreflangMissingSelfReferenceNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		URL:        "http://example.com",
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "en"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangMissingSelfReference()
	if reporter.ErrorType != errors.ErrorHreflangMissingSelfReference {
		t.Errorf("HreflangMissingSelfReference: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMissingSelfReference: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("HreflangMissingSelfReference: reportsIssue should be false")
	}
}

// Test the HreflangMissingSelfReference reporter with an pagereport that has
// hreflang tags without a self-referencing hreflang tag.
// Should return true.
func TestHreflangMissingSelfReferenceIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		URL:        "http://example.com",
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com/en", Lang: "en"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangMissingSelfReference()
	if reporter.ErrorType != errors.ErrorHreflangMissingSelfReference {
		t.Errorf("HreflangMissingSelfReference: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMissingSelfReference: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("HreflangMissingSelfReference: reportsIssue should be true")
	}
}

// Test the HreflangMismatchingLang reporter with an pagereport that has
// a self-referencing hreflang tag using the same lang code as the page language.
// Should return false.
func TestHreflangMismatchingLangNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		Lang:       "en",
		StatusCode: 200,
		URL:        "http://example.com",
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "en"},
			{URL: "http://example.com", Lang: "x-default"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangMismatchingLang()
	if reporter.ErrorType != errors.ErrorHreflangMismatchLang {
		t.Errorf("HreflangMismatchingLang: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMismatchingLang: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("HreflangMismatchingLang: reportsIssue should be false")
	}
}

// Test the HreflangMismatchingLang reporter with an pagereport that has
// a self-referencing hreflang tag using a different lang code.
// Should return false.
func TestHreflangMismatchingLangIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		Lang:       "it",
		StatusCode: 200,
		URL:        "http://example.com",
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "en"},
			{URL: "http://example.com", Lang: "x-default"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangMismatchingLang()
	if reporter.ErrorType != errors.ErrorHreflangMismatchLang {
		t.Errorf("HreflangMismatchingLang: error type is not correct")
	}

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMismatchingLang: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("HreflangMismatchingLang: reportsIssue should be true")
	}
}

// Test the HreflangRelativeURL reporter with an pagereport that has
// hreflang tags with absolute URLs.
// Should return false.
func TestHreflangRelativeURLNoIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewHreflangRelativeURL()
	if reporter.ErrorType != errors.ErrorHreflangRelativeURL {
		t.Errorf("HreflangRelativeURL: error type is not correct")
	}

	source := `
		<html>
			<head>
				<link rel="alternate" href="http://example.com" hreflang="x-default" />
				<link rel="alternate" href="http://example.com/am" hreflang="am" />
			</head>
		</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("HreflangRelativeURL: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == true {
		t.Errorf("HreflangRelativeURL: reportsIssue should be false")
	}
}

// Test the HreflangRelativeURL reporter with an pagereport that has
// hreflang tags with relative URLs.
// Should return true.
func TestHreflangRelativeURLIssues(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewHreflangRelativeURL()
	if reporter.ErrorType != errors.ErrorHreflangRelativeURL {
		t.Errorf("HreflangRelativeURL: error type is not correct")
	}

	source := `
		<html>
			<head>
				<link rel="alternate" href="/" hreflang="x-default" />
				<link rel="alternate" href="/am" hreflang="am" />
			</head>
		</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("HreflangRelativeURL: Error parsing html source")
	}

	reportsIssue := reporter.Callback(pageReport, doc, &http.Header{})

	if reportsIssue == false {
		t.Errorf("HreflangRelativeURL: reportsIssue should be true")
	}
}

// Test the HreflangXDefaultMissing reporter with an x-default value using
// mixed case. Language codes are case-insensitive so it should return false.
func TestHreflangXDefaultMissingMixedCase(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "X-Default"},
			{URL: "http://example.com/en", Lang: "EN"},
		},
	}

	reporter := page.NewHreflangXDefaultMissingReporter()

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangXDefaultMissing: Error parsing html source")
	}

	if reporter.Callback(pageReport, doc, &http.Header{}) == true {
		t.Errorf("HreflangXDefaultMissing mixed case: reportsIssue should be false")
	}
}

// Test the HreflangMissingSelfReference reporter with a self-referencing
// hreflang URL that differs from the page URL only in host case.
// Should return false.
func TestHreflangMissingSelfReferenceCaseInsensitive(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
		URL:        "http://example.com/en",
		Hreflangs: []models.Hreflang{
			{URL: "http://EXAMPLE.com/en", Lang: "en"},
			{URL: "http://example.com/fr", Lang: "fr"},
		},
	}

	reporter := page.NewHreflangMissingSelfReference()

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMissingSelfReference: Error parsing html source")
	}

	if reporter.Callback(pageReport, doc, &http.Header{}) == true {
		t.Errorf("HreflangMissingSelfReference case: reportsIssue should be false")
	}
}

// Test the HreflangMismatchingLang reporter with a self-referencing hreflang
// lang that matches the page lang but uses a different case.
// Should return false.
func TestHreflangMismatchingLangCaseInsensitive(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		Lang:       "en",
		StatusCode: 200,
		URL:        "http://example.com",
		Hreflangs: []models.Hreflang{
			{URL: "http://example.com", Lang: "EN"},
		},
	}

	reporter := page.NewHreflangMismatchingLang()

	doc, err := html.Parse(strings.NewReader(""))
	if err != nil {
		t.Errorf("HreflangMismatchingLang: Error parsing html source")
	}

	if reporter.Callback(pageReport, doc, &http.Header{}) == true {
		t.Errorf("HreflangMismatchingLang case: reportsIssue should be false")
	}
}

// Test the HreflangRelativeURL reporter with a relative rel="alternate" link
// tag that has no hreflang attribute. It should be ignored and return false.
func TestHreflangRelativeURLIgnoresMissingHreflang(t *testing.T) {
	pageReport := &models.PageReport{
		Crawled:    true,
		MediaType:  "text/html",
		StatusCode: 200,
	}

	reporter := page.NewHreflangRelativeURL()

	source := `
		<html>
			<head>
				<link rel="alternate" href="/feed" />
				<link rel="alternate" href="http://example.com/en" hreflang="en" />
			</head>
		</html>
	`

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Errorf("HreflangRelativeURL: Error parsing html source")
	}

	if reporter.Callback(pageReport, doc, &http.Header{}) == true {
		t.Errorf("HreflangRelativeURL without hreflang attr: reportsIssue should be false")
	}
}

package page

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/urlutils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang values do not include an x-default option.
func NewHreflangXDefaultMissingReporter() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if len(pageReport.Hreflangs) == 0 {
			return false
		}

		for _, hreflang := range pageReport.Hreflangs {
			if strings.EqualFold(models.NormalizeLang(hreflang.Lang), "x-default") {
				return false
			}
		}

		return true
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHreflangMissingXDefault,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang values don't include a self-referencing link.
func NewHreflangMissingSelfReference() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if len(pageReport.Hreflangs) == 0 {
			return false
		}

		pageURL := urlutils.NormalizeURL(pageReport.URL)
		for _, hl := range pageReport.Hreflangs {
			if urlutils.NormalizeURL(hl.URL) == pageURL {
				return false
			}
		}

		return true
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHreflangMissingSelfReference,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the self-referencing hreflang lang doesn't match the page's lang.
func NewHreflangMismatchingLang() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		if pageReport.Lang == "" {
			return false
		}

		pageURL := urlutils.NormalizeURL(pageReport.URL)
		for _, hl := range pageReport.Hreflangs {
			if urlutils.NormalizeURL(hl.URL) == pageURL &&
				!strings.EqualFold(hl.Lang, "x-default") &&
				!strings.EqualFold(models.NormalizeLang(hl.Lang), models.NormalizeLang(pageReport.Lang)) {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHreflangMismatchLang,
		Callback:  c,
	}
}

// Returns a report_manager.PageIssueReporter with a callback function that returns true if
// the hreflang URLs are relative. Only link tags with a non-empty hreflang attribute
// are taken into account.
func NewHreflangRelativeURL() *models.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		if !pageReport.Crawled {
			return false
		}

		if pageReport.MediaType != "text/html" {
			return false
		}

		hreflangs, err := htmlquery.QueryAll(htmlNode, "//head/link[@rel=\"alternate\" and @hreflang]")
		if err != nil || hreflangs == nil {
			return false
		}

		for _, hl := range hreflangs {
			if models.NormalizeLang(htmlquery.SelectAttr(hl, "hreflang")) == "" {
				continue
			}

			parsedURL, err := url.Parse(htmlquery.SelectAttr(hl, "href"))
			if err != nil {
				return false
			}

			if !parsedURL.IsAbs() {
				return true
			}
		}

		return false
	}

	return &models.PageIssueReporter{
		ErrorType: errors.ErrorHreflangRelativeURL,
		Callback:  c,
	}
}

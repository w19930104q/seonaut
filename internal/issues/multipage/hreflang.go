package multipage

import (
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have missing
// hreflang return links. The return link is checked for each source and target pair following the
// target's redirect chain to its final URL, so a self-referencing hreflang or a link from an
// unrelated page does not count as a return link. Each page is reported only once.
func (sr *SqlReporter) MissingHrelangReturnLinks(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			DISTINCT pagereports.id
		FROM hreflangs
		INNER JOIN pagereports ON hreflangs.pagereport_id = pagereports.id
		LEFT JOIN pagereports target
			ON target.crawl_id = hreflangs.crawl_id AND target.url_hash = hreflangs.to_hash
		LEFT JOIN pagereports final_page
			ON final_page.crawl_id = hreflangs.crawl_id AND final_page.url_hash = target.redirect_hash
		WHERE hreflangs.crawl_id = ?
			AND hreflangs.to_lang != "x-default"
			AND pagereports.status_code < 300
			AND pagereports.crawled = 1
			AND (pagereports.canonical = "" OR pagereports.canonical = pagereports.URL)
			AND NOT EXISTS (
				SELECT 1
				FROM hreflangs back
				INNER JOIN pagereports bp ON bp.id = back.pagereport_id
				WHERE back.crawl_id = hreflangs.crawl_id
					AND back.to_hash = hreflangs.from_hash
					AND bp.url_hash = COALESCE(final_page.url_hash, target.url_hash, "")
			)`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangsReturnLink,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-canonical pages.
func (sr *SqlReporter) HreflangsToNonCanonical(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			DISTINCT pagereports.id
		FROM pagereports
		LEFT JOIN hreflangs ON hreflangs.to_hash = pagereports.url_hash AND hreflangs.crawl_id = ?
		WHERE media_type = "text/html"
			AND status_code >= 200
			AND status_code < 300
			AND (canonical IS NOT NULL AND canonical != "" AND canonical != url)
			AND pagereports.crawl_id = ?
			AND hreflangs.id IS NOT NULL
			AND crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorHreflangToNonCanonical,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-indexable pages. Each source page is reported only once.
func (sr *SqlReporter) HreflangNoindexable(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			DISTINCT pr.id
		FROM pagereports AS pr
		INNER JOIN hreflangs AS hl ON pr.id = hl.pagereport_id
		INNER JOIN pagereports AS pr2 ON hl.to_hash = pr2.url_hash AND pr2.crawl_id = hl.crawl_id
		WHERE pr.crawl_id = ?
			AND pr2.noindex = 1
			AND pr2.crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangNoindexable,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that have hreflang links pointing to redirects. Each source page is reported only once.
func (sr *SqlReporter) HreflangToRedirect(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			DISTINCT pr.id
		FROM pagereports AS pr
		INNER JOIN hreflangs AS hl ON pr.id = hl.pagereport_id
		INNER JOIN pagereports AS pr2 ON hl.to_hash = pr2.url_hash AND pr2.crawl_id = hl.crawl_id
		WHERE pr.crawl_id = ?
			AND pr2.status_code >= 300
			AND pr2.status_code < 400;`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangToRedirect,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that have hreflang links pointing to error pages with status code 40x or 50x.
// Each source page is reported only once.
func (sr *SqlReporter) HreflangToError(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			DISTINCT pr.id
		FROM pagereports AS pr
		INNER JOIN hreflangs AS hl ON pr.id = hl.pagereport_id
		INNER JOIN pagereports AS pr2 ON hl.to_hash = pr2.url_hash AND pr2.crawl_id = hl.crawl_id
		WHERE pr.crawl_id = ?
			AND pr2.status_code >= 400;`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangToError,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are referenced from hreflang with more than one language.
func (sr *SqlReporter) MultipleLangReference(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT pagereports.id
		FROM hreflangs
		INNER JOIN pagereports
			ON pagereports.url_hash = hreflangs.to_hash AND pagereports.crawl_id = hreflangs.crawl_id
		WHERE hreflangs.crawl_id = ? AND hreflangs.to_lang != "x-default"
		GROUP BY pagereports.id
		HAVING count(distinct hreflangs.to_lang) > 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorMultipleLangReference,
	}
}

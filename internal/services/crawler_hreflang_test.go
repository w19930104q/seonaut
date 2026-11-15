package services_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

// hreflangCrawlRepository is a CrawlerServiceRepository mock that signals
// when the crawl has finished updating the crawl record.
type hreflangCrawlRepository struct {
	done chan struct{}
	once sync.Once
}

func newHreflangCrawlRepository() *hreflangCrawlRepository {
	return &hreflangCrawlRepository{done: make(chan struct{})}
}

func (r *hreflangCrawlRepository) SaveCrawl(p models.Project) (*models.Crawl, error) {
	return &models.Crawl{Id: 1, ProjectId: p.Id, URL: p.URL}, nil
}

func (r *hreflangCrawlRepository) GetLastCrawl(p *models.Project) models.Crawl {
	return models.Crawl{}
}

func (r *hreflangCrawlRepository) GetLastCrawls(p models.Project, limit int) []models.Crawl {
	return []models.Crawl{}
}

func (r *hreflangCrawlRepository) DeleteCrawlData(c *models.Crawl) {}

func (r *hreflangCrawlRepository) CountIssuesByPriority(crawlId int64, priority int) int {
	return 0
}

func (r *hreflangCrawlRepository) UpdateCrawl(c *models.Crawl) {
	r.once.Do(func() { close(r.done) })
}

// hreflangPageReportRepository is a CrawlerHandlerRepository mock that
// captures the saved page reports indexed by their URL.
type hreflangPageReportRepository struct {
	mu      sync.Mutex
	reports map[string]*models.PageReport
}

func newHreflangPageReportRepository() *hreflangPageReportRepository {
	return &hreflangPageReportRepository{reports: make(map[string]*models.PageReport)}
}

func (r *hreflangPageReportRepository) SavePageReport(pr *models.PageReport, crawlId int64) (*models.PageReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports[pr.URL] = pr

	return pr, nil
}

func (r *hreflangPageReportRepository) get(u string) *models.PageReport {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reports[u]
}

// hreflangTestSite returns an httptest server serving a four language site
// (en, fr, de, es). The /en page mixes HTML and Link header hreflangs with
// duplicates, relative URLs and mixed case language codes. The /de page
// points its es hreflang to /es-old, which redirects to /es.
func hreflangTestSite(t *testing.T) *httptest.Server {
	t.Helper()

	var base string
	page := func(lang, self string, alternates ...string) string {
		head := fmt.Sprintf(`<link rel="alternate" hreflang="%s" href="%s" />`, lang, self)
		for _, a := range alternates {
			head += a
		}

		return fmt.Sprintf(`<html lang="%s"><head><title>%s</title>%s</head><body><h1>%s</h1></body></html>`,
			lang, lang, head, lang)
	}

	alternate := func(lang, href string) string {
		return fmt.Sprintf(`<link rel="alternate" hreflang="%s" href="%s" />`, lang, href)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/en", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Add("Link", fmt.Sprintf(`<%s/es>; rel="alternate"; hreflang="es"`, base))
		w.Header().Add("Link", `</de>; rel="alternate"; hreflang="de"`)
		w.Header().Add("Link", fmt.Sprintf(`<%s/fr>; rel="alternate"; hreflang="fr"`, base))
		w.Header().Add("Link", fmt.Sprintf(`<%s/en>; rel="alternate"; hreflang="x-default"`, base))

		fmt.Fprint(w, page("en", "/en",
			alternate("FR", "/fr"),                                    // relative URL, upper case lang
			alternate("fr", base+"/fr"),                               // duplicate of the previous one
			alternate("de", base+"/de"),                               // duplicate of the Link header one
			`<link rel="alternate" href="/no-lang" />`,                // no hreflang attribute
			`<link rel="alternate" hreflang="" href="/empty-lang" />`, // empty hreflang
		))
	})

	mux.HandleFunc("/fr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, page("fr", base+"/fr",
			alternate("en", base+"/en"),
			alternate("x-default", base+"/en"),
		))
	})

	mux.HandleFunc("/de", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, page("de", base+"/de",
			alternate("en", base+"/en"),
			alternate("es", base+"/es-old"), // hreflang to a redirecting URL
			alternate("x-default", base+"/en"),
		))
	})

	mux.HandleFunc("/es", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, page("es", base+"/es",
			alternate("en", base+"/en"),
			alternate("x-default", base+"/en"),
		))
	})

	mux.HandleFunc("/es-old", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", base+"/es")
		w.WriteHeader(http.StatusMovedPermanently)
	})

	server := httptest.NewServer(mux)
	base = server.URL

	t.Cleanup(server.Close)

	return server
}

// TestCrawlHreflangSite crawls a local four language site and checks that
// hreflangs from the HTML tags and the Link headers are merged, deduplicated
// and normalized, that entries without hreflang are skipped and that hreflang
// URLs pointing to redirects are crawled through to their final URL.
func TestCrawlHreflangSite(t *testing.T) {
	server := hreflangTestSite(t)
	base := server.URL

	crawlRepo := newHreflangCrawlRepository()
	pageRepo := newHreflangPageReportRepository()

	broker := services.NewPubSubBroker()
	reportManager := services.NewReportManager(&crawlerReportManagerTestRepository{})
	handler := services.NewCrawlerHandler(pageRepo, broker, reportManager)

	svc := services.NewCrawlerService(crawlRepo, services.CrawlerServicesContainer{
		Broker:         broker,
		ReportManager:  reportManager,
		CrawlerHandler: handler,
		ArchiveService: services.NewArchiveService(""),
		Config:         &config.CrawlerConfig{Agent: "testbot"},
	})

	p := models.Project{Id: 1, URL: base + "/en"}
	if err := svc.StartCrawler(p, models.BasicAuth{}); err != nil {
		t.Fatalf("StartCrawler: %v", err)
	}

	select {
	case <-crawlRepo.done:
	case <-time.After(30 * time.Second):
		t.Fatal("crawl did not finish in time")
	}

	// The /en page merges the HTML and header hreflangs, removes duplicates
	// and entries without hreflang, resolves relative URLs and normalizes
	// the language codes.
	en := pageRepo.get(base + "/en")
	if en == nil {
		t.Fatalf("no page report for %s/en", base)
	}

	want := map[models.Hreflang]int{
		{URL: base + "/en", Lang: "en"}:        1,
		{URL: base + "/fr", Lang: "fr"}:        1,
		{URL: base + "/de", Lang: "de"}:        1,
		{URL: base + "/es", Lang: "es"}:        1,
		{URL: base + "/en", Lang: "x-default"}: 1,
	}

	got := hreflangSet(en.Hreflangs)
	if len(got) != len(want) {
		t.Errorf("/en hreflangs: got %v, want %v", got, want)
	}

	for h := range want {
		if got[h] != 1 {
			t.Errorf("/en hreflangs: %v found %d times, want 1", h, got[h])
		}
	}

	// All the linked language versions are crawled through the hreflang URLs.
	for _, path := range []string{"/fr", "/de", "/es"} {
		if pageRepo.get(base+path) == nil {
			t.Errorf("no page report for %s%s", base, path)
		}
	}

	// The /de page points its es hreflang to /es-old, which redirects to /es.
	// Both the redirect and its final URL are crawled.
	de := pageRepo.get(base + "/de")
	if de == nil {
		t.Fatalf("no page report for %s/de", base)
	}

	deSet := hreflangSet(de.Hreflangs)
	if deSet[models.Hreflang{URL: base + "/es-old", Lang: "es"}] != 1 {
		t.Errorf("/de hreflangs: missing es hreflang to /es-old, got %v", deSet)
	}

	old := pageRepo.get(base + "/es-old")
	if old == nil {
		t.Fatalf("no page report for %s/es-old", base)
	}

	if old.StatusCode != http.StatusMovedPermanently {
		t.Errorf("/es-old status code: got %d, want %d", old.StatusCode, http.StatusMovedPermanently)
	}

	if old.RedirectURL != base+"/es" {
		t.Errorf("/es-old redirect URL: got %q, want %q", old.RedirectURL, base+"/es")
	}
}

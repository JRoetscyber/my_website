package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) (*fiber.App, *PublicHandler, *gorm.DB) {
	initTestFilters()

	db, err := gorm.Open(sqlite.Open("file:canonical_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	_ = db.AutoMigrate(&models.BlogPost{}, &models.Project{}, &models.FAQ{}, &models.Service{})

	cfg := &config.Config{BaseURL: "https://jo4.co.za"}
	cm := cache.NewManager(cfg)
	handler := NewPublicHandler(db, cfg, cm, nil)

	engine := django.New("../../views", ".html")
	engine.AddFunc("static_url", func(fn string) string { return "/static/" + fn })
	engine.AddFunc("asset_version", func() string { return "v1" })
	engine.AddFunc("asset_url", func(p string) string { return p })
	engine.AddFunc("url_for", func(ep string, args ...string) string { return "/" + ep })

	app := fiber.New(fiber.Config{Views: engine})

	// Trailing slash redirect middleware
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			clean := strings.TrimRight(path, "/")
			query := string(c.Request().URI().QueryString())
			if query != "" {
				clean += "?" + query
			}
			return c.Redirect(clean, fiber.StatusMovedPermanently)
		}
		return c.Next()
	})

	app.Get("/", handler.Home)
	app.Get("/services", handler.Services)
	app.Get("/web-design", handler.WebDesign)
	app.Get("/seo", handler.SEO)
	app.Get("/appsec", handler.AppSec)
	app.Get("/mobile-apps", handler.AppDev)
	app.Get("/app-development", func(c *fiber.Ctx) error {
		return c.Redirect("/mobile-apps", fiber.StatusMovedPermanently)
	})
	app.Get("/automation", handler.Automation)
	app.Get("/projects", handler.Projects)
	app.Get("/projects/:slug", handler.ProjectDetail)
	app.Get("/blog", handler.BlogList)
	app.Get("/blog/:slug", handler.BlogDetail)
	app.Get("/faq", handler.FAQList)
	app.Get("/faq/:slug", handler.FAQDetail)
	app.Get("/book", handler.BookPage)
	app.Get("/webp-converter", handler.WebPConverterPage)
	app.Get("/sitemap.xml", handler.Sitemap)
	app.Get("/robots.txt", handler.Robots)

	return app, handler, db
}

func extractCanonical(html string) string {
	idx := strings.Index(html, `rel="canonical"`)
	if idx == -1 {
		return ""
	}
	start := strings.Index(html[idx:], `href="`)
	if start == -1 {
		return ""
	}
	start += idx + 6
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		return ""
	}
	return html[start : start+end]
}

func TestCanonicalMatchesSitemap(t *testing.T) {
	app, _, db := setupTestApp(t)

	now := time.Now()
	// Seed 1 project, 1 blog post, 1 faq
	db.Create(&models.Project{Title: "Portal", Slug: "client-portal"})
	db.Create(&models.BlogPost{Title: "Go Fiber Scalability", Slug: "go-fiber-scalability", Status: "published", PublishedAt: now.Add(-1 * time.Hour), Content: "Content"})
	db.Create(&models.FAQ{Question: "What is Go?", Slug: "what-is-go", Answer: "Fast", IsPublished: true})

	// 1. Fetch sitemap.xml
	sitemapReq := httptest.NewRequest("GET", "/sitemap.xml", nil)
	sitemapResp, err := app.Test(sitemapReq)
	if err != nil {
		t.Fatalf("failed to fetch sitemap: %v", err)
	}
	sitemapBytes, _ := io.ReadAll(sitemapResp.Body)
	sitemapXML := string(sitemapBytes)

	// Test URL paths that must have identical canonical tags
	testPaths := []struct {
		RequestPath string
		ExpectedURL string
	}{
		{"/", "https://jo4.co.za/"},
		{"/services", "https://jo4.co.za/services"},
		{"/web-design", "https://jo4.co.za/web-design"},
		{"/seo", "https://jo4.co.za/seo"},
		{"/appsec", "https://jo4.co.za/appsec"},
		{"/mobile-apps", "https://jo4.co.za/mobile-apps"},
		{"/automation", "https://jo4.co.za/automation"},
		{"/projects", "https://jo4.co.za/projects"},
		{"/projects/client-portal", "https://jo4.co.za/projects/client-portal"},
		{"/blog", "https://jo4.co.za/blog"},
		{"/blog/go-fiber-scalability", "https://jo4.co.za/blog/go-fiber-scalability"},
		{"/faq", "https://jo4.co.za/faq"},
		{"/faq/what-is-go", "https://jo4.co.za/faq/what-is-go"},
		{"/book", "https://jo4.co.za/book"},
		{"/webp-converter", "https://jo4.co.za/webp-converter"},
	}

	for _, tc := range testPaths {
		t.Run(tc.RequestPath, func(t *testing.T) {
			// Verify expected URL is in sitemap
			locTag := "<loc>" + tc.ExpectedURL + "</loc>"
			if !strings.Contains(sitemapXML, locTag) {
				t.Fatalf("sitemap.xml does NOT contain canonical URL: %s", locTag)
			}

			// Verify HTML canonical tag on the page matches expected URL exactly
			req := httptest.NewRequest("GET", tc.RequestPath, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to GET %s: %v", tc.RequestPath, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 OK for %s, got %d", tc.RequestPath, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			html := string(body)
			canonical := extractCanonical(html)

			if canonical != tc.ExpectedURL {
				t.Errorf("[%s] Canonical mismatch: expected %q, got %q", tc.RequestPath, tc.ExpectedURL, canonical)
			}
		})
	}
}

func TestTrailingSlashAndDuplicateRedirects(t *testing.T) {
	app, _, _ := setupTestApp(t)

	// 1. Trailing slash on subpath -> 301 redirect to non-trailing slash
	req := httptest.NewRequest("GET", "/services/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed test request: %v", err)
	}
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected 301 Moved Permanently for /services/, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/services" {
		t.Errorf("expected Location /services, got %q", loc)
	}

	// 2. /app-development -> 301 redirect to /mobile-apps
	req2 := httptest.NewRequest("GET", "/app-development", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("failed test request: %v", err)
	}
	if resp2.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected 301 for /app-development, got %d", resp2.StatusCode)
	}
	loc2 := resp2.Header.Get("Location")
	if loc2 != "/mobile-apps" {
		t.Errorf("expected Location /mobile-apps, got %q", loc2)
	}
}

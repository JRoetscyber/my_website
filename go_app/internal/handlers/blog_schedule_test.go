package handlers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func TestParsePublishInfo(t *testing.T) {
	now := time.Now()

	// 1. Empty values -> default to published now
	s1, t1 := parsePublishInfo("", "")
	if s1 != "published" {
		t.Errorf("expected published, got %s", s1)
	}
	if t1.After(now.Add(time.Second)) || t1.Before(now.Add(-time.Second)) {
		t.Errorf("expected publish time close to now, got %v", t1)
	}

	// 2. Future date with status="published" -> auto-aligns to scheduled
	futureStr := now.Add(48 * time.Hour).Format("2006-01-02T15:04")
	s2, _ := parsePublishInfo("published", futureStr)
	if s2 != "scheduled" {
		t.Errorf("expected auto-alignment to scheduled, got %s", s2)
	}

	// 3. Past date with status="scheduled" -> auto-aligns to published
	pastStr := now.Add(-48 * time.Hour).Format("2006-01-02T15:04")
	s3, _ := parsePublishInfo("scheduled", pastStr)
	if s3 != "published" {
		t.Errorf("expected auto-alignment to published, got %s", s3)
	}

	// 4. Draft stays draft regardless of date
	s4, _ := parsePublishInfo("draft", futureStr)
	if s4 != "draft" {
		t.Errorf("expected draft, got %s", s4)
	}
}

func TestBlogPublicFiltering(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	_ = db.AutoMigrate(&models.BlogPost{}, &models.Project{}, &models.FAQ{})

	now := time.Now()

	// Seed 4 posts:
	// 1. Published (past)
	// 2. Legacy post (empty status/published_at)
	// 3. Future scheduled
	// 4. Draft
	p1 := models.BlogPost{
		Title:       "Live Post",
		Slug:        "live-post",
		Content:     "Live content",
		Status:      "published",
		PublishedAt: now.Add(-2 * time.Hour),
	}
	p2 := models.BlogPost{
		Title:   "Legacy Post",
		Slug:    "legacy-post",
		Content: "Legacy content",
		// Status and PublishedAt empty
	}
	p3 := models.BlogPost{
		Title:       "Future Post",
		Slug:        "future-post",
		Content:     "Future content",
		Status:      "scheduled",
		PublishedAt: now.Add(24 * time.Hour),
	}
	p4 := models.BlogPost{
		Title:       "Draft Post",
		Slug:        "draft-post",
		Content:     "Draft content",
		Status:      "draft",
		PublishedAt: now.Add(-1 * time.Hour),
	}

	db.Create(&p1)
	db.Create(&p2)
	db.Create(&p3)
	db.Create(&p4)

	// Query public visible posts
	var visible []models.BlogPost
	queryTime := time.Now()
	err = db.Where("status != 'draft' AND (published_at <= ? OR published_at IS NULL OR published_at = '0001-01-01 00:00:00+00:00' OR published_at = '0001-01-01 00:00:00')", queryTime).
		Order("published_at desc, created_at desc").
		Find(&visible).Error
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	if len(visible) != 2 {
		t.Fatalf("expected exactly 2 visible posts (live and legacy), got %d", len(visible))
	}

	for _, v := range visible {
		if v.Slug == "future-post" {
			t.Errorf("future-post should NOT be visible to public")
		}
		if v.Slug == "draft-post" {
			t.Errorf("draft-post should NOT be visible to public")
		}
	}
}

func TestSitemapExcludesScheduledAndDrafts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	_ = db.AutoMigrate(&models.BlogPost{}, &models.Project{}, &models.FAQ{})

	now := time.Now()
	db.Create(&models.BlogPost{
		Title:       "Active Live",
		Slug:        "active-live",
		Content:     "Live",
		Status:      "published",
		PublishedAt: now.Add(-1 * time.Hour),
	})
	db.Create(&models.BlogPost{
		Title:       "Secret Scheduled",
		Slug:        "secret-scheduled",
		Content:     "Secret",
		Status:      "scheduled",
		PublishedAt: now.Add(48 * time.Hour),
	})
	db.Create(&models.BlogPost{
		Title:       "Draft Idea",
		Slug:        "draft-idea",
		Content:     "Draft",
		Status:      "draft",
		PublishedAt: now.Add(-10 * time.Hour),
	})

	cfg := &config.Config{BaseURL: "https://jo4.co.za"}
	cm := cache.NewManager(cfg)
	handler := NewPublicHandler(db, cfg, cm, nil)

	app := fiber.New()
	app.Get("/sitemap.xml", handler.Sitemap)

	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test error: %v", err)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	xmlOutput := string(bodyBytes)

	if !strings.Contains(xmlOutput, "https://jo4.co.za/blog/active-live") {
		t.Errorf("expected active-live in sitemap.xml")
	}
	if strings.Contains(xmlOutput, "secret-scheduled") {
		t.Errorf("scheduled future post leaked into sitemap.xml")
	}
	if strings.Contains(xmlOutput, "draft-idea") {
		t.Errorf("draft post leaked into sitemap.xml")
	}
}

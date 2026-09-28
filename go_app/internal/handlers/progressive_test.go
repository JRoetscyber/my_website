package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&models.Project{},
		&models.BlogPost{},
		&models.Lead{},
		&models.Order{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestProgressiveHandler_Home(t *testing.T) {
	db := setupTestDB(t)
	cm := cache.NewManager(&config.Config{})
	handler := NewProgressiveHandler(db, cm)

	app := fiber.New()
	app.Get("/api/v1/page/home", handler.HomeProgressive)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/page/home", nil)
	resp, err := app.Test(req, 1000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var envelope map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	meta, ok := envelope["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta block missing or invalid: %v", envelope)
	}
	if meta["cached"] != true {
		t.Errorf("expected meta.cached=true, got %v", meta["cached"])
	}

	data, ok := envelope["data"].(map[string]interface{})
	if !ok || data["brand"] != "JO4 Dev" {
		t.Errorf("unexpected data envelope: %v", data)
	}

	subResources, ok := envelope["sub_resources"].([]interface{})
	if !ok || len(subResources) == 0 {
		t.Errorf("expected sub_resources list, got %v", subResources)
	}
}

func TestProgressiveHandler_SubResources(t *testing.T) {
	db := setupTestDB(t)
	cm := cache.NewManager(&config.Config{})
	handler := NewProgressiveHandler(db, cm)

	// Seed test data
	db.Create(&models.Project{Title: "Case 1", Slug: "case-1"})
	db.Create(&models.BlogPost{Title: "Blog 1", Slug: "blog-1"})

	app := fiber.New()
	app.Get("/api/v1/page/home/projects", handler.HomeProjectsSubResource)
	app.Get("/api/v1/page/home/blogs", handler.HomeBlogsSubResource)
	app.Get("/api/v1/dashboard/overview", handler.DashboardOverviewSkeleton)

	// Test Projects Sub-Resource
	req := httptest.NewRequest(http.MethodGet, "/api/v1/page/home/projects", nil)
	resp, err := app.Test(req, 1000)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("projects subresource failed: err=%v, code=%d", err, resp.StatusCode)
	}

	// Test Blogs Sub-Resource
	req = httptest.NewRequest(http.MethodGet, "/api/v1/page/home/blogs", nil)
	resp, err = app.Test(req, 1000)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("blogs subresource failed: err=%v, code=%d", err, resp.StatusCode)
	}

	// Test Dashboard Overview Skeleton
	req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/overview", nil)
	resp, err = app.Test(req, 1000)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("dashboard overview skeleton failed: err=%v, code=%d", err, resp.StatusCode)
	}
}

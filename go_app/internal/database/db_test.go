package database

import (
	"testing"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureSchemaColumns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:test_migration_schema?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// 1. Create table with legacy schema (WITHOUT status and published_at)
	err = db.Exec(`CREATE TABLE blog_posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		slug TEXT UNIQUE,
		summary TEXT,
		content TEXT,
		media_path TEXT,
		views INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	if err != nil {
		t.Fatalf("failed to create old table: %v", err)
	}

	// Insert an existing row
	db.Exec(`INSERT INTO blog_posts (title, slug, content, created_at, updated_at) VALUES ('Old Post', 'old-post', 'Old Content', '2026-01-01 12:00:00', '2026-01-01 12:00:00')`)

	// 2. Call ensureSchemaColumns
	ensureSchemaColumns(db)

	// 3. Verify status was added and backfilled to 'published'
	var status string
	err = db.Raw("SELECT status FROM blog_posts WHERE slug = 'old-post'").Scan(&status).Error
	if err != nil {
		t.Fatalf("failed to query status: %v", err)
	}
	if status != "published" {
		t.Errorf("expected status 'published', got %q", status)
	}

	// 4. Verify published_at was added and backfilled
	var post models.BlogPost
	if err := db.Where("slug = ?", "old-post").First(&post).Error; err != nil {
		t.Fatalf("failed to fetch BlogPost via GORM: %v", err)
	}
	if post.PublishedAt.IsZero() {
		t.Errorf("expected published_at to be backfilled from created_at, got %v", post.PublishedAt)
	}
	if post.Status != "published" {
		t.Errorf("expected post.Status 'published', got %q", post.Status)
	}

	// 5. Calling ensureSchemaColumns a second time must be completely idempotent
	ensureSchemaColumns(db)
}

package database

import (
	"testing"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPageContentCRUDAndReset(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.PageContent{}); err != nil {
		t.Fatalf("failed to migrate page_contents: %v", err)
	}

	SeedPageContents(db)

	home := GetHomeContent(db)
	if home.HeroTitleLine1 == "" || len(home.Pillars) == 0 {
		t.Fatalf("expected seeded home content, got: %+v", home)
	}

	home.HeroTitleLine1 = "Custom Hero From Admin"
	if err := SavePageContent(db, "home", home); err != nil {
		t.Fatalf("SavePageContent failed: %v", err)
	}

	reloaded := GetHomeContent(db)
	if reloaded.HeroTitleLine1 != "Custom Hero From Admin" {
		t.Fatalf("expected updated hero title, got %q", reloaded.HeroTitleLine1)
	}

	if _, err := ResetPageContent(db, "home"); err != nil {
		t.Fatalf("ResetPageContent failed: %v", err)
	}

	reset := GetHomeContent(db)
	if reset.HeroTitleLine1 != models.DefaultHomeContent().HeroTitleLine1 {
		t.Fatalf("expected reset hero title, got %q", reset.HeroTitleLine1)
	}

	if len(GetAboutContent(db).SkillCategories) == 0 {
		t.Fatal("expected seeded about skill categories")
	}
	if len(GetSecurityContent(db).Tools) == 0 {
		t.Fatal("expected seeded security tools")
	}
	if len(GetEngineeringContent(db).Builds) == 0 {
		t.Fatal("expected seeded engineering builds")
	}
}

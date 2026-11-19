package database

import (
	"log"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/goccy/go-json"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func loadPageContent[T any](db *gorm.DB, key string, fallback T) T {
	if db == nil {
		return fallback
	}
	var rec models.PageContent
	if err := db.Where("page_key = ?", key).First(&rec).Error; err != nil || rec.DataJSON == "" {
		return fallback
	}
	result := fallback
	if err := json.Unmarshal([]byte(rec.DataJSON), &result); err != nil {
		log.Printf("[CMS] Failed to unmarshal page_content '%s': %v", key, err)
		return fallback
	}
	return result
}

func GetHomeContent(db *gorm.DB) models.HomeContent {
	return loadPageContent(db, "home", models.DefaultHomeContent())
}

func GetAboutContent(db *gorm.DB) models.AboutContent {
	return loadPageContent(db, "about", models.DefaultAboutContent())
}

func GetSecurityContent(db *gorm.DB) models.SecurityContent {
	return loadPageContent(db, "security", models.DefaultSecurityContent())
}

func GetEngineeringContent(db *gorm.DB) models.EngineeringContent {
	return loadPageContent(db, "engineering", models.DefaultEngineeringContent())
}

func SavePageContent(db *gorm.DB, key string, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	rec := models.PageContent{
		PageKey:  key,
		DataJSON: string(raw),
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "page_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"data_json", "updated_at"}),
	}).Create(&rec).Error
}

func ResetPageContent(db *gorm.DB, key string) (interface{}, error) {
	var def interface{}
	switch key {
	case "home":
		def = models.DefaultHomeContent()
	case "about":
		def = models.DefaultAboutContent()
	case "security":
		def = models.DefaultSecurityContent()
	case "engineering":
		def = models.DefaultEngineeringContent()
	default:
		return nil, gorm.ErrRecordNotFound
	}
	if err := SavePageContent(db, key, def); err != nil {
		return nil, err
	}
	return def, nil
}

func SeedPageContents(db *gorm.DB) {
	defaults := map[string]interface{}{
		"home":        models.DefaultHomeContent(),
		"about":       models.DefaultAboutContent(),
		"security":    models.DefaultSecurityContent(),
		"engineering": models.DefaultEngineeringContent(),
	}
	for key, data := range defaults {
		var count int64
		db.Model(&models.PageContent{}).Where("page_key = ?", key).Count(&count)
		if count == 0 {
			if err := SavePageContent(db, key, data); err != nil {
				log.Printf("[CMS] Failed seeding default content for '%s': %v", key, err)
			} else {
				log.Printf("[CMS] Seeded default editable content for page '%s'", key)
			}
		}
	}
}

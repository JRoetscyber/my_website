package database

import (
	"log"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dbPath := cfg.DatabaseURL
	if strings.HasPrefix(dbPath, "sqlite:///") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite:///")
	} else if strings.HasPrefix(dbPath, "sqlite://") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite://")
	}
	if dbPath == "" {
		dbPath = "jo4dev.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// SQLite Concurrency & Performance Pragmas
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA synchronous = NORMAL;")
	db.Exec("PRAGMA cache_size = -64000;")
	db.Exec("PRAGMA busy_timeout = 5000;")

	// Database Connection Pool Tuning
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(15 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	// Auto Migrate all models
	err = db.AutoMigrate(
		&models.Lead{},
		&models.Project{},
		&models.BlogPost{},
		&models.FAQ{},
		&models.Service{},
		&models.BookingSettings{},
		&models.InvoiceSettings{},
		&models.AutomationLog{},
		&models.Analytics{},
		&models.User{},
		&models.Transaction{},
		&models.FAQSubmission{},
		&models.Order{},
	)
	if err != nil {
		log.Printf("[DB] Migration warning: %v", err)
	}

	DB = db
	SeedDefaults(db, cfg)

	log.Printf("[DB] Database connected and migrated successfully (%s)", dbPath)
	return db, nil
}

func SeedDefaults(db *gorm.DB, cfg *config.Config) {
	// Seed Booking Settings
	var bookingCount int64
	db.Model(&models.BookingSettings{}).Count(&bookingCount)
	if bookingCount == 0 {
		db.Create(&models.BookingSettings{
			CalendarID:             "primary",
			WorkdayStart:           "09:00",
			WorkdayEnd:             "17:00",
			MeetingDurationMinutes: 30,
			BufferMinutes:          30,
			SlotStepMinutes:        30,
			BookingHorizonDays:     21,
			MinNoticeHours:         4,
			ReminderMinutes:        30,
			CreateGoogleMeet:       true,
			MeetingLocation:        "Google Meet",
		})
	}

	// Seed Invoice Settings
	var invoiceCount int64
	db.Model(&models.InvoiceSettings{}).Count(&invoiceCount)
	if invoiceCount == 0 {
		db.Create(&models.InvoiceSettings{
			BizName:      "JO4 Dev",
			BizEmail:     "jroetscyber@gmail.com",
			BizPhone:     "+27 82 000 0000",
			BizAddress:   "Cape Town, South Africa",
			PaymentTerms: "Payment due within 14 days of invoice date.",
		})
	}

	// Ensure Admin User from environment configuration is always synchronized and active
	if cfg.DefaultAdminUser != "" && cfg.DefaultAdminPass != "" {
		var adminUser models.User
		err := db.Where("LOWER(username) = LOWER(?)", cfg.DefaultAdminUser).First(&adminUser).Error
		if err != nil {
			// Admin user does not exist: create it with password from environment
			hash, err := utils.HashPassword(cfg.DefaultAdminPass)
			if err == nil {
				db.Create(&models.User{
					Username:     cfg.DefaultAdminUser,
					PasswordHash: hash,
				})
				log.Printf("[DB] Created admin user '%s' from environment config", cfg.DefaultAdminUser)
			}
		} else {
			// Admin user exists: ensure the password from environment config matches
			if !utils.CheckPasswordHash(adminUser.PasswordHash, cfg.DefaultAdminPass) {
				hash, err := utils.HashPassword(cfg.DefaultAdminPass)
				if err == nil {
					db.Model(&adminUser).Update("password_hash", hash)
					log.Printf("[DB] Synchronized password hash for admin user '%s' from environment config", cfg.DefaultAdminUser)
				}
			}
		}
	}

	// Seed Core Services
	var serviceCount int64
	db.Model(&models.Service{}).Count(&serviceCount)
	if serviceCount == 0 {
		services := []models.Service{
			{
				Title:            "Web Development & Business Tools",
				Slug:             "web-design",
				Eyebrow:          "Service 01 — Core Engine",
				LeadText:         "Powered by Dragon and Go Fiber to build the fastest possible websites and bespoke business tools. We automate redundant tasks, skyrocket productivity, and reject cookie-cutter templates.",
				Description:      "We engineer custom high-throughput web systems and internal productivity tools in Go Fiber and the Dragon engine. We don't use cookie-cutter templates because they cannot set you apart from your competition. Every system is built from the ground up to eliminate repetitive manual work, streamline operations, and load in milliseconds.",
				Features:         "Go Fiber & Dragon ultra-high-speed backend\nZero cookie-cutter templates — 100% custom-crafted architecture\nRedundant task automation & internal business workflow tools\nSub-second TTFB (Time To First Byte) and perfect Core Web Vitals\nFull source code ownership — zero platform lock-in\nDatabase indexing & enterprise concurrency with Go goroutines",
				PriceRange:       "Custom Scoped",
				PriceLabel:       "Fixed-price quote",
				PriceNote:        "Engineered for high ROI and operational speed",
				IconSVG:          `<polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5"/><line x1="12" y1="22" x2="12" y2="15.5"/><polyline points="22 8.5 12 15.5 2 8.5"/>`,
				PanelTitle:       "What we build",
				PanelType:        "use-case",
				PanelContent:     `[{"title":"Custom High-Speed Websites","desc":"Lightning-fast web platforms built with Go Fiber that keep visitors engaged and convert traffic.","icon":"monitor"},{"title":"Business Automation Tools","desc":"Eliminate redundant tasks, sync spreadsheets and databases, and give hours back to your team.","icon":"cpu"},{"title":"Client Portals & Dashboards","desc":"Secure customer management systems with real-time updates and zero external SaaS dependencies.","icon":"shield"}]`,
				IsPublished:      true,
				HasDedicatedPage: true,
				DisplayOrder:     10,
			},
			{
				Title:            "Search Engine Optimization (SEO)",
				Slug:             "seo",
				Eyebrow:          "Service 02 — Organic Growth",
				LeadText:         "Technical SEO built directly into compiled code. We build sites that search engines love crawling, giving you an unfair advantage in organic ranking.",
				Description:      "A website nobody can find is useless. Because our systems are built on Go Fiber with clean semantic markup and zero script bloat, search engine bots index pages instantly. We combine technical Core Web Vitals optimization with local keyword dominance and structured schema data.",
				Features:         "Technical SEO baked into the compiled application\nInstant crawlability — sub-300ms server response times\nStructured JSON-LD schema markup & OpenGraph integrations\nLocal search dominance & Google Business Profile alignment\nIn-depth competitor keyword gap analysis & ranking tracking",
				PriceRange:       "Custom Scoped",
				PriceLabel:       "Audit & Implementation",
				PriceNote:        "Proven organic rankings that convert",
				IconSVG:          `<circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><path d="M11 8v6M8 11h6"/>`,
				PanelTitle:       "SEO Pillars",
				PanelType:        "audit",
				PanelContent:     `[{"num":"01","title":"Speed & Core Web Vitals","desc":"Near-instant page loads that Google's algorithm prioritizes over bloated WordPress sites."},{"num":"02","title":"Semantic Indexing","desc":"Clean HTML5 structure, automated sitemaps, and robots configuration for fast indexing."},{"num":"03","title":"Local Market Dominance","desc":"Location-based targeting and search optimization to capture high-intent buyers."},{"num":"04","title":"Data-Driven Tracking","desc":"Google Search Console, analytics, and rank verification to track real business leads."}]`,
				IsPublished:      true,
				HasDedicatedPage: true,
				DisplayOrder:     20,
			},
			{
				Title:            "Mobile App Development",
				Slug:             "mobile-apps",
				Eyebrow:          "Service 03 — Native Performance",
				LeadText:         "Native iOS and Android engineering in C++, Kotlin, and Swift. True native performance, fluid touch response, and rock-solid reliability.",
				Description:      "When hybrid web wrappers aren't enough, we build real native mobile applications. Using C++ for ultra-fast computational engines and high-performance algorithms, Kotlin for modern Android architecture, and Swift for iOS elegance, we deliver mobile tools that feel instant and never stutter.",
				Features:         "Native Android development with Kotlin & Jetpack Compose\nNative iOS development with Swift & SwiftUI\nHigh-performance core logic and algorithms in C++\nSeamless offline synchronization and local caching\nEnd-to-end App Store & Google Play deployment\nIntegration with your Go Fiber backend APIs",
				PriceRange:       "Custom Scoped",
				PriceLabel:       "Milestone-based",
				PriceNote:        "Native code you own completely",
				IconSVG:          `<rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/>`,
				PanelTitle:       "Mobile Engineering Languages",
				PanelType:        "use-case",
				PanelContent:     `[{"title":"C++ High-Performance Core","desc":"Maximum computational throughput, custom algorithms, and embedded speed.","icon":"cpu"},{"title":"Kotlin for Android","desc":"Modern, reactive Android apps adhering to Google Material Design standards.","icon":"smartphone"},{"title":"Swift for iOS","desc":"Fluid, high-framerate Apple ecosystem apps for iPhone and iPad.","icon":"tablet"}]`,
				IsPublished:      true,
				HasDedicatedPage: true,
				DisplayOrder:     30,
			},
			{
				Title:            "DevSecOps, Web Security & Hosting",
				Slug:             "appsec",
				Eyebrow:          "Service 04 — Battle-Ready Ops",
				LeadText:         "Blue/Green deployments, Docker containers, Nginx routing, Cloudflare Tunnels, and Hostinger high-performance servers so you stay battle-ready.",
				Description:      "We take web security and uptime seriously. We deploy systems using Docker isolation, Nginx reverse proxy routing, and encrypted Cloudflare Tunnels with zero open ingress ports. Utilizing Blue/Green (Red) zero-downtime deployments hosted on Hostinger high-performance servers, your business remains resilient, secure, and battle-ready at all times.",
				Features:         "Blue/Green (Red) zero-downtime deployment pipelines\nDocker containerized architecture for reliable, repeatable environments\nNginx reverse proxy routing with hardened HTTP security headers\nCloudflare Tunnels — origin server IPs stay 100% hidden and secure\nHardened against OWASP Top 10 vulnerabilities & automated bot attacks\nHosted on high-performance Hostinger servers for 24/7 battle readiness",
				PriceRange:       "Custom Scoped",
				PriceLabel:       "Infrastructure Setup & Management",
				PriceNote:        "Military-grade security posture",
				IconSVG:          `<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>`,
				PanelTitle:       "Security & Deployment Stack",
				PanelType:        "stats",
				PanelContent:     `[{"num":"0","suffix":"ms","desc":"downtime with our Blue/Green rolling release deployment strategies"},{"num":"100","suffix":"%","desc":"origin server protection with Cloudflare Tunnels and isolated Nginx routing"},{"num":"24","suffix":"/7","desc":"battle-ready monitoring on high-performance Hostinger dedicated environments"}]`,
				IsPublished:      true,
				HasDedicatedPage: true,
				DisplayOrder:     40,
			},
		}

		for _, s := range services {
			db.Create(&s)
		}
		log.Println("[DB] Seeded 4 core services (Web Dev, SEO, Mobile Apps, DevSecOps)")
	}
}

func GetBookingSettings(db *gorm.DB) *models.BookingSettings {
	var settings models.BookingSettings
	if err := db.First(&settings).Error; err != nil {
		settings = models.BookingSettings{
			CalendarID:             "primary",
			WorkdayStart:           "09:00",
			WorkdayEnd:             "17:00",
			MeetingDurationMinutes: 30,
			BufferMinutes:          30,
			SlotStepMinutes:        30,
			BookingHorizonDays:     21,
			MinNoticeHours:         4,
		}
		db.Create(&settings)
	}
	return &settings
}

func GetInvoiceSettings(db *gorm.DB) *models.InvoiceSettings {
	var settings models.InvoiceSettings
	if err := db.First(&settings).Error; err != nil {
		settings = models.InvoiceSettings{
			BizName:  "JO4 Dev",
			BizEmail: "jroetscyber@gmail.com",
		}
		db.Create(&settings)
	}
	return &settings
}

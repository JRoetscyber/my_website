package database

import (
	"log"
	"strings"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
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

	// Seed Default Admin User if no users exist
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPass), bcrypt.DefaultCost)
		if err == nil {
			db.Create(&models.User{
				Username:     cfg.DefaultAdminUser,
				PasswordHash: string(hash),
			})
			log.Printf("[DB] Created default admin user: %s", cfg.DefaultAdminUser)
		}
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

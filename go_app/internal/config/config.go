package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	SecretKey           string
	Environment         string
	MailServer          string
	MailPort            string
	MailUsername        string
	MailPassword        string
	MailSender          string
	GoogleCalendarFile  string
	GoogleCalendarID    string
	DefaultAdminUser    string
	DefaultAdminPass    string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	port := getEnv("PORT", "5000")
	dbURL := getEnv("SQLALCHEMY_DATABASE_URI", "")
	if dbURL == "" {
		dbURL = getEnv("DATABASE_URL", "jo4dev.db")
	}

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		SecretKey:          getEnv("SECRET_KEY", "dev-secret-key-123"),
		Environment:        getEnv("ENV", "development"),
		MailServer:         getEnv("MAIL_SERVER", "smtp-relay.brevo.com"),
		MailPort:           getEnv("MAIL_PORT", "587"),
		MailUsername:       getEnv("MAIL_USERNAME", ""),
		MailPassword:       getEnv("MAIL_PASSWORD", ""),
		MailSender:         getEnv("MAIL_DEFAULT_SENDER", "info@jo4.co.za"),
		GoogleCalendarFile: getEnv("GOOGLE_CALENDAR_SERVICE_ACCOUNT_FILE", ""),
		GoogleCalendarID:   getEnv("GOOGLE_CALENDAR_ID", "primary"),
		DefaultAdminUser:   getEnv("ADMIN_USERNAME", "admin"),
		DefaultAdminPass:   getEnv("ADMIN_PASSWORD", "admin123!"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

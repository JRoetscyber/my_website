package middleware

import (
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func TrackAnalytics(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Skip tracking for static files, api routes, admin routes, and assets
		if strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/favicon") ||
			strings.HasPrefix(path, "/robots") ||
			strings.HasPrefix(path, "/sitemap") {
			return c.Next()
		}

		ip := c.IP()

		// Log visitor asynchronously so it never slows down page response
		go func(p, visitorIP string) {
			db.Create(&models.Analytics{
				PagePath:  p,
				VisitorIP: visitorIP,
				Timestamp: time.Now(),
			})
		}(path, ip)

		return c.Next()
	}
}

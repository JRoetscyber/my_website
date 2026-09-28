package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/worker"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TrackAnalytics logs visitor metrics asynchronously using a bounded worker pool.
// CRITICAL FIBER SAFETY: Extracts and clones string values (strings.Clone) to eliminate
// fasthttp buffer reuse race conditions.
func TrackAnalytics(db *gorm.DB, pool *worker.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Skip tracking for static assets, administrative endpoints, and meta files
		if strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/favicon") ||
			strings.HasPrefix(path, "/robots") ||
			strings.HasPrefix(path, "/sitemap") ||
			path == "/health" {
			return c.Next()
		}

		// FASTHTTP BUFFER SAFETY: clone strings to decouple from fasthttp request byte buffer pool
		clonedPath := strings.Clone(path)
		clonedIP := strings.Clone(c.IP())
		now := time.Now()

		// Enqueue non-blocking task into bounded worker pool
		pool.Enqueue(func(ctx context.Context) {
			db.WithContext(ctx).Create(&models.Analytics{
				PagePath:  clonedPath,
				VisitorIP: clonedIP,
				Timestamp: now,
			})
		})

		return c.Next()
	}
}

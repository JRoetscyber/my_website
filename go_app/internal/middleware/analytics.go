package middleware

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/worker"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// analyticsSampleRate controls what fraction of requests are tracked.
// 0.20 = 20% sampling: accurate enough for trend analysis, 5× less SQLite write pressure.
// Raise to 1.0 for 100% tracking (fine for low-traffic sites, degrades under heavy load).
const analyticsSampleRate = 0.20

// TrackAnalytics logs visitor metrics asynchronously using a bounded worker pool.
// CRITICAL FIBER SAFETY: Extracts and clones string values (strings.Clone) to eliminate
// fasthttp buffer reuse race conditions.
// PERFORMANCE: Probabilistic sampling caps SQLite write rate under high concurrency.
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

		// Probabilistic sampling: only track a fraction of requests to cap SQLite write rate.
		// This keeps analytics meaningful without becoming a bottleneck under high concurrency.
		if rand.Float64() > analyticsSampleRate {
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


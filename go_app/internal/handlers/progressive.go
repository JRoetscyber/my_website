package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ResponseMeta standardizes progressive API envelope shapes for frontend skeleton loading
type ResponseMeta struct {
	Timestamp   time.Time `json:"timestamp"`
	Cached      bool      `json:"cached"`
	ETag        string    `json:"etag,omitempty"`
	IsSkeleton  bool      `json:"is_skeleton"`
	DurationMs  int64     `json:"duration_ms"`
}

type SubResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ProgressiveEnvelope[T any] struct {
	Meta         ResponseMeta  `json:"meta"`
	Data         T             `json:"data"`
	SubResources []SubResource `json:"sub_resources,omitempty"`
}

type ProgressiveHandler struct {
	DB    *gorm.DB
	Cache *cache.Manager
}

func NewProgressiveHandler(db *gorm.DB, cm *cache.Manager) *ProgressiveHandler {
	return &ProgressiveHandler{
		DB:    db,
		Cache: cm,
	}
}

// HomeProgressive provides an instantaneous skeleton payload (<1ms) for the homepage,
// allowing client apps and web frontends to resolve layouts immediately while streaming
// or fetching heavier nested sub-resources concurrently.
func (h *ProgressiveHandler) HomeProgressive(c *fiber.Ctx) error {
	start := time.Now()
	ctx := context.Background()
	cacheKey := "page:home:skeleton:v1"

	type HomeSkeletonData struct {
		Brand       string   `json:"brand"`
		Tagline     string   `json:"tagline"`
		TechStack   []string `json:"tech_stack"`
		HeroMetrics []string `json:"hero_metrics"`
	}

	data, err := cache.GetOrSet(h.Cache, ctx, cacheKey, 1*time.Hour, func(ctx context.Context) (HomeSkeletonData, error) {
		return HomeSkeletonData{
			Brand:   "JO4 Dev",
			Tagline: "Engineered in Dragon & Go Fiber. High-speed web systems and native apps.",
			TechStack: []string{
				"Go Fiber (fasthttp)",
				"C++ Core Engine",
				"Kotlin Android",
				"Swift iOS",
				"Redis Cache-Aside",
			},
			HeroMetrics: []string{
				"120 FPS Fluid ProMotion",
				"<50ms Direct Touch Latency",
				"0% Hybrid Bloat",
				"100% Native Code Ownership",
			},
		}, nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	etag := cache.GenerateETag([]byte(fmt.Sprintf("%v", data)))
	if cache.HandleETag(c, etag) {
		return nil
	}

	cache.SetCacheHeaders(c, 10*time.Minute, etag)

	return c.JSON(ProgressiveEnvelope[HomeSkeletonData]{
		Meta: ResponseMeta{
			Timestamp:  time.Now(),
			Cached:     true,
			ETag:       etag,
			IsSkeleton: false,
			DurationMs: time.Since(start).Milliseconds(),
		},
		Data: data,
		SubResources: []SubResource{
			{Name: "recent_projects", URL: "/api/v1/page/home/projects"},
			{Name: "recent_blogs", URL: "/api/v1/page/home/blogs"},
			{Name: "services_deck", URL: "/api/v1/services"},
		},
	})
}

// HomeProjectsSubResource returns the 3 featured project cards for the homepage skeleton
func (h *ProgressiveHandler) HomeProjectsSubResource(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := "page:home:projects:v1"

	projects, err := cache.GetOrSet(h.Cache, ctx, cacheKey, 15*time.Minute, func(ctx context.Context) ([]models.Project, error) {
		var list []models.Project
		err := h.DB.WithContext(ctx).Order("deployed_at desc, id desc").Limit(3).Find(&list).Error
		return list, err
	})

	if err != nil {
		// Return empty envelope rather than failing the whole page
		return c.JSON(ProgressiveEnvelope[[]models.Project]{
			Meta: ResponseMeta{Timestamp: time.Now(), IsSkeleton: true},
			Data: []models.Project{},
		})
	}

	etag := cache.GenerateETag([]byte(fmt.Sprintf("%v", projects)))
	if cache.HandleETag(c, etag) {
		return nil
	}

	cache.SetCacheHeaders(c, 15*time.Minute, etag)

	return c.JSON(ProgressiveEnvelope[[]models.Project]{
		Meta: ResponseMeta{
			Timestamp:  time.Now(),
			Cached:     true,
			ETag:       etag,
			IsSkeleton: false,
		},
		Data: projects,
	})
}

// HomeBlogsSubResource returns the 3 recent blogs for the homepage skeleton
func (h *ProgressiveHandler) HomeBlogsSubResource(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := "page:home:blogs:v1"

	blogs, err := cache.GetOrSet(h.Cache, ctx, cacheKey, 15*time.Minute, func(ctx context.Context) ([]models.BlogPost, error) {
		var list []models.BlogPost
		err := h.DB.WithContext(ctx).Order("created_at desc").Limit(3).Find(&list).Error
		return list, err
	})

	if err != nil {
		return c.JSON(ProgressiveEnvelope[[]models.BlogPost]{
			Meta: ResponseMeta{Timestamp: time.Now(), IsSkeleton: true},
			Data: []models.BlogPost{},
		})
	}

	etag := cache.GenerateETag([]byte(fmt.Sprintf("%v", blogs)))
	if cache.HandleETag(c, etag) {
		return nil
	}

	cache.SetCacheHeaders(c, 15*time.Minute, etag)

	return c.JSON(ProgressiveEnvelope[[]models.BlogPost]{
		Meta: ResponseMeta{
			Timestamp:  time.Now(),
			Cached:     true,
			ETag:       etag,
			IsSkeleton: false,
		},
		Data: blogs,
	})
}

// DashboardOverviewSkeleton provides instant summary badges for admin dashboards,
// deferring heavy time-series analytics charts to /api/v1/analytics/timeseries.
func (h *ProgressiveHandler) DashboardOverviewSkeleton(c *fiber.Ctx) error {
	ctx := context.Background()
	cacheKey := cache.DashboardOverviewKey()

	type OverviewStats struct {
		LeadsCount      int64  `json:"leads_count"`
		ProjectsCount   int64  `json:"projects_count"`
		BlogsCount      int64  `json:"blogs_count"`
		TotalBlogViews  int64  `json:"total_blog_views"`
		OrdersPreparing int64  `json:"orders_preparing"`
		OrdersReady     int64  `json:"orders_ready"`
		RedisActive     bool   `json:"redis_active"`
		SystemHealth    string `json:"system_health"`
	}

	stats, err := cache.GetOrSet(h.Cache, ctx, cacheKey, 30*time.Second, func(ctx context.Context) (OverviewStats, error) {
		var leads, projects, blogs, blogViews, prep, ready int64
		h.DB.WithContext(ctx).Model(&models.Lead{}).Count(&leads)
		h.DB.WithContext(ctx).Model(&models.Project{}).Count(&projects)
		h.DB.WithContext(ctx).Model(&models.BlogPost{}).Count(&blogs)
		h.DB.WithContext(ctx).Model(&models.BlogPost{}).Select("COALESCE(SUM(views), 0)").Scan(&blogViews)
		h.DB.WithContext(ctx).Model(&models.Order{}).Where("status IN ('received', 'in_progress')").Count(&prep)
		h.DB.WithContext(ctx).Model(&models.Order{}).Where("status = 'ready'").Count(&ready)

		return OverviewStats{
			LeadsCount:      leads,
			ProjectsCount:   projects,
			BlogsCount:      blogs,
			TotalBlogViews:  blogViews,
			OrdersPreparing: prep,
			OrdersReady:     ready,
			RedisActive:     h.Cache.IsRedisAvailable(),
			SystemHealth:    "optimal",
		}, nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ProgressiveEnvelope[OverviewStats]{
		Meta: ResponseMeta{
			Timestamp:  time.Now(),
			Cached:     true,
			IsSkeleton: false,
		},
		Data: stats,
		SubResources: []SubResource{
			{Name: "timeseries_chart", URL: "/admin/analytics_data"},
			{Name: "leads_paged", URL: "/admin/leads"},
		},
	})
}

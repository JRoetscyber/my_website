package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/database"
	"github.com/JRoetscyber/my_website/go_app/internal/handlers"
	"github.com/JRoetscyber/my_website/go_app/internal/middleware"
	"github.com/JRoetscyber/my_website/go_app/internal/worker"
	"github.com/flosch/pongo2/v6"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	fiberCache "github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/django/v3"
	"github.com/yuin/goldmark"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// Zero-allocation buffer pool for Goldmark markdown parsing
var mdBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var routeMap = map[string]string{
	"home":                         "/",
	"services.services":            "/services",
	"services.web_design":          "/web-design",
	"services.seo_services":        "/seo",
	"services.automation":          "/automation",
	"services.mobile_apps":         "/mobile-apps",
	"services.service_detail":      "/services",
	"portfolio.projects":           "/projects",
	"portfolio.project_detail":     "/projects",
	"blog.blog_list":               "/blog",
	"blog.blog_detail":             "/blog",
	"faq.faq_list":                 "/faq",
	"faq.faq_detail":               "/faq",
	"booking.book":                 "/book",
	"booking.booking_availability": "/api/booking-availability",
	"booking.book_call":            "/api/book-call",
	"admin.dashboard":              "/admin",
	"admin.leads_page":             "/admin/leads",
	"admin.analytics_page":         "/admin/analytics",
	"admin.projects_page":          "/admin/projects",
	"admin.blogs_page":             "/admin/blogs",
	"admin.faqs_page":              "/admin/faqs",
	"admin.services_page":          "/admin/services",
	"admin.booking_settings_page":  "/admin/booking",
	"admin.invoices_page":          "/admin/invoices",
	"admin.funds_page":             "/admin/funds",
	"admin.automation_page":        "/admin/automation",
	"login":                        "/login",
	"logout":                       "/logout",
}

// AssetVersion is appended to static asset URLs for deterministic cache-busting
const AssetVersion = "20260929v2"

func initPongo2Filters() {
	// Register markdown filter using sync.Pool
	pongo2.RegisterFilter("markdown", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		buf := mdBufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer mdBufferPool.Put(buf)

		if err := goldmark.Convert([]byte(in.String()), buf); err != nil {
			return pongo2.AsSafeValue(in.String()), nil
		}
		return pongo2.AsSafeValue(buf.String()), nil
	})

	// Register from_json filter
	pongo2.RegisterFilter("from_json", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		raw := in.String()
		if strings.TrimSpace(raw) == "" {
			return pongo2.AsValue([]interface{}{}), nil
		}
		var res interface{}
		if err := json.Unmarshal([]byte(raw), &res); err != nil {
			return pongo2.AsValue([]interface{}{}), nil
		}
		return pongo2.AsValue(res), nil
	})

	// Register truncate filter
	pongo2.RegisterFilter("truncate", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		length := param.Integer()
		if length <= 0 {
			length = 150
		}
		s := in.String()
		runes := []rune(s)
		if len(runes) > length {
			return pongo2.AsValue(string(runes[:length]) + "..."), nil
		}
		return pongo2.AsValue(s), nil
	})

	// Register striptags filter
	pongo2.RegisterFilter("striptags", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		cleaned := htmlTagRegex.ReplaceAllString(in.String(), "")
		return pongo2.AsValue(cleaned), nil
	})

	// Register splitlines filter
	pongo2.RegisterFilter("splitlines", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		lines := strings.Split(strings.ReplaceAll(in.String(), "\r\n", "\n"), "\n")
		var res []string
		for _, l := range lines {
			if tr := strings.TrimSpace(l); tr != "" {
				res = append(res, tr)
			}
		}
		return pongo2.AsValue(res), nil
	})

	// Register tojson filter
	pongo2.RegisterFilter("tojson", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		b, err := json.Marshal(in.Interface())
		if err != nil {
			return pongo2.AsValue("{}"), nil
		}
		return pongo2.AsSafeValue(string(b)), nil
	})

	// Register split filter (auto-trims whitespace and ignores empty items)
	pongo2.RegisterFilter("split", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		sep := param.String()
		if sep == "" {
			sep = ","
		}
		rawParts := strings.Split(in.String(), sep)
		var res []string
		for _, p := range rawParts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				res = append(res, trimmed)
			}
		}
		return pongo2.AsValue(res), nil
	})

	// Register strip and trim filters
	pongo2.RegisterFilter("strip", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(strings.TrimSpace(in.String())), nil
	})
	pongo2.RegisterFilter("trim", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(strings.TrimSpace(in.String())), nil
	})

	// Register format filter
	pongo2.RegisterFilter("format", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(fmt.Sprintf(in.String(), param.Interface())), nil
	})

	// Register selectattr filter
	pongo2.RegisterFilter("selectattr", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return in, nil
	})

	// Register forceescape filter
	pongo2.RegisterFilter("forceescape", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(pongo2.MustApplyFilter("escape", in, nil)), nil
	})

	// Global helpers
	pongo2.Globals["asset_version"] = AssetVersion
	pongo2.Globals["static_url"] = func(filename string) string {
		clean := "/static/" + strings.TrimPrefix(filename, "/")
		if strings.Contains(clean, "?") {
			return clean + "&v=" + AssetVersion
		}
		return clean + "?v=" + AssetVersion
	}

	pongo2.Globals["asset_url"] = func(path string) string {
		return path
	}

	// Flask-compatible url_for helper
	routeMap := map[string]string{
		"home":                       "/",
		"services.services":          "/services",
		"services.web_design":        "/web-design",
		"services.seo_services":      "/seo",
		"services.automation":        "/automation",
		"services.service_detail":    "/services",
		"portfolio.projects":         "/projects",
		"portfolio.project_detail":   "/projects",
		"blog.blog_list":             "/blog",
		"blog.blog_detail":           "/blog",
		"faq.faq_list":               "/faq",
		"faq.faq_detail":             "/faq",
		"booking.book":               "/book",
		"booking.booking_availability": "/api/booking-availability",
		"booking.book_call":          "/api/book-call",
		"admin.dashboard":            "/admin",
		"admin.leads_page":           "/admin/leads",
		"admin.analytics_page":       "/admin/analytics",
		"admin.projects_page":        "/admin/projects",
		"admin.blogs_page":           "/admin/blogs",
		"admin.faqs_page":            "/admin/faqs",
		"admin.services_page":        "/admin/services",
		"admin.booking_settings_page": "/admin/booking",
		"admin.invoices_page":        "/admin/invoices",
		"admin.funds_page":           "/admin/funds",
		"admin.automation_page":      "/admin/automation",
		"login":                      "/login",
		"logout":                     "/logout",
	}

	pongo2.Globals["url_for"] = func(endpoint string, args ...string) string {
		if path, ok := routeMap[endpoint]; ok {
			return path
		}
		return "/" + strings.TrimPrefix(endpoint, "/")
	}
}

func main() {
	cfg := config.LoadConfig()

	// Initialize Database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Initialize Pongo2 template engine
	initPongo2Filters()

	// Locate templates directory
	templatesDir := os.Getenv("TEMPLATES_DIR")
	if templatesDir == "" {
		if _, err := os.Stat("./views"); err == nil {
			templatesDir = "./views"
		} else if _, err := os.Stat("./go_app/views"); err == nil {
			templatesDir = "./go_app/views"
		} else if _, err := os.Stat("../views"); err == nil {
			templatesDir = "../views"
		} else {
			templatesDir = "./views"
		}
	}
	engine := django.New(templatesDir, ".html")
	engine.Reload(cfg.Environment == "development")

	// Register functions on the engine instance
	engine.AddFunc("static_url", func(filename string) string {
		clean := "/static/" + strings.TrimPrefix(filename, "/")
		if strings.Contains(clean, "?") {
			return clean + "&v=" + AssetVersion
		}
		return clean + "?v=" + AssetVersion
	})
	engine.AddFunc("asset_version", func() string {
		return AssetVersion
	})
	engine.AddFunc("asset_url", func(path string) string {
		return path
	})
	engine.AddFunc("url_for", func(endpoint string, args ...string) string {
		if path, ok := routeMap[endpoint]; ok {
			return path
		}
		return "/" + strings.TrimPrefix(endpoint, "/")
	})

	app := fiber.New(fiber.Config{
		Views:        engine,
		AppName:      "JO4 Dev High-Performance Go Fiber Server",
		BodyLimit:    32 * 1024 * 1024, // 32MB max upload
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ProxyHeader:  fiber.HeaderXForwardedFor,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	// Initialize Cache-Aside Manager with Redis & Singleflight Thundering Herd Protection
	cacheManager := cache.NewManager(cfg)

	// Initialize Bounded Worker Pool for non-blocking asynchronous operations
	workerPool := worker.NewPool(16, 2048, 10*time.Second)

	// Global Middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		CrossOriginEmbedderPolicy: "unsafe-none",
		CrossOriginOpenerPolicy:   "same-origin-allow-popups",
		CrossOriginResourcePolicy: "cross-origin",
	}))
	app.Use(cors.New())
	app.Use(middleware.TrackAnalytics(db, workerPool))

	// In-memory page cache for lightning-fast Core Web Vitals (<1ms TTFB, zero DB load for crawlers)
	app.Use(fiberCache.New(fiberCache.Config{
		Next: func(c *fiber.Ctx) bool {
			if c.Method() != fiber.MethodGet {
				return true
			}
			path := c.Path()
			if strings.HasPrefix(path, "/admin") ||
				strings.HasPrefix(path, "/login") ||
				strings.HasPrefix(path, "/logout") ||
				strings.HasPrefix(path, "/api") ||
				strings.HasPrefix(path, "/order") ||
				strings.HasPrefix(path, "/track") ||
				path == "/health" {
				return true
			}
			return false
		},
		Expiration:   10 * time.Minute,
		CacheControl: true,
	}))

	// Rate limiter for API routes
	apiLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
	})

	// Brute-force rate limiter for Login (20 attempts per minute per IP)
	loginLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).Render("login", fiber.Map{
				"error":   "Too many failed login attempts. Please wait 1 minute before trying again.",
				"request": c,
			})
		},
	})

	// Static Assets
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		if _, err := os.Stat("./static"); err == nil {
			staticDir = "./static"
		} else if _, err := os.Stat("../static"); err == nil {
			staticDir = "../static"
		} else {
			staticDir = "./static"
		}
	}
	app.Static("/static", staticDir, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 7 * 24 * time.Hour,
	})

	// Self-hosted fonts route (1 year immutable cache)
	app.Static("/fonts", filepath.Join(staticDir, "fonts"), fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 365 * 24 * time.Hour,
	})

	// Favicon and root assets
	faviconPath := filepath.Join(filepath.Dir(staticDir), "favicon.ico")
	if !fileExists(faviconPath) {
		faviconPath = filepath.Join(staticDir, "favicon.ico")
	}
	if fileExists(faviconPath) {
		app.Static("/favicon.ico", faviconPath, fiber.Static{
			CacheDuration: 30 * 24 * time.Hour,
		})
	}

	appleIcon := filepath.Join(filepath.Dir(staticDir), "apple-touch-icon.png")
	if !fileExists(appleIcon) {
		appleIcon = filepath.Join(staticDir, "apple-touch-icon.png")
	}
	if fileExists(appleIcon) {
		app.Static("/apple-touch-icon.png", appleIcon, fiber.Static{
			CacheDuration: 30 * 24 * time.Hour,
		})
	}

	// Handlers
	publicHandler := handlers.NewPublicHandler(db, cfg, cacheManager, workerPool)
	authHandler := handlers.NewAuthHandler(db, cfg)
	adminHandler := handlers.NewAdminHandler(db, cfg, cacheManager, workerPool)
	orderHandler := handlers.NewOrderHandler(db, cfg, adminHandler)
	progressiveHandler := handlers.NewProgressiveHandler(db, cacheManager)

	// Public Routes
	app.Get("/", publicHandler.Home)
	app.Get("/services", publicHandler.Services)
	app.Get("/web-design", publicHandler.WebDesign)
	app.Get("/seo", publicHandler.SEO)
	app.Get("/appsec", publicHandler.AppSec)
	app.Get("/mobile-apps", publicHandler.AppDev)
	app.Get("/app-development", publicHandler.AppDev)
	app.Get("/automation", publicHandler.Automation)
	app.Get("/projects", publicHandler.Projects)
	app.Get("/projects/:slug", publicHandler.ProjectDetail)
	app.Get("/blog", publicHandler.BlogList)
	app.Get("/blog/:slug", publicHandler.BlogDetail)
	app.Get("/faq", publicHandler.FAQList)
	app.Get("/faq/:slug", publicHandler.FAQDetail)
	app.Post("/faq/submit", publicHandler.SubmitFAQ)
	app.Get("/book", publicHandler.BookPage)
	app.Get("/webp-converter", publicHandler.WebPConverterPage)
	app.Get("/order/:code", orderHandler.OrderTracker)
	app.Get("/track/:code", orderHandler.OrderTracker)
	app.Get("/orders/display", orderHandler.OrderDisplay)
	app.Get("/orders/callout", orderHandler.OrderDisplay)
	app.Get("/robots.txt", publicHandler.Robots)
	app.Get("/sitemap.xml", publicHandler.Sitemap)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("OK")
	})

	// API Endpoints
	api := app.Group("/api", apiLimiter)
	api.Get("/booking-availability", publicHandler.BookingAvailability)
	api.Post("/book-call", publicHandler.BookCall)
	api.Post("/new-lead", publicHandler.CreateLead)
	api.Post("/leads", publicHandler.CreateLead)
	api.Post("/convert-webp", publicHandler.ConvertWebP)
	api.Get("/orders/:code/status", orderHandler.OrderStatusAPI)
	api.Get("/orders/display-data", orderHandler.OrderDisplayDataAPI)

	// Progressive API Endpoints (<2ms Envelope & Deferred Hydration)
	api.Get("/v1/page/home", progressiveHandler.HomeProgressive)
	api.Get("/v1/page/home/projects", progressiveHandler.HomeProjectsSubResource)
	api.Get("/v1/page/home/blogs", progressiveHandler.HomeBlogsSubResource)
	api.Get("/v1/dashboard/overview", progressiveHandler.DashboardOverviewSkeleton)

	// Auth Routes
	app.Get("/login", authHandler.LoginPage)
	app.Post("/login", loginLimiter, authHandler.LoginSubmit)
	app.Get("/logout", authHandler.Logout)

	// Admin Routes (Protected)
	admin := app.Group("/admin", middleware.RequireAuth(cfg.SecretKey))
	admin.Get("/", adminHandler.Dashboard)

	// Orders & Call-Out Management
	admin.Get("/orders", orderHandler.AdminOrdersPage)
	admin.Post("/orders/new", orderHandler.AdminCreateOrder)
	admin.Post("/orders/:id/status", orderHandler.AdminUpdateOrderStatus)
	admin.Post("/orders/:id/notify", orderHandler.AdminNotifyOrder)
	admin.Post("/orders/:id/delete", orderHandler.AdminDeleteOrder)

	// Analytics
	admin.Get("/analytics", adminHandler.AnalyticsPage)
	admin.Get("/analytics_data", adminHandler.AnalyticsData)

	// Leads
	admin.Get("/leads", adminHandler.LeadsPage)
	admin.Get("/leads/export", adminHandler.ExportLeads)
	admin.Post("/leads/new", adminHandler.AddLead)
	admin.Post("/add_lead", adminHandler.AddLead)
	admin.Get("/get_lead/:id", adminHandler.GetLead)
	admin.Post("/leads/:id/edit", adminHandler.UpdateLead)
	admin.Post("/update_lead/:id", adminHandler.UpdateLead)
	admin.Patch("/update_lead/:id", adminHandler.UpdateLead)
	admin.Post("/leads/:id/delete", adminHandler.DeleteLead)
	admin.Post("/delete_lead/:id", adminHandler.DeleteLead)
	admin.Delete("/delete_lead/:id", adminHandler.DeleteLead)

	// Projects
	admin.Get("/projects", adminHandler.ProjectsPage)
	admin.Post("/projects/new", adminHandler.SaveProject)
	admin.Post("/add_project", adminHandler.SaveProject)
	admin.Post("/update_project/:id", adminHandler.UpdateProject)
	admin.Post("/projects/:id/delete", adminHandler.DeleteProject)
	admin.Post("/delete_project/:id", adminHandler.DeleteProject)
	admin.Delete("/delete_project/:id", adminHandler.DeleteProject)

	// Blogs
	admin.Get("/blogs", adminHandler.BlogsPage)
	admin.Post("/blogs", adminHandler.SaveBlogPost)
	admin.Post("/blogs/new", adminHandler.SaveBlogPost)
	admin.Post("/add_blog", adminHandler.SaveBlogPost)
	admin.Post("/update_blog/:id", adminHandler.UpdateBlog)
	admin.Post("/blogs/:id/delete", adminHandler.DeleteBlogPost)
	admin.Post("/delete_blog/:id", adminHandler.DeleteBlogPost)
	admin.Delete("/delete_blog/:id", adminHandler.DeleteBlogPost)

	// FAQs
	admin.Get("/faqs", adminHandler.FAQsPage)
	admin.Post("/faqs/new", adminHandler.SaveFAQ)
	admin.Post("/add_faq", adminHandler.SaveFAQ)
	admin.Post("/update_faq/:id", adminHandler.UpdateFAQ)
	admin.Post("/faqs/:id/delete", adminHandler.DeleteFAQ)
	admin.Post("/delete_faq/:id", adminHandler.DeleteFAQ)
	admin.Delete("/delete_faq/:id", adminHandler.DeleteFAQ)
	admin.Post("/delete_faq_submission/:id", adminHandler.DeleteFAQSubmission)
	admin.Delete("/delete_faq_submission/:id", adminHandler.DeleteFAQSubmission)

	// Services
	admin.Get("/services", adminHandler.ServicesPage)
	admin.Post("/add_service", adminHandler.AddService)
	admin.Post("/update_service/:id", adminHandler.UpdateService)
	admin.Post("/delete_service/:id", adminHandler.DeleteService)
	admin.Delete("/delete_service/:id", adminHandler.DeleteService)

	// Booking Settings
	admin.Get("/booking", adminHandler.BookingSettingsPage)
	admin.Post("/booking", adminHandler.UpdateBookingSettings)

	// Invoices & Quotes
	admin.Get("/invoices", adminHandler.InvoicesPage)
	admin.Post("/save_invoice_settings", adminHandler.SaveInvoiceSettings)
	admin.Post("/invoices/generate", adminHandler.GenerateInvoicePDF)

	// Funds & Financials
	admin.Get("/funds", adminHandler.FundsPage)
	admin.Post("/add_transaction", adminHandler.AddTransaction)
	admin.Post("/delete_transaction/:id", adminHandler.DeleteTransaction)
	admin.Delete("/delete_transaction/:id", adminHandler.DeleteTransaction)
	admin.Get("/export_transactions", adminHandler.ExportTransactions)

	// Automation Ops
	admin.Get("/automation", adminHandler.AutomationPage)
	admin.Post("/run_script", adminHandler.RunScript)

	// Logout inside admin
	admin.Get("/logout", authHandler.Logout)

	// Graceful Shutdown
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("🚀 JO4 Dev Go Fiber Server listening on http://localhost%s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("Server listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	if err := workerPool.Shutdown(5 * time.Second); err != nil {
		log.Printf("Worker pool shutdown error: %v", err)
	}
	log.Println("Server and background workers exited cleanly.")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/database"
	"github.com/JRoetscyber/my_website/go_app/internal/handlers"
	"github.com/JRoetscyber/my_website/go_app/internal/middleware"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
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

var routeMap = map[string]string{
	"home":                         "/",
	"services.services":            "/services",
	"services.web_design":          "/web-design",
	"services.seo_services":        "/seo",
	"services.automation":          "/automation",
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

func initPongo2Filters() {
	// Register markdown filter
	pongo2.RegisterFilter("markdown", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(in.String()), &buf); err != nil {
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

	// Register split filter
	pongo2.RegisterFilter("split", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		sep := param.String()
		return pongo2.AsValue(strings.Split(in.String(), sep)), nil
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
	pongo2.Globals["static_url"] = func(filename string) string {
		return "/static/" + strings.TrimPrefix(filename, "/")
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
	templatesDir := "./views"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		templatesDir = "../views"
	}
	engine := django.New(templatesDir, ".html")
	engine.Reload(cfg.Environment == "development")

	// Register functions on the engine instance
	engine.AddFunc("static_url", func(filename string) string {
		return "/static/" + strings.TrimPrefix(filename, "/")
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
		Views:       engine,
		AppName:     "JO4 Dev High-Performance Go Fiber Server",
		BodyLimit:   32 * 1024 * 1024, // 32MB max upload
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

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
	app.Use(middleware.TrackAnalytics(db))

	// Rate limiter for API routes
	apiLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
	})

	// Static Assets
	staticDir := "../static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "./static"
	}
	app.Static("/static", staticDir, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
	})

	// Favicon and root assets
	if faviconPath := filepath.Join(filepath.Dir(staticDir), "favicon.ico"); fileExists(faviconPath) {
		app.Static("/favicon.ico", faviconPath)
	}
	if appleIcon := filepath.Join(filepath.Dir(staticDir), "apple-touch-icon.png"); fileExists(appleIcon) {
		app.Static("/apple-touch-icon.png", appleIcon)
	}

	// Handlers
	publicHandler := handlers.NewPublicHandler(db, cfg)
	authHandler := handlers.NewAuthHandler(db, cfg)
	adminHandler := handlers.NewAdminHandler(db, cfg)

	// Public Routes
	app.Get("/", publicHandler.Home)
	app.Get("/services", publicHandler.Services)
	app.Get("/web-design", publicHandler.WebDesign)
	app.Get("/seo", publicHandler.SEO)
	app.Get("/appsec", publicHandler.AppSec)
	app.Get("/automation", publicHandler.Automation)
	app.Get("/projects", publicHandler.Projects)
	app.Get("/blog", publicHandler.BlogList)
	app.Get("/blog/:slug", publicHandler.BlogDetail)
	app.Get("/faq", publicHandler.FAQList)
	app.Get("/faq/:slug", publicHandler.FAQDetail)
	app.Post("/faq/submit", publicHandler.SubmitFAQ)
	app.Get("/book", publicHandler.BookPage)
	app.Get("/robots.txt", publicHandler.Robots)
	app.Get("/sitemap.xml", publicHandler.Sitemap)

	// API Endpoints
	api := app.Group("/api", apiLimiter)
	api.Get("/booking-availability", publicHandler.BookingAvailability)
	api.Post("/book-call", publicHandler.BookCall)
	api.Post("/new-lead", publicHandler.CreateLead)
	api.Post("/leads", publicHandler.CreateLead)

	// Auth Routes
	app.Get("/login", authHandler.LoginPage)
	app.Post("/login", authHandler.LoginSubmit)
	app.Get("/logout", authHandler.Logout)

	// Admin Routes (Protected)
	admin := app.Group("/admin", middleware.RequireAuth(cfg.SecretKey))
	admin.Get("/", adminHandler.Dashboard)
	admin.Get("/leads", adminHandler.LeadsPage)
	admin.Get("/leads/export", adminHandler.ExportLeads)
	admin.Post("/leads/new", adminHandler.CreateLead)
	admin.Post("/leads/:id/edit", adminHandler.UpdateLead)
	admin.Post("/leads/:id/delete", adminHandler.DeleteLead)
	admin.Get("/analytics", adminHandler.AnalyticsPage)
	admin.Get("/projects", adminHandler.ProjectsPage)
	admin.Post("/projects/new", adminHandler.SaveProject)
	admin.Post("/projects/:id/delete", adminHandler.DeleteProject)
	admin.Get("/blogs", adminHandler.BlogsPage)
	admin.Post("/blogs/new", adminHandler.SaveBlogPost)
	admin.Post("/blogs/:id/delete", adminHandler.DeleteBlogPost)
	admin.Get("/faqs", adminHandler.FAQsPage)
	admin.Post("/faqs/new", adminHandler.SaveFAQ)
	admin.Post("/faqs/:id/delete", adminHandler.DeleteFAQ)
	admin.Get("/services", adminHandler.ServicesPage)
	admin.Get("/booking", adminHandler.BookingSettingsPage)
	admin.Post("/booking", adminHandler.UpdateBookingSettings)
	admin.Get("/invoices", adminHandler.InvoicesPage)
	admin.Post("/invoices/generate", adminHandler.GenerateInvoicePDF)
	admin.Get("/funds", adminHandler.FundsPage)
	admin.Get("/automation", adminHandler.AutomationPage)

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
	_ = app.Shutdown()
	log.Println("Server exited cleanly.")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

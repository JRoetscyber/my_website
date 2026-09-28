package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/cache"
	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/database"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/services"
	"github.com/JRoetscyber/my_website/go_app/internal/worker"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PublicHandler struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Cache  *cache.Manager
	Worker *worker.Pool
}

func NewPublicHandler(db *gorm.DB, cfg *config.Config, cm *cache.Manager, pool *worker.Pool) *PublicHandler {
	return &PublicHandler{
		DB:     db,
		Cfg:    cfg,
		Cache:  cm,
		Worker: pool,
	}
}

// Home page with Redis / Memory Cache-Aside & Thundering Herd Singleflight Protection
func (h *PublicHandler) Home(c *fiber.Ctx) error {
	ctx := context.Background()

	projects, _ := cache.GetOrSet(h.Cache, ctx, cache.ProjectListKey(), 10*time.Minute, func(ctx context.Context) ([]models.Project, error) {
		var list []models.Project
		err := h.DB.WithContext(ctx).Order("id desc").Limit(6).Find(&list).Error
		return list, err
	})

	blogPosts, _ := cache.GetOrSet(h.Cache, ctx, cache.BlogListKey(), 10*time.Minute, func(ctx context.Context) ([]models.BlogPost, error) {
		var list []models.BlogPost
		err := h.DB.WithContext(ctx).Order("created_at desc").Limit(3).Find(&list).Error
		return list, err
	})

	faqs, _ := cache.GetOrSet(h.Cache, ctx, cache.FAQsKey(), 30*time.Minute, func(ctx context.Context) ([]models.FAQ, error) {
		var list []models.FAQ
		err := h.DB.WithContext(ctx).Where("is_published = ?", true).Order("display_order asc, id asc").Limit(6).Find(&list).Error
		return list, err
	})

	servicesList, _ := cache.GetOrSet(h.Cache, ctx, cache.ServicesKey(), 30*time.Minute, func(ctx context.Context) ([]models.Service, error) {
		var list []models.Service
		err := h.DB.WithContext(ctx).Where("is_published = ?", true).Order("display_order asc, id asc").Find(&list).Error
		return list, err
	})

	return c.Render("index", fiber.Map{
		"projects": projects,
		"blogs":    blogPosts,
		"faqs":     faqs,
		"services": servicesList,
		"request":  c,
	})
}

// Services overview
func (h *PublicHandler) Services(c *fiber.Ctx) error {
	servicesList, _ := cache.GetOrSet(h.Cache, context.Background(), cache.ServicesKey(), 30*time.Minute, func(ctx context.Context) ([]models.Service, error) {
		var list []models.Service
		err := h.DB.WithContext(ctx).Where("is_published = ?", true).Order("display_order asc, id asc").Find(&list).Error
		return list, err
	})

	return c.Render("services", fiber.Map{
		"services": servicesList,
		"request":  c,
	})
}

// Dedicated service pages
func (h *PublicHandler) WebDesign(c *fiber.Ctx) error {
	return c.Render("web_design", fiber.Map{"request": c})
}

func (h *PublicHandler) SEO(c *fiber.Ctx) error {
	return c.Render("seo", fiber.Map{"request": c})
}

func (h *PublicHandler) AppSec(c *fiber.Ctx) error {
	return c.Render("appsec", fiber.Map{"request": c})
}

func (h *PublicHandler) AppDev(c *fiber.Ctx) error {
	return c.Render("app_development", fiber.Map{"request": c})
}

func (h *PublicHandler) Automation(c *fiber.Ctx) error {
	return c.Render("automation", fiber.Map{"request": c})
}

// Projects / Portfolio
func (h *PublicHandler) Projects(c *fiber.Ctx) error {
	projects, _ := cache.GetOrSet(h.Cache, context.Background(), cache.ProjectListKey(), 15*time.Minute, func(ctx context.Context) ([]models.Project, error) {
		var list []models.Project
		err := h.DB.WithContext(ctx).Order("id desc").Find(&list).Error
		return list, err
	})

	return c.Render("projects", fiber.Map{
		"projects": projects,
		"request":  c,
	})
}

// Project detail
func (h *PublicHandler) ProjectDetail(c *fiber.Ctx) error {
	slug := strings.Clone(c.Params("slug"))
	project, err := cache.GetOrSet(h.Cache, context.Background(), cache.ProjectKey(slug), 15*time.Minute, func(ctx context.Context) (models.Project, error) {
		var p models.Project
		err := h.DB.WithContext(ctx).Where("slug = ?", slug).First(&p).Error
		return p, err
	})
	if err != nil || project.ID == 0 {
		return c.Status(http.StatusNotFound).SendString("Project not found")
	}

	// Increment view count asynchronously via bounded worker pool (fasthttp safe)
	if h.Worker != nil {
		projID := project.ID
		h.Worker.Enqueue(func(ctx context.Context) {
			h.DB.WithContext(ctx).Model(&models.Project{}).Where("id = ?", projID).UpdateColumn("views", gorm.Expr("views + ?", 1))
		})
	}

	return c.Render("projects", fiber.Map{
		"project":         project,
		"seo_title":       project.Title + " — JO4 Dev Case Study",
		"seo_description": project.Description,
		"request":         c,
	})
}

// Blog list
func (h *PublicHandler) BlogList(c *fiber.Ctx) error {
	posts, _ := cache.GetOrSet(h.Cache, context.Background(), cache.BlogListKey(), 15*time.Minute, func(ctx context.Context) ([]models.BlogPost, error) {
		var list []models.BlogPost
		err := h.DB.WithContext(ctx).Order("created_at desc").Find(&list).Error
		return list, err
	})

	return c.Render("blog", fiber.Map{
		"posts":   posts,
		"request": c,
	})
}

// Blog detail
func (h *PublicHandler) BlogDetail(c *fiber.Ctx) error {
	slug := strings.Clone(c.Params("slug"))
	post, err := cache.GetOrSet(h.Cache, context.Background(), cache.BlogKey(slug), 15*time.Minute, func(ctx context.Context) (models.BlogPost, error) {
		var p models.BlogPost
		err := h.DB.WithContext(ctx).Where("slug = ?", slug).First(&p).Error
		return p, err
	})
	if err != nil || post.ID == 0 {
		return c.Status(http.StatusNotFound).SendString("Post not found")
	}

	// Increment view count asynchronously via bounded worker pool (fasthttp safe)
	if h.Worker != nil {
		postID := post.ID
		h.Worker.Enqueue(func(ctx context.Context) {
			h.DB.WithContext(ctx).Model(&models.BlogPost{}).Where("id = ?", postID).UpdateColumn("views", gorm.Expr("views + 1"))
		})
	}

	var recentPosts []models.BlogPost
	h.DB.Where("id != ?", post.ID).Order("created_at desc").Limit(3).Find(&recentPosts)

	return c.Render("blog_detail", fiber.Map{
		"post":         post,
		"recent_posts": recentPosts,
		"request":      c,
	})
}

// FAQ list
func (h *PublicHandler) FAQList(c *fiber.Ctx) error {
	faqs, _ := cache.GetOrSet(h.Cache, context.Background(), cache.FAQsKey(), 30*time.Minute, func(ctx context.Context) ([]models.FAQ, error) {
		var list []models.FAQ
		err := h.DB.WithContext(ctx).Where("is_published = ?", true).Order("display_order asc, id asc").Find(&list).Error
		return list, err
	})

	return c.Render("faq", fiber.Map{
		"faqs":    faqs,
		"request": c,
	})
}

// FAQ detail
func (h *PublicHandler) FAQDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var faq models.FAQ
	if err := h.DB.Where("slug = ?", slug).First(&faq).Error; err != nil {
		return c.Status(http.StatusNotFound).SendString("FAQ not found")
	}

	return c.Render("faq_detail", fiber.Map{
		"faq":     faq,
		"request": c,
	})
}

// Submit FAQ question
func (h *PublicHandler) SubmitFAQ(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	phone := strings.TrimSpace(c.FormValue("phone"))
	question := strings.TrimSpace(c.FormValue("question"))

	if name == "" || email == "" || question == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name, email, and question are required."})
	}

	sub := models.FAQSubmission{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Question: question,
	}
	h.DB.Create(&sub)

	return c.JSON(fiber.Map{"status": "success", "message": "Thank you! Your question has been submitted."})
}

// Book Call Page
func (h *PublicHandler) BookPage(c *fiber.Ctx) error {
	return c.Render("book", fiber.Map{"request": c})
}

// Booking Availability API
func (h *PublicHandler) BookingAvailability(c *fiber.Ctx) error {
	dateVal := c.Query("date")
	if dateVal == "" {
		dateVal = time.Now().In(services.SAST).Format("2006-01-02")
	}

	settings := database.GetBookingSettings(h.DB)
	result := services.BuildAvailableSlots(settings, dateVal)

	return c.JSON(result)
}

// Book Call API
type BookCallRequest struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Company     string  `json:"company"`
	Phone       string  `json:"phone"`
	ProjectType string  `json:"project_type"`
	Budget      float64 `json:"budget"`
	ContactRole string  `json:"contact_role"`
	Notes       string  `json:"notes"`
	ScheduledAt string  `json:"scheduled_at"`
}

func (h *PublicHandler) BookCall(c *fiber.Ctx) error {
	var req BookCallRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid request body."})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.ScheduledAt = strings.TrimSpace(req.ScheduledAt)

	if req.Name == "" || req.Email == "" || req.ScheduledAt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Name, email, and meeting time are required."})
	}

	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		scheduledTime, err = time.ParseInLocation("2006-01-02T15:04:05", req.ScheduledAt, services.SAST)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid meeting time format."})
		}
	}

	if scheduledTime.Before(time.Now().In(services.SAST)) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Please choose a future meeting time."})
	}

	contactRole := req.ContactRole
	if contactRole == "" {
		contactRole = "Owner"
	}
	projectType := req.ProjectType
	if projectType == "" {
		projectType = "static"
	}

	scoreRes := services.CalculateLeadScore(services.LeadScoringInput{
		ClientCompany:      req.Company,
		ContactRole:        contactRole,
		Budget:             req.Budget,
		ProjectType:        projectType,
		VisitedPages:       []string{"/book"},
		WhatsappEngagement: func() string { if req.Phone != "" { return "replied" }; return "read" }(),
		PhoneNumber:        req.Phone,
		LastActivityDate:   time.Now(),
	})

	lead := models.Lead{
		ClientName:         req.Name,
		ClientCompany:      req.Company,
		ProjectType:        projectType,
		Budget:             req.Budget,
		ContactRole:        contactRole,
		PhoneNumber:        req.Phone,
		WhatsappEngagement: func() string { if req.Phone != "" { return "replied" }; return "read" }(),
		TargetProject:      projectType,
		Score:              int(math.Round(scoreRes.Score)),
		ExplicitScore:      scoreRes.Breakdown.Explicit,
		ImplicitScore:      scoreRes.Breakdown.Implicit,
		UrgencyScore:       scoreRes.Breakdown.Urgency,
		Status:             "New",
		LastActivityDate:   time.Now(),
	}

	if err := h.DB.Create(&lead).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Failed to save booking: %v", err)})
	}

	// Send email confirmation asynchronously with .ics calendar attachment in protected goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Worker] Panic recovered in SendBookingInvite: %v", r)
			}
		}()
		_ = services.SendBookingInvite(h.Cfg, req.Email, req.Name, scheduledTime)
	}()

	calURL := services.GoogleCalendarURL(scheduledTime, req.Name, req.Company, req.Email, req.Phone, projectType, req.Notes)

	return c.JSON(fiber.Map{
		"status":                 "success",
		"message":                "Your discovery call is booked!",
		"calendar_url":           calURL,
		"calendar_event_created": true,
		"lead_id":                lead.ID,
		"score":                  scoreRes.Score,
		"classification":         scoreRes.Classification,
		"admin_url":              "/admin/leads",
	})
}

// Generic lead creation API endpoint
func (h *PublicHandler) CreateLead(c *fiber.Ctx) error {
	var body map[string]interface{}
	_ = json.Unmarshal(c.Body(), &body)

	name := fmt.Sprintf("%v", body["client_name"])
	if name == "" || name == "<nil>" {
		name = fmt.Sprintf("%v", body["name"])
	}
	company := fmt.Sprintf("%v", body["client_company"])
	if company == "<nil>" {
		company = ""
	}
	phone := fmt.Sprintf("%v", body["phone_number"])
	if phone == "<nil>" {
		phone = fmt.Sprintf("%v", body["phone"])
	}
	if phone == "<nil>" {
		phone = ""
	}
	projectType := fmt.Sprintf("%v", body["project_type"])
	if projectType == "<nil>" {
		projectType = "Static"
	}

	var budget float64
	if b, ok := body["budget"].(float64); ok {
		budget = b
	} else if bStr, ok := body["budget"].(string); ok {
		budget, _ = strconv.ParseFloat(bStr, 64)
	}

	scoreRes := services.CalculateLeadScore(services.LeadScoringInput{
		ClientCompany:      company,
		ContactRole:        "Owner",
		Budget:             budget,
		ProjectType:        projectType,
		VisitedPages:       []string{"/"},
		WhatsappEngagement: "read",
		PhoneNumber:        phone,
		LastActivityDate:   time.Now(),
	})

	lead := models.Lead{
		ClientName:         name,
		ClientCompany:      company,
		ProjectType:        projectType,
		Budget:             budget,
		ContactRole:        "Owner",
		PhoneNumber:        phone,
		WhatsappEngagement: "read",
		TargetProject:      projectType,
		Score:              int(math.Round(scoreRes.Score)),
		ExplicitScore:      scoreRes.Breakdown.Explicit,
		ImplicitScore:      scoreRes.Breakdown.Implicit,
		UrgencyScore:       scoreRes.Breakdown.Urgency,
		Status:             "New",
		LastActivityDate:   time.Now(),
	}

	if err := h.DB.Create(&lead).Error; err != nil {
		log.Printf("[DB] Error creating lead: %v", err)
	}

	// Asynchronously handle lead notification in a worker goroutine
	go func(leadData models.Lead) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Worker] Panic recovered in lead notification: %v", r)
			}
		}()
		// If mail notifications are enabled, notify admin
		if h.Cfg.MailUsername != "" && h.Cfg.MailPassword != "" {
			_ = services.SendLeadNotification(h.Cfg, leadData)
		}
	}(lead)

	return c.JSON(fiber.Map{"status": "success", "lead_id": lead.ID, "score": scoreRes.Score})
}

// Robots.txt
func (h *PublicHandler) Robots(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain")
	return c.SendString("User-agent: *\nDisallow: /admin\nDisallow: /login\nSitemap: https://jo4.co.za/sitemap.xml\n")
}

// Sitemap.xml
func (h *PublicHandler) Sitemap(c *fiber.Ctx) error {
	var (
		projects []models.Project
		posts    []models.BlogPost
		faqs     []models.FAQ
		wg       sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		h.DB.Find(&projects)
	}()
	go func() {
		defer wg.Done()
		h.DB.Find(&posts)
	}()
	go func() {
		defer wg.Done()
		h.DB.Where("is_published = ?", true).Find(&faqs)
	}()
	wg.Wait()

	nowStr := time.Now().Format("2006-01-02")

	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")

	staticRoutes := []struct {
		Path     string
		Priority string
		Freq     string
	}{
		{"/", "1.0", "weekly"},
		{"/services", "0.8", "monthly"},
		{"/web-design", "0.8", "monthly"},
		{"/seo", "0.8", "monthly"},
		{"/appsec", "0.8", "monthly"},
		{"/automation", "0.8", "monthly"},
		{"/projects", "0.9", "weekly"},
		{"/blog", "0.8", "weekly"},
		{"/faq", "0.7", "weekly"},
		{"/book", "0.9", "weekly"},
		{"/webp-converter", "0.8", "monthly"},
	}

	for _, sr := range staticRoutes {
		sb.WriteString(fmt.Sprintf("  <url><loc>https://jo4.co.za%s</loc><lastmod>%s</lastmod><changefreq>%s</changefreq><priority>%s</priority></url>\n", sr.Path, nowStr, sr.Freq, sr.Priority))
	}

	for _, p := range projects {
		if p.Slug != "" {
			dStr := nowStr
			if !p.DeployedAt.IsZero() {
				dStr = p.DeployedAt.Format("2006-01-02")
			}
			sb.WriteString(fmt.Sprintf("  <url><loc>https://jo4.co.za/projects/%s</loc><lastmod>%s</lastmod><changefreq>monthly</changefreq><priority>0.9</priority></url>\n", p.Slug, dStr))
		}
	}

	for _, b := range posts {
		if b.Slug != "" {
			uStr := nowStr
			if !b.UpdatedAt.IsZero() {
				uStr = b.UpdatedAt.Format("2006-01-02")
			}
			sb.WriteString(fmt.Sprintf("  <url><loc>https://jo4.co.za/blog/%s</loc><lastmod>%s</lastmod><changefreq>weekly</changefreq><priority>0.7</priority></url>\n", b.Slug, uStr))
		}
	}

	for _, f := range faqs {
		if f.Slug != "" {
			sb.WriteString(fmt.Sprintf("  <url><loc>https://jo4.co.za/faq/%s</loc><lastmod>%s</lastmod><changefreq>monthly</changefreq><priority>0.6</priority></url>\n", f.Slug, nowStr))
		}
	}

	sb.WriteString("</urlset>\n")

	c.Set("Content-Type", "application/xml")
	return c.SendString(sb.String())
}

// WebP Converter Page
func (h *PublicHandler) WebPConverterPage(c *fiber.Ctx) error {
	return c.Render("webp_converter", fiber.Map{
		"current_path": "/webp-converter",
		"title": "Free High-Speed WebP Converter — JO4 Dev",
		"meta_description": "Convert JPG, PNG, and BMP images into high-compression, next-gen WebP format with sub-second execution. Engineered by JO4 Dev.",
	})
}

// ConvertWebP API
func (h *PublicHandler) ConvertWebP(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No image file provided"})
	}

	quality := strings.TrimSpace(c.FormValue("quality"))
	if quality == "" {
		quality = "80"
	}
	method := strings.TrimSpace(c.FormValue("method"))
	if method == "" {
		method = "6"
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".bmp" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only JPG, PNG, and BMP formats are supported"})
	}

	// Create upload folder
	uploadDir := filepath.Join("static", "uploads", "webp")
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		uploadDir = filepath.Join("../static", "uploads", "webp")
	}
	_ = os.MkdirAll(uploadDir, 0755)

	randSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%100000)
	baseName := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	inputFilename := fmt.Sprintf("%s_%s%s", baseName, randSuffix, ext)
	outputFilename := fmt.Sprintf("%s_%s.webp", baseName, randSuffix)

	inputPath := filepath.Join(uploadDir, inputFilename)
	outputPath := filepath.Join(uploadDir, outputFilename)

	if err := c.SaveFile(file, inputPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save uploaded image"})
	}
	defer os.Remove(inputPath)

	res, err := services.ConvertToWebP(inputPath, outputPath, quality, method)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}


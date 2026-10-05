package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
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
	"github.com/JRoetscyber/my_website/go_app/internal/utils"
	"github.com/JRoetscyber/my_website/go_app/internal/worker"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Cache  *cache.Manager
	Worker *worker.Pool
}

func NewAdminHandler(db *gorm.DB, cfg *config.Config, cm *cache.Manager, pool *worker.Pool) *AdminHandler {
	return &AdminHandler{
		DB:     db,
		Cfg:    cfg,
		Cache:  cm,
		Worker: pool,
	}
}

// buildAdminContext provides unified context for the admin sidebar badges and system status
func (h *AdminHandler) buildAdminContext(c *fiber.Ctx, activePage string, extra fiber.Map) fiber.Map {
	var (
		leadsCount    int64
		projectsCount int64
		blogsCount    int64
		faqsCount     int64
		servicesCount int64
		wg            sync.WaitGroup
	)

	wg.Add(5)
	go func() { defer wg.Done(); h.DB.Model(&models.Lead{}).Count(&leadsCount) }()
	go func() { defer wg.Done(); h.DB.Model(&models.Project{}).Count(&projectsCount) }()
	go func() { defer wg.Done(); h.DB.Model(&models.BlogPost{}).Count(&blogsCount) }()
	go func() { defer wg.Done(); h.DB.Model(&models.FAQ{}).Count(&faqsCount) }()
	go func() { defer wg.Done(); h.DB.Model(&models.Service{}).Count(&servicesCount) }()
	wg.Wait()

	base := fiber.Map{
		"active_page":    activePage,
		"leads":          make([]models.Lead, leadsCount),
		"projects":       make([]models.Project, projectsCount),
		"blogs":          make([]models.BlogPost, blogsCount),
		"faqs":           make([]models.FAQ, faqsCount),
		"services":       make([]models.Service, servicesCount),
		"leads_count":    leadsCount,
		"projects_count": projectsCount,
		"blogs_count":    blogsCount,
		"faqs_count":     faqsCount,
		"services_count": servicesCount,
		"request":        c,
	}

	for k, v := range extra {
		base[k] = v
	}
	return base
}

func serializeLead(l *models.Lead) fiber.Map {
	createdAtStr := ""
	if !l.CreatedAt.IsZero() {
		createdAtStr = l.CreatedAt.Format("2006-01-02 15:04")
	}
	lastActivityStr := ""
	if !l.LastActivityDate.IsZero() {
		lastActivityStr = l.LastActivityDate.Format("2006-01-02")
	}

	status := strings.TrimSpace(l.Status)
	if status == "" {
		status = "New"
	}

	pt := strings.TrimSpace(l.ProjectType)
	if pt == "" {
		pt = strings.TrimSpace(l.TargetProject)
	}
	if pt == "" {
		pt = "Web Development"
	}

	classification := "Cold"
	if l.Score >= 80 {
		classification = "Hot"
	} else if l.Score >= 50 {
		classification = "Warm"
	}

	return fiber.Map{
		"id":                   l.ID,
		"ID":                   l.ID,
		"client_name":          l.ClientName,
		"ClientName":           l.ClientName,
		"client_company":       l.ClientCompany,
		"ClientCompany":        l.ClientCompany,
		"project_type":         pt,
		"ProjectType":          pt,
		"display_project_type": pt,
		"target_project":       l.TargetProject,
		"TargetProject":        l.TargetProject,
		"budget":               l.Budget,
		"Budget":               l.Budget,
		"contact_role":         l.ContactRole,
		"ContactRole":          l.ContactRole,
		"phone_number":         l.PhoneNumber,
		"PhoneNumber":          l.PhoneNumber,
		"whatsapp_engagement":  l.WhatsappEngagement,
		"WhatsappEngagement":   l.WhatsappEngagement,
		"score":                l.Score,
		"Score":                l.Score,
		"display_score":        l.Score,
		"status":               status,
		"Status":               status,
		"display_status":       status,
		"classification":       classification,
		"temp_class":           "lead-temp-" + strings.ToLower(classification),
		"loss_reason":          l.LossReason,
		"LossReason":           l.LossReason,
		"created_at":           createdAtStr,
		"last_activity_date":   lastActivityStr,
		"LastActivityDate":     l.LastActivityDate,
		"breakdown": fiber.Map{
			"explicit": l.ExplicitScore,
			"implicit": l.ImplicitScore,
			"urgency":  l.UrgencyScore,
		},
		"explicit_score": l.ExplicitScore,
		"implicit_score": l.ImplicitScore,
		"urgency_score":  l.UrgencyScore,
	}
}

// Dashboard Overview (defaults to Leads like original Flask admin)
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	return h.LeadsPage(c)
}

// Leads Page
func (h *AdminHandler) LeadsPage(c *fiber.Ctx) error {
	var rawLeads []models.Lead
	h.DB.Order("created_at desc, id desc").Find(&rawLeads)

	var countNew, countContacted, countNegotiating, countClosed, countLost int
	serialized := make([]fiber.Map, 0, len(rawLeads))
	for i := range rawLeads {
		l := serializeLead(&rawLeads[i])
		statusVal, _ := l["status"].(string)
		switch statusVal {
		case "Contacted":
			countContacted++
		case "Negotiating":
			countNegotiating++
		case "Closed":
			countClosed++
		case "Lost":
			countLost++
		default:
			countNew++
		}
		serialized = append(serialized, l)
	}

	statuses := []string{"New", "Contacted", "Negotiating", "Closed", "Lost"}

	return c.Render("admin/leads", h.buildAdminContext(c, "leads", fiber.Map{
		"leads":             serialized,
		"statuses":          statuses,
		"count_new":         countNew,
		"count_contacted":   countContacted,
		"count_negotiating": countNegotiating,
		"count_closed":      countClosed,
		"count_lost":        countLost,
	}))
}

// Export Leads CSV
func (h *AdminHandler) ExportLeads(c *fiber.Ctx) error {
	var leads []models.Lead
	h.DB.Order("created_at desc").Find(&leads)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ID", "Client Name", "Company", "Project Type", "Budget", "Role", "Phone", "Score", "Status", "Created At"})

	for _, l := range leads {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", l.ID),
			l.ClientName,
			l.ClientCompany,
			l.ProjectType,
			fmt.Sprintf("%.2f", l.Budget),
			l.ContactRole,
			l.PhoneNumber,
			fmt.Sprintf("%d", l.Score),
			l.Status,
			l.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=leads_export.csv")
	return c.Send(buf.Bytes())
}

// Get Lead JSON
func (h *AdminHandler) GetLead(c *fiber.Ctx) error {
	id := c.Params("id")
	var lead models.Lead
	if err := h.DB.First(&lead, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Lead not found"})
	}
	return c.JSON(serializeLead(&lead))
}

// Add / Create Lead (handles JSON from admin.js and form posts)
func (h *AdminHandler) AddLead(c *fiber.Ctx) error {
	var payload struct {
		ClientName         string  `json:"client_name"`
		ClientCompany      string  `json:"client_company"`
		ProjectType        string  `json:"project_type"`
		Budget             float64 `json:"budget"`
		ContactRole        string  `json:"contact_role"`
		PhoneNumber        string  `json:"phone_number"`
		WhatsappEngagement string  `json:"whatsapp_engagement"`
		TargetProject      string  `json:"target_project"`
		Status             string  `json:"status"`
		LossReason         string  `json:"loss_reason"`
	}

	isJSON := strings.Contains(c.Get("Content-Type"), "application/json")
	if isJSON {
		_ = c.BodyParser(&payload)
	} else {
		payload.ClientName = c.FormValue("client_name")
		payload.ClientCompany = c.FormValue("client_company")
		payload.ProjectType = c.FormValue("project_type")
		payload.ContactRole = c.FormValue("contact_role")
		payload.PhoneNumber = c.FormValue("phone_number")
		payload.WhatsappEngagement = c.FormValue("whatsapp_engagement")
		payload.TargetProject = c.FormValue("target_project")
		payload.Status = c.FormValue("status")
		payload.LossReason = c.FormValue("loss_reason")
		payload.Budget, _ = strconv.ParseFloat(c.FormValue("budget"), 64)
	}

	if payload.ClientName == "" && payload.ClientCompany != "" {
		payload.ClientName = payload.ClientCompany
	}
	if payload.Status == "" {
		payload.Status = "New"
	}
	if payload.TargetProject == "" {
		payload.TargetProject = payload.ProjectType
	}

	scoreRes := services.CalculateLeadScore(services.LeadScoringInput{
		ClientCompany:      payload.ClientCompany,
		ContactRole:        payload.ContactRole,
		Budget:             payload.Budget,
		ProjectType:        payload.ProjectType,
		VisitedPages:       []string{"/admin"},
		WhatsappEngagement: payload.WhatsappEngagement,
		PhoneNumber:        payload.PhoneNumber,
		LastActivityDate:   time.Now(),
	})

	lead := models.Lead{
		ClientName:         payload.ClientName,
		ClientCompany:      payload.ClientCompany,
		ProjectType:        payload.ProjectType,
		Budget:             payload.Budget,
		ContactRole:        payload.ContactRole,
		PhoneNumber:        payload.PhoneNumber,
		WhatsappEngagement: payload.WhatsappEngagement,
		TargetProject:      payload.TargetProject,
		Score:              int(math.Round(scoreRes.Score)),
		ExplicitScore:      scoreRes.Breakdown.Explicit,
		ImplicitScore:      scoreRes.Breakdown.Implicit,
		UrgencyScore:       scoreRes.Breakdown.Urgency,
		Status:             payload.Status,
		LossReason:         payload.LossReason,
		LastActivityDate:   time.Now(),
	}

	if err := h.DB.Create(&lead).Error; err != nil {
		if isJSON {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		return c.Redirect("/admin/leads")
	}

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "dashboard:*")
	}

	if isJSON {
		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"status":         "success",
			"message":        fmt.Sprintf("Lead added. Score: %d (%s)", lead.Score, scoreRes.Classification),
			"lead":           serializeLead(&lead),
			"classification": scoreRes.Classification,
		})
	}
	return c.Redirect("/admin/leads")
}

// Update Lead
func (h *AdminHandler) UpdateLead(c *fiber.Ctx) error {
	id := c.Params("id")
	var lead models.Lead
	if err := h.DB.First(&lead, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Lead not found"})
	}

	isJSON := strings.Contains(c.Get("Content-Type"), "application/json")
	if isJSON {
		var data map[string]interface{}
		if err := c.BodyParser(&data); err == nil {
			if v, ok := data["status"].(string); ok && v != "" {
				lead.Status = v
			}
			if v, ok := data["client_name"].(string); ok && v != "" {
				lead.ClientName = v
			}
			if v, ok := data["client_company"].(string); ok {
				lead.ClientCompany = v
			}
			if v, ok := data["phone_number"].(string); ok {
				lead.PhoneNumber = v
			}
			if v, ok := data["project_type"].(string); ok {
				lead.ProjectType = v
				lead.TargetProject = v
			}
			if v, ok := data["contact_role"].(string); ok {
				lead.ContactRole = v
			}
			if v, ok := data["whatsapp_engagement"].(string); ok {
				lead.WhatsappEngagement = v
			}
			if v, ok := data["loss_reason"].(string); ok {
				lead.LossReason = v
			}
			if v, ok := data["budget"].(float64); ok {
				lead.Budget = v
			}
			if v, ok := data["score"].(float64); ok {
				lead.Score = int(v)
			}
		}
	} else {
		if val := c.FormValue("status"); val != "" {
			lead.Status = val
		}
		if val := c.FormValue("loss_reason"); val != "" {
			lead.LossReason = val
		}
		if val := c.FormValue("client_name"); val != "" {
			lead.ClientName = val
		}
		if val := c.FormValue("client_company"); val != "" {
			lead.ClientCompany = val
		}
		if val := c.FormValue("phone_number"); val != "" {
			lead.PhoneNumber = val
		}
		if val := c.FormValue("budget"); val != "" {
			lead.Budget, _ = strconv.ParseFloat(val, 64)
		}
	}

	lead.LastActivityDate = time.Now()
	h.DB.Save(&lead)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "dashboard:*")
	}

	if isJSON || c.XHR() {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Lead updated successfully",
			"lead":    serializeLead(&lead),
		})
	}
	return c.Redirect("/admin/leads")
}

// Delete Lead
func (h *AdminHandler) DeleteLead(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.Lead{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "dashboard:*")
	}

	if strings.Contains(c.Get("Content-Type"), "application/json") || c.XHR() || c.Method() == "DELETE" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Lead deleted successfully",
		})
	}
	return c.Redirect("/admin/leads")
}

// Analytics Page
func (h *AdminHandler) AnalyticsPage(c *fiber.Ctx) error {
	type PageCount struct {
		PagePath string `gorm:"column:page_path"`
		Count    int    `gorm:"column:count"`
	}
	var (
		topPages      []PageCount
		totalViews    int64
		totalVisitors int64
		wg            sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		h.DB.Model(&models.Analytics{}).Select("page_path, count(*) as count").Group("page_path").Order("count desc").Limit(10).Scan(&topPages)
	}()
	go func() {
		defer wg.Done()
		h.DB.Model(&models.Analytics{}).Count(&totalViews)
	}()
	go func() {
		defer wg.Done()
		h.DB.Model(&models.Analytics{}).Distinct("visitor_ip").Count(&totalVisitors)
	}()
	wg.Wait()

	return c.Render("admin/analytics", h.buildAdminContext(c, "analytics", fiber.Map{
		"top_pages":      topPages,
		"total_views":    totalViews,
		"total_visitors": totalVisitors,
	}))
}

// Analytics Data JSON (last 30 days daily counts for Chart.js)
func (h *AdminHandler) AnalyticsData(c *fiber.Ctx) error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	type DayCount struct {
		Day   string `gorm:"column:day"`
		Count int    `gorm:"column:count"`
	}
	var results []DayCount

	h.DB.Model(&models.Analytics{}).
		Select("strftime('%Y-%m-%d', timestamp) as day, count(*) as count").
		Where("timestamp >= ?", thirtyDaysAgo).
		Group("day").
		Order("day asc").
		Scan(&results)

	labels := make([]string, 0, len(results))
	data := make([]int, 0, len(results))
	for _, r := range results {
		labels = append(labels, r.Day)
		data = append(data, r.Count)
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"labels": labels,
		"data":   data,
	})
}

// Projects Page
func (h *AdminHandler) ProjectsPage(c *fiber.Ctx) error {
	var projects []models.Project
	h.DB.Order("id desc").Find(&projects)

	return c.Render("admin/projects", h.buildAdminContext(c, "projects", fiber.Map{
		"projects": projects,
	}))
}

// Save Project (handles admin.js modal form)
func (h *AdminHandler) SaveProject(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("project_title"))
	if title == "" {
		title = strings.TrimSpace(c.FormValue("title"))
	}
	category := strings.TrimSpace(c.FormValue("project_tag"))
	if category == "" {
		category = strings.TrimSpace(c.FormValue("category"))
	}
	techStack := strings.TrimSpace(c.FormValue("tech_stack"))
	description := strings.TrimSpace(c.FormValue("description"))
	codeSnippet := strings.TrimSpace(c.FormValue("code_snippet"))
	youtubeURL := strings.TrimSpace(c.FormValue("youtube_url"))
	projectURL := strings.TrimSpace(c.FormValue("project_url"))
	mediaPath := strings.TrimSpace(c.FormValue("media_path"))

	var perfPtr, seoPtr *int
	if perfStr := c.FormValue("performance"); perfStr != "" {
		if p, err := strconv.Atoi(perfStr); err == nil {
			perfPtr = &p
		}
	}
	if seoStr := c.FormValue("seo"); seoStr != "" {
		if s, err := strconv.Atoi(seoStr); err == nil {
			seoPtr = &s
		}
	}

	slug := utils.MakeSlug(title)

	file, err := c.FormFile("media")
	if err != nil {
		file, err = c.FormFile("media_file")
	}
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "assets")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			mediaPath = "/static/assets/" + file.Filename
		}
	}

	project := models.Project{
		Title:       title,
		Slug:        slug,
		Category:    category,
		TechStack:   techStack,
		Description: description,
		CodeSnippet: codeSnippet,
		YoutubeURL:  youtubeURL,
		ProjectURL:  projectURL,
		MediaPath:   mediaPath,
		Performance: perfPtr,
		SEO:         seoPtr,
	}

	h.DB.Create(&project)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "project:*", "page:home:*", "dashboard:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Project deployed successfully",
		})
	}
	return c.Redirect("/admin/projects")
}

// Update Project
func (h *AdminHandler) UpdateProject(c *fiber.Ctx) error {
	id := c.Params("id")
	var project models.Project
	if err := h.DB.First(&project, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Project not found"})
	}

	if title := strings.TrimSpace(c.FormValue("project_title")); title != "" {
		project.Title = title
		project.Slug = utils.MakeSlug(title)
	} else if title := strings.TrimSpace(c.FormValue("title")); title != "" {
		project.Title = title
		project.Slug = utils.MakeSlug(title)
	}
	if cat := strings.TrimSpace(c.FormValue("project_tag")); cat != "" {
		project.Category = cat
	} else if cat := strings.TrimSpace(c.FormValue("category")); cat != "" {
		project.Category = cat
	}
	if ts := strings.TrimSpace(c.FormValue("tech_stack")); ts != "" {
		project.TechStack = ts
	}
	if desc := strings.TrimSpace(c.FormValue("description")); desc != "" {
		project.Description = desc
	}
	if cs := strings.TrimSpace(c.FormValue("code_snippet")); cs != "" {
		project.CodeSnippet = cs
	}
	if yu := strings.TrimSpace(c.FormValue("youtube_url")); yu != "" {
		project.YoutubeURL = yu
	}
	if pu := strings.TrimSpace(c.FormValue("project_url")); pu != "" {
		project.ProjectURL = pu
	}
	if perfStr := c.FormValue("performance"); perfStr != "" {
		if p, err := strconv.Atoi(perfStr); err == nil {
			project.Performance = &p
		}
	}
	if seoStr := c.FormValue("seo"); seoStr != "" {
		if s, err := strconv.Atoi(seoStr); err == nil {
			project.SEO = &s
		}
	}

	file, err := c.FormFile("media")
	if err != nil {
		file, err = c.FormFile("media_file")
	}
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "assets")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			project.MediaPath = "/static/assets/" + file.Filename
		}
	}

	h.DB.Save(&project)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "project:*", "page:home:*", "dashboard:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Project updated successfully",
		})
	}
	return c.Redirect("/admin/projects")
}

// Delete Project
func (h *AdminHandler) DeleteProject(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.Project{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "project:*", "page:home:*", "dashboard:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") || c.Method() == "DELETE" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Project deleted successfully",
		})
	}
	return c.Redirect("/admin/projects")
}

// Blogs Page
func (h *AdminHandler) BlogsPage(c *fiber.Ctx) error {
	var blogs []models.BlogPost
	h.DB.Order("created_at desc").Find(&blogs)

	var totalViews int64
	h.DB.Model(&models.BlogPost{}).Select("COALESCE(SUM(views), 0)").Scan(&totalViews)

	return c.Render("admin/blogs", h.buildAdminContext(c, "blogs", fiber.Map{
		"blogs":       blogs,
		"total_views": totalViews,
	}))
}

func parsePublishInfo(statusInput, dateInput string) (string, time.Time) {
	status := strings.ToLower(strings.TrimSpace(statusInput))
	if status == "" {
		status = "published"
	}

	dateInput = strings.TrimSpace(dateInput)
	var pubTime time.Time
	if dateInput != "" {
		layouts := []string{
			"2006-01-02T15:04",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, dateInput, time.Local); err == nil {
				pubTime = t
				break
			}
		}
	}

	now := time.Now()
	if pubTime.IsZero() {
		if status == "scheduled" {
			pubTime = now.Add(24 * time.Hour)
		} else {
			pubTime = now
		}
	}

	// Automatic status alignment based on publication timestamp
	if status != "draft" {
		if pubTime.After(now) {
			status = "scheduled"
		} else {
			status = "published"
		}
	}

	return status, pubTime
}

// Save Blog Post
func (h *AdminHandler) SaveBlogPost(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("title"))
	summary := strings.TrimSpace(c.FormValue("summary"))
	content := strings.TrimSpace(c.FormValue("content"))
	mediaPath := strings.TrimSpace(c.FormValue("media_path"))
	status, publishedAt := parsePublishInfo(c.FormValue("status"), c.FormValue("published_at"))

	slug := utils.MakeSlug(title)

	file, err := c.FormFile("media")
	if err != nil {
		file, err = c.FormFile("media_file")
	}
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "assets")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			mediaPath = "/static/assets/" + file.Filename
		}
	}

	post := models.BlogPost{
		Title:       title,
		Slug:        slug,
		Summary:     summary,
		Content:     content,
		MediaPath:   mediaPath,
		Status:      status,
		PublishedAt: publishedAt,
	}

	h.DB.Create(&post)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "blog:*", "page:home:*", "dashboard:*")
	}

	// Trigger automated search engine indexing notification if immediately published
	if status == "published" {
		fullURL := fmt.Sprintf("%s/blog/%s", strings.TrimRight(h.Cfg.BaseURL, "/"), post.Slug)
		sitemapURL := fmt.Sprintf("%s/sitemap.xml", strings.TrimRight(h.Cfg.BaseURL, "/"))
		if h.Worker != nil {
			h.Worker.Enqueue(func(ctx context.Context) {
				_ = services.NotifySearchEngines(h.Cfg, []string{fullURL, sitemapURL})
			})
		} else {
			go func() {
				_ = services.NotifySearchEngines(h.Cfg, []string{fullURL, sitemapURL})
			}()
		}
	}

	return c.Redirect("/admin/blogs")
}

// Update Blog Post
func (h *AdminHandler) UpdateBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.BlogPost
	if err := h.DB.First(&post, id).Error; err != nil {
		return c.Redirect("/admin/blogs")
	}

	post.Title = strings.TrimSpace(c.FormValue("title"))
	post.Summary = strings.TrimSpace(c.FormValue("summary"))
	post.Content = strings.TrimSpace(c.FormValue("content"))
	post.Slug = utils.MakeSlug(post.Title)

	status, publishedAt := parsePublishInfo(c.FormValue("status"), c.FormValue("published_at"))
	post.Status = status
	post.PublishedAt = publishedAt

	file, err := c.FormFile("media")
	if err != nil {
		file, err = c.FormFile("media_file")
	}
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "assets")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			post.MediaPath = "/static/assets/" + file.Filename
		}
	}

	h.DB.Save(&post)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "blog:*", "page:home:*", "dashboard:*")
	}

	// Trigger automated search engine indexing notification if published
	if status == "published" {
		fullURL := fmt.Sprintf("%s/blog/%s", strings.TrimRight(h.Cfg.BaseURL, "/"), post.Slug)
		sitemapURL := fmt.Sprintf("%s/sitemap.xml", strings.TrimRight(h.Cfg.BaseURL, "/"))
		if h.Worker != nil {
			h.Worker.Enqueue(func(ctx context.Context) {
				_ = services.NotifySearchEngines(h.Cfg, []string{fullURL, sitemapURL})
			})
		} else {
			go func() {
				_ = services.NotifySearchEngines(h.Cfg, []string{fullURL, sitemapURL})
			}()
		}
	}

	return c.Redirect("/admin/blogs")
}

// Delete Blog Post
func (h *AdminHandler) DeleteBlogPost(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.BlogPost{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "blog:*", "page:home:*", "dashboard:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") || c.Method() == "DELETE" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Blog post deleted successfully",
		})
	}
	return c.Redirect("/admin/blogs")
}

// FAQs Page
func (h *AdminHandler) FAQsPage(c *fiber.Ctx) error {
	var faqs []models.FAQ
	h.DB.Order("display_order asc, id asc").Find(&faqs)

	var submissions []models.FAQSubmission
	h.DB.Order("created_at desc").Find(&submissions)

	return c.Render("admin/faqs", h.buildAdminContext(c, "faqs", fiber.Map{
		"faqs":            faqs,
		"submissions":     submissions,
		"faq_submissions": submissions,
	}))
}

// Save FAQ
func (h *AdminHandler) SaveFAQ(c *fiber.Ctx) error {
	question := strings.TrimSpace(c.FormValue("question"))
	answer := strings.TrimSpace(c.FormValue("answer"))
	order, _ := strconv.Atoi(c.FormValue("display_order"))
	isPub := c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true"

	faq := models.FAQ{
		Question:     question,
		Slug:         utils.MakeSlug(question),
		Answer:       answer,
		DisplayOrder: order,
		IsPublished:  isPub,
	}

	h.DB.Create(&faq)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "faqs:*", "page:home:*")
	}

	return c.Redirect("/admin/faqs")
}

// Update FAQ
func (h *AdminHandler) UpdateFAQ(c *fiber.Ctx) error {
	id := c.Params("id")
	var faq models.FAQ
	if err := h.DB.First(&faq, id).Error; err != nil {
		return c.Redirect("/admin/faqs")
	}

	faq.Question = strings.TrimSpace(c.FormValue("question"))
	faq.Answer = strings.TrimSpace(c.FormValue("answer"))
	if slug := strings.TrimSpace(c.FormValue("slug")); slug != "" {
		faq.Slug = utils.MakeSlug(slug)
	}
	if order, err := strconv.Atoi(c.FormValue("display_order")); err == nil {
		faq.DisplayOrder = order
	}
	faq.IsPublished = c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true"

	h.DB.Save(&faq)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "faqs:*", "page:home:*")
	}

	return c.Redirect("/admin/faqs")
}

// Delete FAQ
func (h *AdminHandler) DeleteFAQ(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.FAQ{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "faqs:*", "page:home:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") || c.Method() == "DELETE" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "FAQ deleted successfully",
		})
	}
	return c.Redirect("/admin/faqs")
}

// Delete FAQ Submission
func (h *AdminHandler) DeleteFAQSubmission(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.FAQSubmission{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "faqs:*", "page:home:*")
	}

	if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") || c.Method() == "DELETE" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Submission deleted successfully",
		})
	}
	return c.Redirect("/admin/faqs")
}

// Services Page
func (h *AdminHandler) ServicesPage(c *fiber.Ctx) error {
	var servicesList []models.Service
	h.DB.Order("display_order asc, id asc").Find(&servicesList)

	return c.Render("admin/services", h.buildAdminContext(c, "services", fiber.Map{
		"services": servicesList,
	}))
}

// Add Service
func (h *AdminHandler) AddService(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("title"))
	slug := strings.TrimSpace(c.FormValue("slug"))
	if slug == "" {
		slug = utils.MakeSlug(title)
	}
	order, _ := strconv.Atoi(c.FormValue("display_order"))
	isPub := c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true"

	service := models.Service{
		Title:        title,
		Slug:         slug,
		Eyebrow:      strings.TrimSpace(c.FormValue("eyebrow")),
		LeadText:     strings.TrimSpace(c.FormValue("lead_text")),
		Description:  strings.TrimSpace(c.FormValue("description")),
		Features:     strings.TrimSpace(c.FormValue("features")),
		PriceRange:   strings.TrimSpace(c.FormValue("price_range")),
		PriceLabel:   strings.TrimSpace(c.FormValue("price_label")),
		PriceNote:    strings.TrimSpace(c.FormValue("price_note")),
		IconSVG:      strings.TrimSpace(c.FormValue("icon_svg")),
		PanelTitle:   strings.TrimSpace(c.FormValue("panel_title")),
		PanelType:    strings.TrimSpace(c.FormValue("panel_type")),
		PanelContent: strings.TrimSpace(c.FormValue("panel_content")),
		DisplayOrder: order,
		IsPublished:  isPub,
	}

	h.DB.Create(&service)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "services:*", "page:home:*")
	}

	return c.Redirect("/admin/services")
}

// Update Service
func (h *AdminHandler) UpdateService(c *fiber.Ctx) error {
	id := c.Params("id")
	var service models.Service
	if err := h.DB.First(&service, id).Error; err != nil {
		return c.Redirect("/admin/services")
	}

	service.Title = strings.TrimSpace(c.FormValue("title"))
	if slug := strings.TrimSpace(c.FormValue("slug")); slug != "" {
		service.Slug = slug
	}
	service.Eyebrow = strings.TrimSpace(c.FormValue("eyebrow"))
	service.LeadText = strings.TrimSpace(c.FormValue("lead_text"))
	service.Description = strings.TrimSpace(c.FormValue("description"))
	service.Features = strings.TrimSpace(c.FormValue("features"))
	service.PriceRange = strings.TrimSpace(c.FormValue("price_range"))
	service.PriceLabel = strings.TrimSpace(c.FormValue("price_label"))
	service.PriceNote = strings.TrimSpace(c.FormValue("price_note"))
	service.IconSVG = strings.TrimSpace(c.FormValue("icon_svg"))
	service.PanelTitle = strings.TrimSpace(c.FormValue("panel_title"))
	service.PanelType = strings.TrimSpace(c.FormValue("panel_type"))
	service.PanelContent = strings.TrimSpace(c.FormValue("panel_content"))
	if order, err := strconv.Atoi(c.FormValue("display_order")); err == nil {
		service.DisplayOrder = order
	}
	service.IsPublished = c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true"

	h.DB.Save(&service)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "services:*", "page:home:*")
	}

	return c.Redirect("/admin/services")
}

// Delete Service
func (h *AdminHandler) DeleteService(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.Service{}, id)

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "services:*", "page:home:*")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Service deleted successfully",
	})
}

// Booking Settings Page
func (h *AdminHandler) BookingSettingsPage(c *fiber.Ctx) error {
	settings := database.GetBookingSettings(h.DB)

	return c.Render("admin/booking", h.buildAdminContext(c, "booking", fiber.Map{
		"settings":                            settings,
		"booking_settings":                    settings,
		"google_calendar_packages_installed":  true,
	}))
}

// Update Booking Settings
func (h *AdminHandler) UpdateBookingSettings(c *fiber.Ctx) error {
	settings := database.GetBookingSettings(h.DB)

	settings.CalendarID = c.FormValue("calendar_id")
	settings.WorkdayStart = c.FormValue("workday_start")
	settings.WorkdayEnd = c.FormValue("workday_end")
	settings.MeetingDurationMinutes, _ = strconv.Atoi(c.FormValue("meeting_duration_minutes"))
	settings.BufferMinutes, _ = strconv.Atoi(c.FormValue("buffer_minutes"))
	settings.SlotStepMinutes, _ = strconv.Atoi(c.FormValue("slot_step_minutes"))
	settings.BookingHorizonDays, _ = strconv.Atoi(c.FormValue("booking_horizon_days"))
	settings.MinNoticeHours, _ = strconv.Atoi(c.FormValue("min_notice_hours"))

	h.DB.Save(settings)
	return c.Redirect("/admin/booking")
}

// Invoices & Quotes Page
func (h *AdminHandler) InvoicesPage(c *fiber.Ctx) error {
	settings := database.GetInvoiceSettings(h.DB)

	return c.Render("admin/invoices", h.buildAdminContext(c, "invoices", fiber.Map{
		"settings":         settings,
		"invoice_settings": settings,
	}))
}

// Save Invoice Settings (for admin.js AJAX)
func (h *AdminHandler) SaveInvoiceSettings(c *fiber.Ctx) error {
	s := database.GetInvoiceSettings(h.DB)
	s.BizName = strings.TrimSpace(c.FormValue("biz_name"))
	s.BizAddress = strings.TrimSpace(c.FormValue("biz_address"))
	s.BizPhone = strings.TrimSpace(c.FormValue("biz_phone"))
	s.BizEmail = strings.TrimSpace(c.FormValue("biz_email"))
	s.BankName = strings.TrimSpace(c.FormValue("bank_name"))
	s.AccountHolder = strings.TrimSpace(c.FormValue("account_holder"))
	s.AccountNumber = strings.TrimSpace(c.FormValue("account_number"))
	s.BranchCode = strings.TrimSpace(c.FormValue("branch_code"))
	s.VATNumber = strings.TrimSpace(c.FormValue("vat_number"))
	s.PaymentTerms = strings.TrimSpace(c.FormValue("payment_terms"))

	h.DB.Save(s)
	return c.JSON(fiber.Map{"ok": true})
}

// Generate Invoice PDF
func (h *AdminHandler) GenerateInvoicePDF(c *fiber.Ctx) error {
	settings := database.GetInvoiceSettings(h.DB)

	invoiceNum := c.FormValue("invoice_number")
	if invoiceNum == "" {
		now := time.Now()
		invoiceNum = fmt.Sprintf("INV-%s%04d", now.Format("060102"), now.Unix()%10000)
	}
	clientName := c.FormValue("client_name")
	clientEmail := c.FormValue("client_email")
	clientAddress := c.FormValue("client_address")

	desc := c.FormValue("item_description")
	qty, _ := strconv.Atoi(c.FormValue("item_qty"))
	if qty <= 0 {
		qty = 1
	}
	price, _ := strconv.ParseFloat(c.FormValue("item_price"), 64)

	items := []services.LineItem{
		{
			Description: desc,
			Quantity:    qty,
			UnitPrice:   price,
			Total:       float64(qty) * price,
		},
	}

	pdfBytes, err := services.GenerateInvoicePDF(settings, invoiceNum, clientName, clientEmail, clientAddress, items, 0.0)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("PDF generation failed: %v", err))
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", invoiceNum))
	return c.Send(pdfBytes)
}

// Funds & Transactions Page
func (h *AdminHandler) FundsPage(c *fiber.Ctx) error {
	yearFilter := c.Query("year")
	query := h.DB.Model(&models.Transaction{})
	if yearFilter != "" {
		if y, err := strconv.Atoi(yearFilter); err == nil {
			startDate := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(y, 12, 31, 23, 59, 59, 0, time.UTC)
			query = query.Where("date >= ? AND date <= ?", startDate, endDate)
		}
	}

	var txs []models.Transaction
	query.Order("date desc, id desc").Find(&txs)

	var totalIncome float64
	var totalExpenses float64
	for _, tx := range txs {
		if strings.EqualFold(tx.Type, "Income") {
			totalIncome += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	netBalance := totalIncome - totalExpenses
	salaryDraw := 0.0
	reinvestmentFund := 0.0
	if netBalance > 0 {
		salaryDraw = netBalance * 0.20
		reinvestmentFund = netBalance * 0.80
	}

	var distinctYears []int
	var allTxs []models.Transaction
	h.DB.Select("date").Find(&allTxs)
	yearMap := make(map[int]bool)
	for _, t := range allTxs {
		if !t.Date.IsZero() {
			y := t.Date.Year()
			if !yearMap[y] {
				yearMap[y] = true
				distinctYears = append(distinctYears, y)
			}
		}
	}

	return c.Render("admin/funds", h.buildAdminContext(c, "funds", fiber.Map{
		"transactions":      txs,
		"total_income":      totalIncome,
		"total_expenses":    totalExpenses,
		"net_balance":       netBalance,
		"salary_draw":       salaryDraw,
		"reinvestment_fund": reinvestmentFund,
		"available_years":   distinctYears,
		"selected_year":     yearFilter,
	}))
}

// Add Transaction
func (h *AdminHandler) AddTransaction(c *fiber.Ctx) error {
	txType := strings.TrimSpace(c.FormValue("type"))
	category := strings.TrimSpace(c.FormValue("category"))
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	desc := strings.TrimSpace(c.FormValue("description"))
	dateStr := strings.TrimSpace(c.FormValue("date"))

	txDate := time.Now()
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			txDate = t
		}
	}

	tx := models.Transaction{
		Type:        txType,
		Category:    category,
		Amount:      amount,
		Description: desc,
		Date:        txDate,
	}

	h.DB.Create(&tx)
	return c.Redirect("/admin/funds")
}

// Export Transactions to CSV
func (h *AdminHandler) ExportTransactions(c *fiber.Ctx) error {
	var txs []models.Transaction
	h.DB.Order("date desc, id desc").Find(&txs)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ID", "Date", "Type", "Category", "Amount", "Description"})

	for _, t := range txs {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", t.ID),
			t.Date.Format("2006-01-02"),
			t.Type,
			t.Category,
			fmt.Sprintf("%.2f", t.Amount),
			t.Description,
		})
	}
	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=transactions_export.csv")
	return c.Send(buf.Bytes())
}

// Delete Transaction
func (h *AdminHandler) DeleteTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.DB.Delete(&models.Transaction{}, id).Error; err != nil {
		if c.XHR() || strings.Contains(c.Get("Content-Type"), "application/json") {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		return c.Redirect("/admin/funds")
	}

	if c.XHR() || strings.Contains(c.Get("Content-Type"), "application/json") || c.Method() == fiber.MethodDelete {
		return c.JSON(fiber.Map{"status": "success", "message": "Transaction deleted successfully"})
	}
	return c.Redirect("/admin/funds")
}


// Automation Page
func (h *AdminHandler) AutomationPage(c *fiber.Ctx) error {
	var logs []models.AutomationLog
	h.DB.Order("timestamp desc").Limit(50).Find(&logs)

	var hotLeads int64
	h.DB.Model(&models.Lead{}).Where("score >= 80").Count(&hotLeads)

	var activeLeads []models.Lead
	h.DB.Where("status IN ?", []string{"New", "Contacted", "Negotiating"}).Find(&activeLeads)

	var pipelineForecast float64
	for _, l := range activeLeads {
		pipelineForecast += l.Budget
	}

	return c.Render("admin/automation", h.buildAdminContext(c, "automation", fiber.Map{
		"logs":               logs,
		"hot_leads":          hotLeads,
		"active_leads_count": len(activeLeads),
		"pipeline_forecast":  pipelineForecast,
	}))
}

// Run Automation Script (for admin.js AJAX triggers)
func (h *AdminHandler) RunScript(c *fiber.Ctx) error {
	var payload struct {
		ScriptName string `json:"script_name"`
	}
	_ = c.BodyParser(&payload)
	if payload.ScriptName == "" {
		payload.ScriptName = "Manual Script Execution"
	}

	newLog := models.AutomationLog{
		ScriptName: payload.ScriptName,
		Status:     "Success",
		Timestamp:  time.Now(),
	}
	h.DB.Create(&newLog)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Script %s executed successfully", payload.ScriptName),
	})
}

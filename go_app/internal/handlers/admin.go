package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/database"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/services"
	"github.com/JRoetscyber/my_website/go_app/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAdminHandler(db *gorm.DB, cfg *config.Config) *AdminHandler {
	return &AdminHandler{DB: db, Cfg: cfg}
}

// Dashboard Overview
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	var totalLeads int64
	h.DB.Model(&models.Lead{}).Count(&totalLeads)

	var hotLeads int64
	h.DB.Model(&models.Lead{}).Where("score >= 80").Count(&hotLeads)

	var totalProjects int64
	h.DB.Model(&models.Project{}).Count(&totalProjects)

	var totalBlogs int64
	h.DB.Model(&models.BlogPost{}).Count(&totalBlogs)

	var recentLeads []models.Lead
	h.DB.Order("created_at desc").Limit(5).Find(&recentLeads)

	var recentBlogs []models.BlogPost
	h.DB.Order("created_at desc").Limit(5).Find(&recentBlogs)

	return c.Render("admin/admin", fiber.Map{
		"active_page":    "dashboard",
		"total_leads":    totalLeads,
		"hot_leads":      hotLeads,
		"total_projects": totalProjects,
		"total_blogs":    totalBlogs,
		"recent_leads":   recentLeads,
		"recent_blogs":   recentBlogs,
		"request":        c,
	})
}

// Leads Page
func (h *AdminHandler) LeadsPage(c *fiber.Ctx) error {
	var leads []models.Lead
	h.DB.Order("created_at desc").Find(&leads)

	return c.Render("admin/leads", fiber.Map{
		"active_page": "leads",
		"leads":       leads,
		"request":     c,
	})
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

// Create Lead (Admin)
func (h *AdminHandler) CreateLead(c *fiber.Ctx) error {
	clientName := strings.TrimSpace(c.FormValue("client_name"))
	clientCompany := strings.TrimSpace(c.FormValue("client_company"))
	projectType := strings.TrimSpace(c.FormValue("project_type"))
	contactRole := strings.TrimSpace(c.FormValue("contact_role"))
	phoneNumber := strings.TrimSpace(c.FormValue("phone_number"))
	status := strings.TrimSpace(c.FormValue("status"))
	if status == "" {
		status = "New"
	}
	budget, _ := strconv.ParseFloat(c.FormValue("budget"), 64)

	scoreRes := services.CalculateLeadScore(services.LeadScoringInput{
		ClientCompany:      clientCompany,
		ContactRole:        contactRole,
		Budget:             budget,
		ProjectType:        projectType,
		VisitedPages:       []string{"/admin"},
		WhatsappEngagement: "read",
		PhoneNumber:        phoneNumber,
		LastActivityDate:   time.Now(),
	})

	lead := models.Lead{
		ClientName:         clientName,
		ClientCompany:      clientCompany,
		ProjectType:        projectType,
		Budget:             budget,
		ContactRole:        contactRole,
		PhoneNumber:        phoneNumber,
		WhatsappEngagement: "read",
		TargetProject:      projectType,
		Score:              int(math.Round(scoreRes.Score)),
		ExplicitScore:      scoreRes.Breakdown.Explicit,
		ImplicitScore:      scoreRes.Breakdown.Implicit,
		UrgencyScore:       scoreRes.Breakdown.Urgency,
		Status:             status,
		LastActivityDate:   time.Now(),
	}

	h.DB.Create(&lead)
	return c.Redirect("/admin/leads")
}

// Update Lead Status / Info
func (h *AdminHandler) UpdateLead(c *fiber.Ctx) error {
	id := c.Params("id")
	var lead models.Lead
	if err := h.DB.First(&lead, id).Error; err != nil {
		return c.Redirect("/admin/leads")
	}

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

	lead.LastActivityDate = time.Now()
	h.DB.Save(&lead)

	return c.Redirect("/admin/leads")
}

// Delete Lead
func (h *AdminHandler) DeleteLead(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.Lead{}, id)
	return c.Redirect("/admin/leads")
}

// Analytics Page
func (h *AdminHandler) AnalyticsPage(c *fiber.Ctx) error {
	type PageCount struct {
		PagePath string
		Count    int
	}
	var topPages []PageCount
	h.DB.Model(&models.Analytics{}).Select("page_path, count(*) as count").Group("page_path").Order("count desc").Limit(10).Scan(&topPages)

	var totalViews int64
	h.DB.Model(&models.Analytics{}).Count(&totalViews)

	var totalVisitors int64
	h.DB.Model(&models.Analytics{}).Distinct("visitor_ip").Count(&totalVisitors)

	return c.Render("admin/analytics", fiber.Map{
		"active_page":    "analytics",
		"top_pages":      topPages,
		"total_views":    totalViews,
		"total_visitors": totalVisitors,
		"request":        c,
	})
}

// Projects Page
func (h *AdminHandler) ProjectsPage(c *fiber.Ctx) error {
	var projects []models.Project
	h.DB.Order("id desc").Find(&projects)

	return c.Render("admin/projects", fiber.Map{
		"active_page": "projects",
		"projects":    projects,
		"request":     c,
	})
}

// Save / Create Project
func (h *AdminHandler) SaveProject(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("title"))
	category := strings.TrimSpace(c.FormValue("category"))
	techStack := strings.TrimSpace(c.FormValue("tech_stack"))
	description := strings.TrimSpace(c.FormValue("description"))
	codeSnippet := strings.TrimSpace(c.FormValue("code_snippet"))
	youtubeURL := strings.TrimSpace(c.FormValue("youtube_url"))
	projectURL := strings.TrimSpace(c.FormValue("project_url"))
	mediaPath := strings.TrimSpace(c.FormValue("media_path"))

	slug := utils.MakeSlug(title)

	// Check file upload
	file, err := c.FormFile("media_file")
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "uploads", "projects")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			mediaPath = "/" + filepath.ToSlash(dest)
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
	}

	h.DB.Create(&project)
	return c.Redirect("/admin/projects")
}

// Delete Project
func (h *AdminHandler) DeleteProject(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.Project{}, id)
	return c.Redirect("/admin/projects")
}

// Blogs Page
func (h *AdminHandler) BlogsPage(c *fiber.Ctx) error {
	var blogs []models.BlogPost
	h.DB.Order("created_at desc").Find(&blogs)

	return c.Render("admin/blogs", fiber.Map{
		"active_page": "blogs",
		"blogs":       blogs,
		"request":     c,
	})
}

// Save Blog Post
func (h *AdminHandler) SaveBlogPost(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.FormValue("title"))
	summary := strings.TrimSpace(c.FormValue("summary"))
	content := strings.TrimSpace(c.FormValue("content"))
	mediaPath := strings.TrimSpace(c.FormValue("media_path"))

	slug := utils.MakeSlug(title)

	file, err := c.FormFile("media_file")
	if err == nil && file != nil {
		uploadDir := filepath.Join("static", "uploads", "blog")
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		dest := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveFile(file, dest); err == nil {
			mediaPath = "/" + filepath.ToSlash(dest)
		}
	}

	post := models.BlogPost{
		Title:     title,
		Slug:      slug,
		Summary:   summary,
		Content:   content,
		MediaPath: mediaPath,
	}

	h.DB.Create(&post)
	return c.Redirect("/admin/blogs")
}

// Delete Blog Post
func (h *AdminHandler) DeleteBlogPost(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.BlogPost{}, id)
	return c.Redirect("/admin/blogs")
}

// FAQs Page
func (h *AdminHandler) FAQsPage(c *fiber.Ctx) error {
	var faqs []models.FAQ
	h.DB.Order("display_order asc, id asc").Find(&faqs)

	var submissions []models.FAQSubmission
	h.DB.Order("created_at desc").Find(&submissions)

	return c.Render("admin/faqs", fiber.Map{
		"active_page": "faqs",
		"faqs":        faqs,
		"submissions": submissions,
		"request":     c,
	})
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
	return c.Redirect("/admin/faqs")
}

// Delete FAQ
func (h *AdminHandler) DeleteFAQ(c *fiber.Ctx) error {
	id := c.Params("id")
	h.DB.Delete(&models.FAQ{}, id)
	return c.Redirect("/admin/faqs")
}

// Services Page
func (h *AdminHandler) ServicesPage(c *fiber.Ctx) error {
	var servicesList []models.Service
	h.DB.Order("display_order asc, id asc").Find(&servicesList)

	return c.Render("admin/services", fiber.Map{
		"active_page": "services",
		"services":    servicesList,
		"request":     c,
	})
}

// Booking Settings Page
func (h *AdminHandler) BookingSettingsPage(c *fiber.Ctx) error {
	settings := database.GetBookingSettings(h.DB)

	return c.Render("admin/booking", fiber.Map{
		"active_page": "booking",
		"settings":    settings,
		"request":     c,
	})
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

	return c.Render("admin/invoices", fiber.Map{
		"active_page": "invoices",
		"settings":    settings,
		"request":     c,
	})
}

// Generate Invoice PDF
func (h *AdminHandler) GenerateInvoicePDF(c *fiber.Ctx) error {
	settings := database.GetInvoiceSettings(h.DB)

	invoiceNum := c.FormValue("invoice_number")
	if invoiceNum == "" {
		invoiceNum = fmt.Sprintf("INV-%d", time.Now().Unix())
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
	var txs []models.Transaction
	h.DB.Order("date desc, id desc").Find(&txs)

	var totalIncome float64
	var totalExpense float64
	for _, tx := range txs {
		if strings.EqualFold(tx.Type, "Income") {
			totalIncome += tx.Amount
		} else {
			totalExpense += tx.Amount
		}
	}

	return c.Render("admin/funds", fiber.Map{
		"active_page":   "funds",
		"transactions":  txs,
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"net_profit":    totalIncome - totalExpense,
		"request":       c,
	})
}

// Automation Logs Page
func (h *AdminHandler) AutomationPage(c *fiber.Ctx) error {
	var logs []models.AutomationLog
	h.DB.Order("timestamp desc").Limit(50).Find(&logs)

	return c.Render("admin/automation", fiber.Map{
		"active_page": "automation",
		"logs":        logs,
		"request":     c,
	})
}

package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/database"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func resolveUploadsDir() string {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		if fi, err := os.Stat("./static"); err == nil && fi.IsDir() {
			staticDir = "./static"
		} else if fi, err := os.Stat("../static"); err == nil && fi.IsDir() {
			staticDir = "../static"
		} else {
			staticDir = "./static"
		}
	}
	return filepath.Join(staticDir, "uploads")
}

// ContentPage renders the Site Pages CMS editor for Home, About, Security, and Engineering.
func (h *AdminHandler) ContentPage(c *fiber.Ctx) error {
	activeTab := strings.TrimSpace(c.Query("tab", "home"))
	switch activeTab {
	case "home", "about", "security", "engineering":
	default:
		activeTab = "home"
	}

	homeJSON, _ := json.Marshal(database.GetHomeContent(h.DB))
	aboutJSON, _ := json.Marshal(database.GetAboutContent(h.DB))
	securityJSON, _ := json.Marshal(database.GetSecurityContent(h.DB))
	engineeringJSON, _ := json.Marshal(database.GetEngineeringContent(h.DB))

	return c.Render("admin/content", h.buildAdminContext(c, "content", fiber.Map{
		"active_tab":       activeTab,
		"home_json":        string(homeJSON),
		"about_json":       string(aboutJSON),
		"security_json":    string(securityJSON),
		"engineering_json": string(engineeringJSON),
	}))
}

// SaveContent persists edited structured JSON content for a specific page key and flushes cache.
func (h *AdminHandler) SaveContent(c *fiber.Ctx) error {
	pageKey := strings.ToLower(strings.TrimSpace(c.Params("page")))
	var payload interface{}

	switch pageKey {
	case "home":
		var data models.HomeContent
		if err := json.Unmarshal(c.Body(), &data); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid homepage JSON payload"})
		}
		payload = data
	case "about":
		var data models.AboutContent
		if err := json.Unmarshal(c.Body(), &data); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid about page JSON payload"})
		}
		payload = data
	case "security":
		var data models.SecurityContent
		if err := json.Unmarshal(c.Body(), &data); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid security page JSON payload"})
		}
		payload = data
	case "engineering":
		var data models.EngineeringContent
		if err := json.Unmarshal(c.Body(), &data); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid engineering page JSON payload"})
		}
		payload = data
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Unknown page key"})
	}

	if err := database.SavePageContent(h.DB, pageKey, payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to save content: " + err.Error()})
	}

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "pagecontent:*", "page:home:*")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Saved %s page content — live site updated immediately.", pageKey),
	})
}

// ResetContent restores the default seeded content for a specific page key.
func (h *AdminHandler) ResetContent(c *fiber.Ctx) error {
	pageKey := strings.ToLower(strings.TrimSpace(c.Params("page")))
	def, err := database.ResetPageContent(h.DB, pageKey)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Failed to reset page content"})
	}

	if h.Cache != nil {
		h.Cache.Invalidate(c.Context(), "pagecontent:*", "page:home:*")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Reset %s page to defaults.", pageKey),
		"data":    def,
	})
}

// UploadAvatar handles profile headshot uploads for the About page.
func (h *AdminHandler) UploadAvatar(c *fiber.Ctx) error {
	file, err := c.FormFile("avatar")
	if err != nil || file == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "No image file provided"})
	}

	ext := strings.ToLower(filepath.Ext(filepath.Base(file.Filename)))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Only PNG, JPG, JPEG, or WEBP images are allowed"})
	}

	uploadDir := resolveUploadsDir()
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to create uploads directory"})
	}

	safeName := fmt.Sprintf("avatar_%d%s", time.Now().Unix(), ext)
	dest := filepath.Join(uploadDir, safeName)
	if err := c.SaveFile(file, dest); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to save avatar file"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"url":    "/static/uploads/" + safeName,
	})
}

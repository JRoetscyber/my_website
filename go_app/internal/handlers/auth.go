package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/middleware"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/JRoetscyber/my_website/go_app/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

func (h *AuthHandler) LoginPage(c *fiber.Ctx) error {
	// If already authenticated, redirect to admin
	cookie := c.Cookies("jo4_session")
	if cookie != "" {
		if _, valid := middleware.VerifyAuthToken(cookie, h.Cfg.SecretKey); valid {
			return c.Redirect("/admin")
		}
	}

	return c.Render("login", fiber.Map{
		"error":   nil,
		"request": c,
	})
}

func (h *AuthHandler) LoginSubmit(c *fiber.Ctx) error {
	username := strings.TrimSpace(c.FormValue("username"))
	password := strings.TrimSpace(c.FormValue("password"))

	if username == "" || password == "" {
		return c.Status(http.StatusBadRequest).Render("login", fiber.Map{
			"error":   "Please enter both username and password",
			"request": c,
		})
	}

	var user models.User
	if err := h.DB.Where("LOWER(username) = LOWER(?)", username).First(&user).Error; err != nil {
		log.Printf("[AUTH] Login failed: User '%s' not found", username)
		return c.Status(http.StatusUnauthorized).Render("login", fiber.Map{
			"error":   "Invalid username or password",
			"request": c,
		})
	}

	if !utils.CheckPasswordHash(user.PasswordHash, password) {
		log.Printf("[AUTH] Login failed: Incorrect password for user '%s'", username)
		return c.Status(http.StatusUnauthorized).Render("login", fiber.Map{
			"error":   "Invalid username or password",
			"request": c,
		})
	}

	// If password hash was from legacy Flask (Werkzeug) or plaintext, upgrade to bcrypt
	if !strings.HasPrefix(user.PasswordHash, "$2a$") && !strings.HasPrefix(user.PasswordHash, "$2b$") {
		if newHash, err := utils.HashPassword(password); err == nil {
			h.DB.Model(&user).Update("password_hash", newHash)
			log.Printf("[AUTH] Upgraded legacy password hash to Bcrypt for user '%s'", user.Username)
		}
	}

	// Generate session token
	token := middleware.GenerateAuthToken(user.Username, h.Cfg.SecretKey)

	// Set session cookie with root path so it applies across /admin and all subpaths
	c.Cookie(&fiber.Cookie{
		Name:     "jo4_session",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false, // Keep false so it works across HTTP and HTTPS reverse proxies
	})

	log.Printf("[AUTH] User '%s' logged in successfully", user.Username)
	return c.Redirect("/admin")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "jo4_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Redirect("/login")
}

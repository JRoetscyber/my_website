package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
	"github.com/JRoetscyber/my_website/go_app/internal/middleware"
	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

	var user models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(http.StatusUnauthorized).Render("login", fiber.Map{
			"error":   "Invalid username or password",
			"request": c,
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return c.Status(http.StatusUnauthorized).Render("login", fiber.Map{
			"error":   "Invalid username or password",
			"request": c,
		})
	}

	// Generate session token
	token := middleware.GenerateAuthToken(user.Username, h.Cfg.SecretKey)

	c.Cookie(&fiber.Cookie{
		Name:     "jo4_session",
		Value:    token,
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false, // Set to true in production HTTPS
	})

	return c.Redirect("/admin")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.ClearCookie("jo4_session")
	return c.Redirect("/login")
}

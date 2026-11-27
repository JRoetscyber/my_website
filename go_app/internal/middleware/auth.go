package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GenerateAuthToken(username, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(username))
	signature := hex.EncodeToString(mac.Sum(nil))
	return username + ":" + signature
}

func VerifyAuthToken(token, secret string) (string, bool) {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return "", false
	}
	username := parts[0]
	signature := parts[1]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(username))
	expected := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(signature), []byte(expected)) {
		return username, true
	}
	return "", false
}

func isAJAXOrAPI(c *fiber.Ctx) bool {
	p := c.Path()
	return strings.HasPrefix(p, "/admin/api/") ||
		c.XHR() ||
		strings.Contains(c.Get("Accept"), "application/json") ||
		strings.Contains(c.Get("Content-Type"), "application/json")
}

func RequireAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cookie := c.Cookies("jo4_session")
		if cookie == "" {
			if isAJAXOrAPI(c) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  "error",
					"message": "Session expired or not logged in. Please sign in again.",
				})
			}
			return c.Redirect("/login")
		}

		username, valid := VerifyAuthToken(cookie, secret)
		if !valid {
			c.ClearCookie("jo4_session")
			if isAJAXOrAPI(c) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid or expired session. Please sign in again.",
				})
			}
			return c.Redirect("/login")
		}

		c.Locals("user", username)
		return c.Next()
	}
}

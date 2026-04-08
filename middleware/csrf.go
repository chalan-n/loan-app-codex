package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const csrfCookieName = "csrf_token"

func CSRFProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := ensureCSRFCookie(c)
		c.Locals("csrfToken", token)

		if !requiresCSRFFProtection(c) {
			return c.Next()
		}

		if c.Get("X-API-Key") != "" {
			return c.Next()
		}

		if hasValidCSRFTokens(c) {
			return c.Next()
		}

		if hasTrustedFetchMetadata(c) {
			return c.Next()
		}

		if hasTrustedOrigin(c) || hasTrustedReferer(c) {
			return c.Next()
		}

		if c.Get("Accept") == "application/json" || strings.HasPrefix(c.Path(), "/api/") || c.Method() != fiber.MethodGet {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF validation failed",
			})
		}

		return c.Status(fiber.StatusForbidden).SendString("CSRF validation failed")
	}
}

func requiresCSRFFProtection(c *fiber.Ctx) bool {
	switch c.Method() {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		return true
	case fiber.MethodGet:
		return c.Path() == "/logout"
	default:
		return false
	}
}

func ensureCSRFCookie(c *fiber.Ctx) string {
	if existing := c.Cookies(csrfCookieName); existing != "" {
		return existing
	}

	token := randomCSRFToken()
	c.Cookie(&fiber.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		HTTPOnly: false,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})
	return token
}

func randomCSRFToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func hasValidCSRFTokens(c *fiber.Ctx) bool {
	cookieToken := c.Cookies(csrfCookieName)
	if cookieToken == "" {
		return false
	}

	headerToken := c.Get("X-CSRF-Token")
	if headerToken != "" && headerToken == cookieToken {
		return true
	}

	formToken := c.FormValue("_csrf")
	return formToken != "" && formToken == cookieToken
}

func hasTrustedFetchMetadata(c *fiber.Ctx) bool {
	site := strings.ToLower(strings.TrimSpace(c.Get("Sec-Fetch-Site")))
	return site == "same-origin" || site == "none"
}

func hasTrustedOrigin(c *fiber.Ctx) bool {
	origin := c.Get("Origin")
	if origin == "" {
		return false
	}
	return requestMatchesURLHost(c, origin)
}

func hasTrustedReferer(c *fiber.Ctx) bool {
	referer := c.Get("Referer")
	if referer == "" {
		return false
	}
	return requestMatchesURLHost(c, referer)
}

func requestMatchesURLHost(c *fiber.Ctx, raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return false
	}

	requestHost := normalizedHost(c.Hostname())
	parsedHost := normalizedHost(parsed.Hostname())
	return requestHost != "" && parsedHost != "" && requestHost == parsedHost
}

func normalizedHost(host string) string {
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.ToLower(parsedHost)
	}
	return strings.ToLower(host)
}

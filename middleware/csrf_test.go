package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCSRFSetsCookieOnSafeRequest(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		t.Fatal("expected CSRF cookie to be set")
	}
}

func TestCSRFAllowsSameOriginPost(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/step1", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "https://example.com/step1", nil)
	req.Host = "example.com"
	req.Header.Set("Origin", "https://example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestCSRFBLocksCrossSitePost(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/step1", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "https://example.com/step1", nil)
	req.Host = "example.com"
	req.Header.Set("Origin", "https://evil.example")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func TestCSRFAllowsMatchingTokenHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/api/delete-loan", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "https://example.com/api/delete-loan", nil)
	req.Host = "example.com"
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "token-123"})
	req.Header.Set("X-CSRF-Token", "token-123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestCSRFBLocksCrossSiteLogoutGet(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Get("/logout", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/logout", nil)
	req.Host = "example.com"
	req.Header.Set("Referer", "https://evil.example/path")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func TestCSRFSkipsAPIKeyRequests(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/api/update-status", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/update-status", nil)
	req.Header.Set("X-API-Key", "mobile-secret")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

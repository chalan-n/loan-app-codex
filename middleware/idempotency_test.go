package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestScopedIdempotencyKeyUsesUsernameWhenPresent(t *testing.T) {
	app := fiber.New()
	app.Get("/step1", func(c *fiber.Ctx) error {
		c.Locals("username", "570639")
		got := scopedIdempotencyKey(c, "abc123")
		want := "GET:/step1:570639:abc123"
		if got != want {
			t.Fatalf("scopedIdempotencyKey() = %q, want %q", got, want)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/step1", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

func TestScopedIdempotencyKeyFallsBackToAPIKey(t *testing.T) {
	app := fiber.New()
	app.Post("/api/update-status", func(c *fiber.Ctx) error {
		got := scopedIdempotencyKey(c, "dup-key")
		want := "POST:/api/update-status:mobile-secret:dup-key"
		if got != want {
			t.Fatalf("scopedIdempotencyKey() = %q, want %q", got, want)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("POST", "/api/update-status", nil)
	req.Header.Set("X-API-Key", "mobile-secret")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

func TestScopedIdempotencyKeyKeepsPathScopeWithoutIdentity(t *testing.T) {
	app := fiber.New()
	app.Post("/step2", func(c *fiber.Ctx) error {
		got := scopedIdempotencyKey(c, "same-key")
		want := "POST:/step2:same-key"
		if got != want {
			t.Fatalf("scopedIdempotencyKey() = %q, want %q", got, want)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("POST", "/step2", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestRateLimitBlocksAfterMaxRequests(t *testing.T) {
	currentTime := time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)

	app := fiber.New()
	app.Use(RateLimit(RateLimitOptions{
		Name:   "search",
		Max:    2,
		Window: time.Minute,
		Now: func() time.Time {
			return currentTime
		},
	}))
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for i := 0; i < 2; i++ {
		resp := performRateLimitRequest(t, app, "POST", "/", "")
		if resp.Code != fiber.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, app, "POST", "/", "")
	if resp.Code != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.Code)
	}

	if got := resp.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("expected Retry-After=60, got %q", got)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "Too many requests" {
		t.Fatalf("unexpected error body: %#v", body)
	}
}

func TestRateLimitSeparatesUsers(t *testing.T) {
	currentTime := time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("username", c.Get("X-Test-User"))
		return c.Next()
	})
	app.Use(RateLimit(RateLimitOptions{
		Name:   "search",
		Max:    1,
		Window: time.Minute,
		Now: func() time.Time {
			return currentTime
		},
	}))
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	firstAlice := performRateLimitRequest(t, app, "POST", "/", "alice")
	if firstAlice.Code != fiber.StatusOK {
		t.Fatalf("alice first request: expected 200, got %d", firstAlice.Code)
	}

	firstBob := performRateLimitRequest(t, app, "POST", "/", "bob")
	if firstBob.Code != fiber.StatusOK {
		t.Fatalf("bob first request: expected 200, got %d", firstBob.Code)
	}

	secondAlice := performRateLimitRequest(t, app, "POST", "/", "alice")
	if secondAlice.Code != fiber.StatusTooManyRequests {
		t.Fatalf("alice second request: expected 429, got %d", secondAlice.Code)
	}
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	currentTime := time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)

	app := fiber.New()
	app.Use(RateLimit(RateLimitOptions{
		Name:   "upload",
		Max:    1,
		Window: time.Minute,
		Now: func() time.Time {
			return currentTime
		},
	}))
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	first := performRateLimitRequest(t, app, "POST", "/", "")
	if first.Code != fiber.StatusOK {
		t.Fatalf("first request: expected 200, got %d", first.Code)
	}

	second := performRateLimitRequest(t, app, "POST", "/", "")
	if second.Code != fiber.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", second.Code)
	}

	currentTime = currentTime.Add(time.Minute + time.Second)
	third := performRateLimitRequest(t, app, "POST", "/", "")
	if third.Code != fiber.StatusOK {
		t.Fatalf("third request after reset: expected 200, got %d", third.Code)
	}
}

func TestLoginRateLimitExceededRedirectsToLogin(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimit(RateLimitOptions{
		Name:         "login",
		Max:          1,
		Window:       time.Minute,
		BlockHandler: LoginRateLimitExceeded,
	}))
	app.Post("/login", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	first := performRateLimitRequest(t, app, "POST", "/login", "")
	if first.Code != fiber.StatusOK {
		t.Fatalf("first login request: expected 200, got %d", first.Code)
	}

	second := performRateLimitRequest(t, app, "POST", "/login", "")
	if second.Code != fiber.StatusFound {
		t.Fatalf("second login request: expected 302, got %d", second.Code)
	}
	if got := second.Header().Get("Location"); got != "/login?error=too_many_requests" {
		t.Fatalf("expected login redirect, got %q", got)
	}
}

func performRateLimitRequest(t *testing.T, app *fiber.App, method, path, user string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "203.0.113.10:12345"
	if user != "" {
		req.Header.Set("X-Test-User", user)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	recorder := httptest.NewRecorder()
	recorder.Code = resp.StatusCode
	for key, values := range resp.Header {
		for _, value := range values {
			recorder.Header().Add(key, value)
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	_, _ = recorder.Body.Write(body)
	_ = resp.Body.Close()
	return recorder
}

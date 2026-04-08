package middleware

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RateLimitBlockFunc func(c *fiber.Ctx, retryAfter time.Duration) error

type RateLimitOptions struct {
	Name         string
	Max          int
	Window       time.Duration
	Now          func() time.Time
	BlockHandler RateLimitBlockFunc
}

type rateLimitEntry struct {
	Count   int
	ResetAt time.Time
}

type rateLimitStore struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
}

func RateLimit(opts RateLimitOptions) fiber.Handler {
	if opts.Max <= 0 || opts.Window <= 0 {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	store := &rateLimitStore{
		entries: make(map[string]rateLimitEntry),
	}

	return func(c *fiber.Ctx) error {
		now := nowFn()
		key := buildRateLimitKey(c, opts.Name)
		allowed, retryAfter := store.allow(key, opts.Max, opts.Window, now)
		if allowed {
			return c.Next()
		}

		retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
		if retryAfterSeconds < 1 {
			retryAfterSeconds = 1
		}
		c.Set("Retry-After", strconv.Itoa(retryAfterSeconds))

		if opts.BlockHandler != nil {
			return opts.BlockHandler(c, retryAfter)
		}

		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":       "Too many requests",
			"retry_after": retryAfterSeconds,
		})
	}
}

func (s *rateLimitStore) allow(key string, max int, window time.Duration, now time.Time) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := s.entries[key]
	if entry.ResetAt.IsZero() || !now.Before(entry.ResetAt) {
		entry = rateLimitEntry{
			Count:   0,
			ResetAt: now.Add(window),
		}
	}

	entry.Count++
	s.entries[key] = entry

	if entry.Count > max {
		return false, entry.ResetAt.Sub(now)
	}

	return true, 0
}

func buildRateLimitKey(c *fiber.Ctx, name string) string {
	return strings.Join([]string{
		name,
		rateLimitIdentity(c),
	}, "|")
}

func rateLimitIdentity(c *fiber.Ctx) string {
	if username, ok := c.Locals("username").(string); ok && username != "" {
		return "user:" + strings.ToLower(strings.TrimSpace(username))
	}

	if apiKey := strings.TrimSpace(c.Get("X-API-Key")); apiKey != "" {
		return "api-key:" + apiKey
	}

	if clientIP := strings.TrimSpace(c.IP()); clientIP != "" {
		return "ip:" + clientIP
	}

	return "ip:unknown"
}

func LoginRateLimitExceeded(c *fiber.Ctx, _ time.Duration) error {
	return c.Redirect("/login?error=too_many_requests")
}

package middleware

import (
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// cachedResponse stores the HTTP response for replay.
type cachedResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
	CreatedAt   time.Time
}

var (
	store   = make(map[string]*cachedResponse)
	pending = make(map[string]struct{}) // keys currently being processed
	mu      sync.Mutex
	ttl     = 24 * time.Hour
)

func init() {
	// Background goroutine: evict expired entries every 10 minutes.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for k, v := range store {
				if now.Sub(v.CreatedAt) > ttl {
					delete(store, k)
				}
			}
			mu.Unlock()
		}
	}()
}

// Idempotency returns a Fiber middleware that deduplicates requests
// using the X-Idempotency-Key header.
//
// Behaviour:
//  1. No header → skip (non-idempotent request, pass through).
//  2. Key exists in cache → return cached response immediately.
//  3. Key is currently being processed → 409 Conflict (prevents double-click).
//  4. New key → process request, cache the response, return it.
func Idempotency() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-Idempotency-Key")
		if key == "" {
			return c.Next()
		}

		key = scopedIdempotencyKey(c, key)

		mu.Lock()

		// ── Hit: return cached response ──
		if cached, ok := store[key]; ok {
			mu.Unlock()
			log.Printf("[idempotency] cache-hit key=%s", key)
			c.Set("X-Idempotent-Replayed", "true")
			c.Response().Header.SetContentType(cached.ContentType)
			return c.Status(cached.StatusCode).Send(cached.Body)
		}

		// ── In-flight: another request with the same key is still processing ──
		if _, inFlight := pending[key]; inFlight {
			mu.Unlock()
			log.Printf("[idempotency] in-flight duplicate key=%s", key)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "request with this idempotency key is already being processed",
			})
		}

		// ── New key: mark as pending and process ──
		pending[key] = struct{}{}
		mu.Unlock()

		// Execute the actual handler.
		err := c.Next()

		mu.Lock()
		delete(pending, key)

		// Only cache successful (2xx) responses.
		status := c.Response().StatusCode()
		if status >= 200 && status < 300 {
			store[key] = &cachedResponse{
				StatusCode:  status,
				ContentType: string(c.Response().Header.ContentType()),
				Body:        append([]byte(nil), c.Response().Body()...), // copy
				CreatedAt:   time.Now(),
			}
			log.Printf("[idempotency] cached key=%s status=%d", key, status)
		}
		mu.Unlock()

		return err
	}
}

func scopedIdempotencyKey(c *fiber.Ctx, key string) string {
	scope := c.Method() + ":" + c.Path()
	if username, ok := c.Locals("username").(string); ok && username != "" {
		scope += ":" + username
	} else if apiKey := c.Get("X-API-Key"); apiKey != "" {
		scope += ":" + apiKey
	}
	return scope + ":" + key
}

// StoreSize returns the current number of cached entries (for monitoring).
func StoreSize() int {
	mu.Lock()
	defer mu.Unlock()
	return len(store)
}

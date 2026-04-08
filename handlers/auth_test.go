package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"loan-app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestCreateJWTTokenStoresUsernameAndSessionID(t *testing.T) {
	prevNow := sessionNow
	fixedNow := time.Now().UTC().Truncate(time.Second)
	sessionNow = func() time.Time { return fixedNow }
	t.Cleanup(func() { sessionNow = prevNow })

	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken returned error: %v", err)
	}

	if got := parseJWTUsername(token); got != "570639" {
		t.Fatalf("parseJWTUsername() = %q, want %q", got, "570639")
	}

	if got := parseJWTSessionID(token); got != "session-123" {
		t.Fatalf("parseJWTSessionID() = %q, want %q", got, "session-123")
	}

	claims, ok := parseJWTClaims(token)
	if !ok {
		t.Fatal("parseJWTClaims() reported invalid token")
	}

	if _, ok := claims["exp"].(float64); !ok {
		t.Fatalf("expected exp claim to be present, got %#v", claims["exp"])
	}
	if got, ok := parseJWTIssuedAt(token); !ok || !got.Equal(fixedNow) {
		t.Fatalf("parseJWTIssuedAt() = %v, %v, want %v, true", got, ok, fixedNow)
	}
}

func TestParseJWTClaimsRejectsUnexpectedSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"username":   "570639",
		"session_id": "session-123",
	})

	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build unsigned token: %v", err)
	}

	if claims, ok := parseJWTClaims(tokenStr); ok || claims != nil {
		t.Fatalf("parseJWTClaims() accepted non-HMAC token: %#v", claims)
	}
}

func TestClearAuthCookieExpiresToken(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		clearAuthCookie(c)
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	cookie := resp.Header.Get("Set-Cookie")
	cookieLower := strings.ToLower(cookie)
	for _, want := range []string{
		"token=",
		"expires=",
		"httponly",
		"secure",
		"samesite=lax",
	} {
		if !strings.Contains(cookieLower, strings.ToLower(want)) {
			t.Fatalf("Set-Cookie %q does not contain %q", cookie, want)
		}
	}
}

func TestSetAuthCookieSetsSecurityFlags(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		setAuthCookie(c, "token-value")
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	cookie := resp.Header.Get("Set-Cookie")
	cookieLower := strings.ToLower(cookie)
	for _, want := range []string{
		"token=token-value",
		"httponly",
		"secure",
		"samesite=lax",
		"Path=/",
	} {
		if !strings.Contains(cookieLower, strings.ToLower(want)) {
			t.Fatalf("Set-Cookie %q does not contain %q", cookie, want)
		}
	}
}

func testHTTPCookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value}
}

func TestAuthMiddlewareRejectsExpiredIdleSession(t *testing.T) {
	prevLoad := loadSessionUser
	prevRevoke := revokeAllSessionsForUser
	prevNow := sessionNow
	prevTimeout := sessionIdleTimeout
	prevRoleLookup := lookupUserRole
	prevAuditWriter := auditLogWriter
	t.Cleanup(func() {
		loadSessionUser = prevLoad
		revokeAllSessionsForUser = prevRevoke
		sessionNow = prevNow
		sessionIdleTimeout = prevTimeout
		lookupUserRole = prevRoleLookup
		auditLogWriter = prevAuditWriter
	})

	now := time.Now().UTC().Truncate(time.Second)
	sessionNow = func() time.Time { return now }
	sessionIdleTimeout = func() time.Duration { return 30 * time.Minute }
	lookupUserRole = func(username string) string { return models.RoleOfficer }
	auditLogWriter = func(entry *models.AuditLog) {}
	loadSessionUser = func(username string) (*models.User, error) {
		lastSeen := now.Add(-31 * time.Minute)
		return &models.User{
			Username:              username,
			CurrentSessionID:      "session-123",
			SessionLastActivityAt: &lastSeen,
		}, nil
	}

	var revokedUsername string
	revokeAllSessionsForUser = func(username string, when time.Time) error {
		revokedUsername = username
		return nil
	}

	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}

	app := fiber.New()
	app.Use(AuthMiddleware)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(testHTTPCookie("token", token))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusFound)
	}
	if got := resp.Header.Get("Location"); got != "/login?error=session_expired" {
		t.Fatalf("redirect = %q, want %q", got, "/login?error=session_expired")
	}
	if revokedUsername != "570639" {
		t.Fatalf("revokeAllSessionsForUser username = %q, want %q", revokedUsername, "570639")
	}
}

func TestAuthMiddlewareRefreshesActiveSession(t *testing.T) {
	prevLoad := loadSessionUser
	prevTouch := touchSessionActivity
	prevNow := sessionNow
	prevTimeout := sessionIdleTimeout
	prevRefresh := sessionActivityRefreshInterval
	t.Cleanup(func() {
		loadSessionUser = prevLoad
		touchSessionActivity = prevTouch
		sessionNow = prevNow
		sessionIdleTimeout = prevTimeout
		sessionActivityRefreshInterval = prevRefresh
	})

	now := time.Now().UTC().Truncate(time.Second)
	sessionNow = func() time.Time { return now }
	sessionIdleTimeout = func() time.Duration { return 30 * time.Minute }
	sessionActivityRefreshInterval = func() time.Duration { return 5 * time.Minute }
	loadSessionUser = func(username string) (*models.User, error) {
		lastSeen := now.Add(-10 * time.Minute)
		return &models.User{
			Username:              username,
			CurrentSessionID:      "session-123",
			SessionLastActivityAt: &lastSeen,
		}, nil
	}

	var touchedUsername string
	touchSessionActivity = func(username string, when time.Time) error {
		touchedUsername = username
		return nil
	}

	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}

	app := fiber.New()
	app.Use(AuthMiddleware)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(testHTTPCookie("token", token))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if touchedUsername != "570639" {
		t.Fatalf("touchSessionActivity username = %q, want %q", touchedUsername, "570639")
	}
}

func TestAuthMiddlewareRejectsRevokedSession(t *testing.T) {
	prevLoad := loadSessionUser
	prevNow := sessionNow
	t.Cleanup(func() {
		loadSessionUser = prevLoad
		sessionNow = prevNow
	})

	issuedAt := time.Now().UTC().Truncate(time.Second)
	sessionNow = func() time.Time { return issuedAt }
	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}

	loadSessionUser = func(username string) (*models.User, error) {
		revokedAt := issuedAt.Add(2 * time.Minute)
		lastSeen := issuedAt.Add(1 * time.Minute)
		return &models.User{
			Username:              username,
			CurrentSessionID:      "session-123",
			SessionLastActivityAt: &lastSeen,
			SessionRevokedAt:      &revokedAt,
		}, nil
	}

	app := fiber.New()
	app.Use(AuthMiddleware)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(testHTTPCookie("token", token))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusFound)
	}
	if got := resp.Header.Get("Location"); got != "/login" {
		t.Fatalf("redirect = %q, want %q", got, "/login")
	}
}

func TestRevokeAllSessionsHandlerClearsCookie(t *testing.T) {
	prevRevoke := revokeAllSessionsForUser
	prevNow := sessionNow
	prevRoleLookup := lookupUserRole
	prevAuditWriter := auditLogWriter
	t.Cleanup(func() {
		revokeAllSessionsForUser = prevRevoke
		sessionNow = prevNow
		lookupUserRole = prevRoleLookup
		auditLogWriter = prevAuditWriter
	})

	now := time.Now().UTC().Truncate(time.Second)
	sessionNow = func() time.Time { return now }
	lookupUserRole = func(username string) string { return models.RoleOfficer }
	auditLogWriter = func(entry *models.AuditLog) {}

	var revokedUsername string
	revokeAllSessionsForUser = func(username string, when time.Time) error {
		revokedUsername = username
		return nil
	}

	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}

	app := fiber.New()
	app.Post("/session/revoke-all", RevokeAllSessions)

	req := httptest.NewRequest("POST", "/session/revoke-all", nil)
	req.AddCookie(testHTTPCookie("token", token))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if revokedUsername != "570639" {
		t.Fatalf("revokeAllSessionsForUser username = %q, want %q", revokedUsername, "570639")
	}
	if cookie := strings.ToLower(resp.Header.Get("Set-Cookie")); !strings.Contains(cookie, "token=") || !strings.Contains(cookie, "expires=") {
		t.Fatalf("expected revoke-all response to clear cookie, got %q", cookie)
	}
}

func TestRevokeAllSessionsHandlerReturnsServerError(t *testing.T) {
	prevRevoke := revokeAllSessionsForUser
	prevRoleLookup := lookupUserRole
	prevAuditWriter := auditLogWriter
	t.Cleanup(func() {
		revokeAllSessionsForUser = prevRevoke
		lookupUserRole = prevRoleLookup
		auditLogWriter = prevAuditWriter
	})

	revokeAllSessionsForUser = func(username string, when time.Time) error {
		return errors.New("boom")
	}
	lookupUserRole = func(username string) string { return models.RoleOfficer }
	auditLogWriter = func(entry *models.AuditLog) {}

	token, err := createJWTToken("570639", "session-123")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}

	app := fiber.New()
	app.Post("/session/revoke-all", RevokeAllSessions)

	req := httptest.NewRequest("POST", "/session/revoke-all", nil)
	req.AddCookie(testHTTPCookie("token", token))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusInternalServerError)
	}
}

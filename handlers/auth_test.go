package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestCreateJWTTokenStoresUsernameAndSessionID(t *testing.T) {
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

package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	loadLoanApplication = func(loanID interface{}) (*models.LoanApplication, error) {
		var loan models.LoanApplication
		if err := config.DB.First(&loan, loanID).Error; err != nil {
			return nil, err
		}
		return &loan, nil
	}
	lookupUserRole = func(username string) string {
		return getUserRole(username)
	}
)

func logDeniedLoanAccess(c *fiber.Ctx, username string, loanID interface{}, reason string) {
	WriteAuditAs(c, username, "deny_loan_access", "", "loan_id="+fmt.Sprint(loanID)+" reason="+reason)
}

func logDeniedFileAccess(c *fiber.Ctx, username, filename, reason string) {
	WriteAuditAs(c, username, "deny_file_access", "", "filename="+filename+" reason="+reason)
}

func parseJWTClaims(tokenStr string) (jwt.MapClaims, bool) {
	if tokenStr == "" {
		return nil, false
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.GetConfig().JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}

// parseJWTUsername extracts the "username" claim from a JWT token string.
// Returns an empty string if the token is invalid, expired, or empty.
func parseJWTUsername(tokenStr string) string {
	claims, ok := parseJWTClaims(tokenStr)
	if !ok {
		return ""
	}
	u, _ := claims["username"].(string)
	return u
}

func parseJWTSessionID(tokenStr string) string {
	claims, ok := parseJWTClaims(tokenStr)
	if !ok {
		return ""
	}
	sessionID, _ := claims["session_id"].(string)
	return sessionID
}

func newSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func issueUserSession(user *models.User) (string, error) {
	sessionID := newSessionID()
	if sessionID == "" {
		return "", errors.New("failed to generate session id")
	}

	user.CurrentSessionID = sessionID
	if err := config.DB.Model(user).
		Where("id = ?", user.ID).
		Update("current_session_id", sessionID).Error; err != nil {
		return "", err
	}

	return createJWTToken(user.Username, sessionID)
}

func setAuthCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})
}

func clearAuthCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	})
}

func requireLoanAccess(c *fiber.Ctx, loanID interface{}) (*models.LoanApplication, error) {
	username := parseJWTUsername(c.Cookies("token"))
	if username == "" {
		logDeniedLoanAccess(c, "", loanID, "missing_auth")
		return nil, fiber.ErrUnauthorized
	}

	loan, err := loadLoanApplication(loanID)
	if err != nil {
		logDeniedLoanAccess(c, username, loanID, "not_found")
		return nil, fiber.ErrNotFound
	}

	role := lookupUserRole(username)
	if role == models.RoleAdmin || role == models.RoleManager || loan.StaffID == username {
		return loan, nil
	}

	logDeniedLoanAccess(c, username, loanID, "forbidden")
	return nil, fiber.ErrForbidden
}

func loanIDFromFilename(filename string) (int, bool) {
	prefix, _, found := strings.Cut(filename, "_")
	if !found || prefix == "" {
		return 0, false
	}

	loanID, err := strconv.Atoi(prefix)
	if err != nil || loanID <= 0 {
		return 0, false
	}

	return loanID, true
}

func loanHasFile(loan *models.LoanApplication, filename string) bool {
	for _, file := range strings.Split(loan.CarInsuranceFile, ",") {
		if strings.TrimSpace(file) == filename {
			return true
		}
	}
	return false
}

func requireFileAccess(c *fiber.Ctx, filename string) (*models.LoanApplication, error) {
	username := parseJWTUsername(c.Cookies("token"))
	loanID, ok := loanIDFromFilename(filename)
	if !ok {
		logDeniedFileAccess(c, username, filename, "invalid_filename")
		return nil, fiber.ErrNotFound
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		if err == fiber.ErrUnauthorized {
			logDeniedFileAccess(c, username, filename, "missing_auth")
		}
		return nil, err
	}

	if !loanHasFile(loan, filename) {
		logDeniedFileAccess(c, username, filename, "file_not_linked")
		return nil, fiber.ErrNotFound
	}

	return loan, nil
}

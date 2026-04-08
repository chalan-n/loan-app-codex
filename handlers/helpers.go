package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"loan-app/repositories"
	"loan-app/services"
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
	loadSessionUser = func(username string) (*models.User, error) {
		var user models.User
		if err := config.DB.Select("id, username, current_session_id, session_last_activity_at, session_revoked_at").
			Where("username = ?", username).
			First(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	persistIssuedSession = func(user *models.User, sessionID string, now time.Time) error {
		user.CurrentSessionID = sessionID
		user.SessionLastActivityAt = &now
		user.SessionRevokedAt = nil
		return config.DB.Model(user).
			Where("id = ?", user.ID).
			Updates(map[string]interface{}{
				"current_session_id":       sessionID,
				"session_last_activity_at": now,
				"session_revoked_at":       nil,
			}).Error
	}
	touchSessionActivity = func(username string, now time.Time) error {
		return config.DB.Model(&models.User{}).
			Where("username = ?", username).
			Update("session_last_activity_at", now).Error
	}
	revokeAllSessionsForUser = func(username string, now time.Time) error {
		return config.DB.Model(&models.User{}).
			Where("username = ?", username).
			Updates(map[string]interface{}{
				"current_session_id":       "",
				"session_last_activity_at": nil,
				"session_revoked_at":       now,
			}).Error
	}
	sessionNow = func() time.Time {
		return time.Now().UTC()
	}
	sessionIdleTimeout = func() time.Duration {
		return time.Duration(config.GetConfig().SessionIdleTimeoutMinutes) * time.Minute
	}
	sessionActivityRefreshInterval = func() time.Duration {
		return time.Duration(config.GetConfig().SessionActivityRefreshSeconds) * time.Second
	}
	loadLoanFileMetadata = func(filename string) (*models.LoanFile, error) {
		if config.DB == nil {
			return nil, errors.New("loan file metadata unavailable")
		}
		return repositories.NewGormLoanFileRepository(config.DB).FindByStorageKey(filename)
	}
	createLoanFileMetadata = func(file *models.LoanFile) error {
		if config.DB == nil {
			return nil
		}
		return repositories.NewGormLoanFileRepository(config.DB).Create(file)
	}
	deleteLoanFileMetadata = func(loanID int, filename string) error {
		if config.DB == nil {
			return nil
		}
		return repositories.NewGormLoanFileRepository(config.DB).DeleteByLoanAndStorageKey(loanID, filename)
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

func parseJWTIssuedAt(tokenStr string) (time.Time, bool) {
	claims, ok := parseJWTClaims(tokenStr)
	if !ok {
		return time.Time{}, false
	}
	return services.ParseIssuedAtClaim(claims)
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

	now := sessionNow()
	if err := persistIssuedSession(user, sessionID, now); err != nil {
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
	if services.CanAccessLoan(role, username, loan) {
		return loan, nil
	}

	logDeniedLoanAccess(c, username, loanID, "forbidden")
	return nil, fiber.ErrForbidden
}

func loanIDFromFilename(filename string) (int, bool) {
	return services.LoanIDFromFilename(filename)
}

func loanHasFile(loan *models.LoanApplication, filename string) bool {
	return services.LoanHasFile(loan, filename)
}

func requireFileAccess(c *fiber.Ctx, filename string) (*models.LoanApplication, error) {
	username := parseJWTUsername(c.Cookies("token"))
	if fileMeta, err := loadLoanFileMetadata(filename); err == nil && fileMeta != nil {
		loan, accessErr := requireLoanAccess(c, fileMeta.LoanID)
		if accessErr != nil {
			if accessErr == fiber.ErrUnauthorized {
				logDeniedFileAccess(c, username, filename, "missing_auth")
			}
			return nil, accessErr
		}
		if fileMeta.StorageKey != filename || fileMeta.LoanID != loan.ID || !loanHasFile(loan, filename) {
			logDeniedFileAccess(c, username, filename, "metadata_mismatch")
			return nil, fiber.ErrNotFound
		}
		return loan, nil
	}

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
		logDeniedFileAccess(c, username, filename, "legacy_file_not_linked")
		return nil, fiber.ErrNotFound
	}

	return loan, nil
}

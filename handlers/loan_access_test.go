package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

func withLoanStubs(t *testing.T, loan *models.LoanApplication, role string) {
	t.Helper()

	prevLoader := loadLoanApplication
	prevRoleLookup := lookupUserRole
	t.Cleanup(func() {
		loadLoanApplication = prevLoader
		lookupUserRole = prevRoleLookup
	})

	loadLoanApplication = func(loanID interface{}) (*models.LoanApplication, error) {
		copyLoan := *loan
		return &copyLoan, nil
	}
	lookupUserRole = func(username string) string {
		return role
	}
}

func newTokenForTests(t *testing.T, username string) string {
	t.Helper()
	token, err := createJWTToken(username, "test-session")
	if err != nil {
		t.Fatalf("createJWTToken() error: %v", err)
	}
	return token
}

func TestRequireLoanAccessRejectsDifferentOfficer(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 7, StaffID: "owner"}, models.RoleOfficer)

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		_, err := requireLoanAccess(c, 7)
		if err != fiber.ErrForbidden {
			t.Fatalf("requireLoanAccess() error = %v, want %v", err, fiber.ErrForbidden)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")})
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

func TestRequireLoanAccessAllowsAdminForOtherUsersLoan(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 7, StaffID: "owner"}, models.RoleAdmin)

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		loan, err := requireLoanAccess(c, 7)
		if err != nil {
			t.Fatalf("requireLoanAccess() unexpected error: %v", err)
		}
		if loan.ID != 7 {
			t.Fatalf("requireLoanAccess() returned loan ID %d, want %d", loan.ID, 7)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "admin-user")})
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

func TestRequireFileAccessRejectsUnknownFilenameForLoan(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "570639", CarInsuranceFile: "12_123_policy.pdf"}, models.RoleOfficer)

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		_, err := requireFileAccess(c, "12_999_other.pdf")
		if err != fiber.ErrNotFound {
			t.Fatalf("requireFileAccess() error = %v, want %v", err, fiber.ErrNotFound)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "570639")})
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
}

func TestGetFileRejectsOtherUsersLoanFile(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner", CarInsuranceFile: "12_123_policy.pdf"}, models.RoleOfficer)

	prevPresign := presignFileURL
	t.Cleanup(func() { presignFileURL = prevPresign })
	presignFileURL = func(filename string) (string, error) {
		t.Fatal("presignFileURL should not be called for forbidden access")
		return "", nil
	}

	app := fiber.New()
	app.Get("/file/:filename", GetFile)

	req := httptest.NewRequest("GET", "/file/12_123_policy.pdf", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("GetFile status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func TestGetFileReturnsRedirectForAccessibleLoanFile(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "570639", CarInsuranceFile: "12_123_policy.pdf"}, models.RoleOfficer)

	prevPresign := presignFileURL
	t.Cleanup(func() { presignFileURL = prevPresign })
	presignFileURL = func(filename string) (string, error) {
		if filename != "12_123_policy.pdf" {
			t.Fatalf("presignFileURL filename = %q, want %q", filename, "12_123_policy.pdf")
		}
		return "https://example.com/presigned", nil
	}

	app := fiber.New()
	app.Get("/file/:filename", GetFile)

	req := httptest.NewRequest("GET", "/file/12_123_policy.pdf", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "570639")})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GetFile status = %d, want %d; body=%s", resp.StatusCode, fiber.StatusFound, string(body))
	}
	if got := resp.Header.Get("Location"); got != "https://example.com/presigned" {
		t.Fatalf("GetFile redirect = %q, want %q", got, "https://example.com/presigned")
	}
}

func TestStep2PostRedirectsWhenLoanBelongsToAnotherOfficer(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner"}, models.RoleOfficer)

	app := fiber.New()
	app.Post("/step2", Step2Post)

	req := httptest.NewRequest("POST", "/step2", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")})
	req.AddCookie(&http.Cookie{Name: "loan_id", Value: "12"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	if resp.StatusCode != fiber.StatusFound {
		t.Fatalf("Step2Post status = %d, want %d", resp.StatusCode, fiber.StatusFound)
	}
	if got := resp.Header.Get("Location"); got != "/step1" {
		t.Fatalf("Step2Post redirect = %q, want %q", got, "/step1")
	}
}

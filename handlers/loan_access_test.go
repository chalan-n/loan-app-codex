package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func doRequestWithCookies(t *testing.T, app *fiber.App, method, target string, body string, cookies ...*http.Cookie) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", fiber.MIMEApplicationForm)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}

	return resp
}

func formBody(values map[string]string) string {
	data := url.Values{}
	for key, value := range values {
		data.Set(key, value)
	}
	return data.Encode()
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

func TestRestrictedStepPostsRedirectWhenLoanBelongsToAnotherOfficer(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner"}, models.RoleOfficer)

	tests := []struct {
		name string
		path string
		h    fiber.Handler
	}{
		{name: "step3", path: "/step3", h: Step3Post},
		{name: "step4", path: "/step4", h: Step4Post},
		{name: "step5", path: "/step5", h: Step5Post},
		{name: "step6", path: "/step6", h: Step6Post},
		{name: "step7", path: "/step7", h: Step7Post},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post(tc.path, tc.h)

			resp := doRequestWithCookies(
				t,
				app,
				http.MethodPost,
				tc.path,
				"",
				&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")},
				&http.Cookie{Name: "loan_id", Value: "12"},
			)
			defer resp.Body.Close()

			if resp.StatusCode != fiber.StatusFound {
				t.Fatalf("%s status = %d, want %d", tc.name, resp.StatusCode, fiber.StatusFound)
			}
			if got := resp.Header.Get("Location"); got != "/step1" {
				t.Fatalf("%s redirect = %q, want %q", tc.name, got, "/step1")
			}
			if setCookie := resp.Header.Get("Set-Cookie"); !strings.Contains(setCookie, "token=") {
				t.Fatalf("%s should clear auth cookie, got %q", tc.name, setCookie)
			}
		})
	}
}

func TestDeleteLoanRejectsDifferentOfficer(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner"}, models.RoleOfficer)

	app := fiber.New()
	app.Post("/delete-loan", DeleteLoan)

	req := httptest.NewRequest(http.MethodPost, "/delete-loan", strings.NewReader(`{"id":12}`))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("DeleteLoan status = %d, want %d; body=%s", resp.StatusCode, fiber.StatusForbidden, string(body))
	}
}

func TestGuarantorRoutesRejectDifferentOfficer(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner"}, models.RoleOfficer)

	t.Run("add-guarantor-get", func(t *testing.T) {
		app := fiber.New()
		app.Get("/guarantor", AddGuarantorGetV2)

		resp := doRequestWithCookies(
			t,
			app,
			http.MethodGet,
			"/guarantor?loan_id=12&guarantor_id=4",
			"",
			&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")},
		)
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusFound {
			t.Fatalf("AddGuarantorGetV2 status = %d, want %d", resp.StatusCode, fiber.StatusFound)
		}
		if got := resp.Header.Get("Location"); got != "/main" {
			t.Fatalf("AddGuarantorGetV2 redirect = %q, want %q", got, "/main")
		}
	})

	t.Run("add-guarantor-post", func(t *testing.T) {
		app := fiber.New()
		app.Post("/guarantor", AddGuarantorPostV2)

		resp := doRequestWithCookies(
			t,
			app,
			http.MethodPost,
			"/guarantor",
			formBody(map[string]string{
				"loan_id":      "12",
				"first_name":   "Jane",
				"last_name":    "Doe",
				"guarantor_id": "4",
			}),
			&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")},
		)
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusForbidden {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("AddGuarantorPostV2 status = %d, want %d; body=%s", resp.StatusCode, fiber.StatusForbidden, string(body))
		}
	})

	t.Run("delete-guarantor", func(t *testing.T) {
		app := fiber.New()
		app.Post("/guarantor/delete", DeleteGuarantor)

		resp := doRequestWithCookies(
			t,
			app,
			http.MethodPost,
			"/guarantor/delete",
			formBody(map[string]string{
				"id":      "4",
				"loan_id": "12",
			}),
			&http.Cookie{Name: "token", Value: newTokenForTests(t, "other-user")},
		)
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusForbidden {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("DeleteGuarantor status = %d, want %d; body=%s", resp.StatusCode, fiber.StatusForbidden, string(body))
		}
	})
}

package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

func TestValidateInsuranceUploadAcceptsPDF(t *testing.T) {
	fileHeader := newMultipartFileHeader(t, "policy.pdf", []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n"))

	src, contentType, err := validateInsuranceUpload(fileHeader, 1024)
	if err != nil {
		t.Fatalf("validateInsuranceUpload() error: %v", err)
	}
	if contentType != "application/pdf" {
		t.Fatalf("contentType = %q, want %q", contentType, "application/pdf")
	}
	if closer, ok := src.(io.Closer); ok {
		defer closer.Close()
	}
}

func TestValidateInsuranceUploadRejectsUnsupportedExtension(t *testing.T) {
	fileHeader := newMultipartFileHeader(t, "policy.exe", []byte("MZ"))

	_, _, err := validateInsuranceUpload(fileHeader, 1024)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}
}

func TestValidateInsuranceUploadRejectsMimeMismatch(t *testing.T) {
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d}
	fileHeader := newMultipartFileHeader(t, "policy.pdf", pngHeader)

	_, _, err := validateInsuranceUpload(fileHeader, 1024)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected content type mismatch error, got %v", err)
	}
}

func TestValidateInsuranceUploadRejectsOversizedFile(t *testing.T) {
	fileHeader := newMultipartFileHeader(t, "policy.pdf", []byte("%PDF-1.4\n"))

	_, _, err := validateInsuranceUpload(fileHeader, 4)
	if err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("expected size validation error, got %v", err)
	}
}

func TestUploadInsuranceFileRejectsInvalidMimeBeforeR2Upload(t *testing.T) {
	withLoanStubs(t, &models.LoanApplication{ID: 12, StaffID: "owner"}, models.RoleOfficer)

	prevPut := putR2Object
	prevCreateMeta := createLoanFileMetadata
	t.Cleanup(func() {
		putR2Object = prevPut
		createLoanFileMetadata = prevCreateMeta
	})

	putR2Object = func(filename string, body io.ReadSeeker, contentType string) error {
		t.Fatal("putR2Object should not be called for invalid uploads")
		return nil
	}
	createLoanFileMetadata = func(file *models.LoanFile) error {
		t.Fatal("createLoanFileMetadata should not be called for invalid uploads")
		return nil
	}

	app := fiber.New()
	app.Post("/api/upload-insurance-file", UploadInsuranceFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "policy.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile() error: %v", err)
	}
	if _, err := part.Write([]byte("plain text that is not a pdf")); err != nil {
		t.Fatalf("part.Write() error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload-insurance-file", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "loan_id", Value: "12"})
	req.AddCookie(&http.Cookie{Name: "token", Value: newTokenForTests(t, "owner")})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if !strings.Contains(string(respBody), "does not match") && !strings.Contains(string(respBody), "unsupported") {
		t.Fatalf("unexpected response body: %s", string(respBody))
	}
}

func newMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("part.Write() error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := req.ParseMultipartForm(int64(len(body.Bytes()) + 1024)); err != nil {
		t.Fatalf("ParseMultipartForm() error: %v", err)
	}

	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file header, got %d", len(files))
	}
	return files[0]
}

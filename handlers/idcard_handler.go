// handlers/idcard_handler.go
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"loan-app/services"

	"github.com/gofiber/fiber/v2"
)

// reIDCard ตรวจว่าเลขบัตรประชาชนไทยถูกต้อง: 13 หลัก และผ่าน checksum
var reIDCard13 = regexp.MustCompile(`^\d{13}$`)

// OcrIDCard รับรูปบัตรประชาชนแบบ multipart/form-data (field: "image")
// POST /api/v1/ocr/idcard
func OcrIDCard(c *fiber.Ctx) error {
	moUsername := extractUsername(c)
	branch := strings.TrimSpace(c.FormValue("branch"))
	if branch == "" {
		branch = "ไม่ระบุ"
	}

	// ── 1. รับไฟล์ ────────────────────────────────────────────────────
	fileHeader, err := c.FormFile("image")
	if err != nil {
		logOCRRequest(moUsername, branch, "", 0, "IDCARD ERROR: ไม่พบฟิลด์ 'image'")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ต้องแนบรูปภาพในฟิลด์ 'image'",
		})
	}

	filename := fileHeader.Filename
	fileSize := fileHeader.Size

	// ── 2. ตรวจ Extension ────────────────────────────────────────────
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("IDCARD REJECTED: นามสกุล '%s' ไม่อนุญาต", ext))
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("นามสกุลไฟล์ '%s' ไม่ได้รับอนุญาต", ext),
		})
	}

	// ── 3. ตรวจ Size ─────────────────────────────────────────────────
	if fileSize > maxFileSize {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("IDCARD REJECTED: ขนาด %.2f MB เกิน %d MB", float64(fileSize)/1024/1024, maxFileSizeMB))
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("ขนาดไฟล์ %.2f MB เกินกำหนด (สูงสุด %d MB)", float64(fileSize)/1024/1024, maxFileSizeMB),
		})
	}

	// ── 4. อ่าน bytes ─────────────────────────────────────────────────
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": "เปิดไฟล์ไม่ได้",
		})
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": "อ่านไฟล์ไม่ได้",
		})
	}

	// ── 5. Detect MIME ────────────────────────────────────────────────
	mimeType := http.DetectContentType(imageBytes)
	if !isAllowedMIME(mimeType) {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("IDCARD REJECTED: MIME '%s' ไม่อนุญาต", mimeType))
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"success": false,
			"message": "เนื้อไฟล์ไม่ใช่รูปภาพจริง — MIME: " + mimeType,
		})
	}

	// ── 6. เรียก Gemini ───────────────────────────────────────────────
	logOCRRequest(moUsername, branch, filename, fileSize, "IDCARD SENDING → Gemini")

	idInfo, err := services.AnalyzeIDCard(c.Context(), imageBytes, mimeType)
	if err != nil {
		logOCRRequest(moUsername, branch, filename, fileSize, "IDCARD FAILED: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "วิเคราะห์บัตรประชาชนไม่สำเร็จ",
			"error":   err.Error(),
		})
	}

	// ── 7. Validate เลขบัตร 13 หลัก + checksum ───────────────────────
	idValid := false
	if reIDCard13.MatchString(idInfo.IDNumber) {
		idValid = validateThaiIDChecksum(idInfo.IDNumber)
	}

	logOCRRequest(moUsername, branch, filename, fileSize,
		fmt.Sprintf("IDCARD SUCCESS: เลข=%s valid=%v ชื่อ=%s %s",
			idInfo.IDNumber, idValid, idInfo.FirstName, idInfo.LastName))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":  true,
		"message":  "วิเคราะห์บัตรประชาชนสำเร็จ",
		"id_valid": idValid,
		"data":     idInfo,
	})
}

// validateThaiIDChecksum ตรวจ checksum เลขบัตรประชาชน 13 หลักตามอัลกอริทึมกรมการปกครอง
func validateThaiIDChecksum(id string) bool {
	if len(id) != 13 {
		return false
	}
	sum := 0
	for i := 0; i < 12; i++ {
		d := int(id[i] - '0')
		sum += d * (13 - i)
	}
	check := (11 - (sum % 11)) % 10
	last := int(id[12] - '0')
	return check == last
}

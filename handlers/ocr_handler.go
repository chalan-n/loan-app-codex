// handlers/ocr_handler.go
package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"loan-app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// maxFileSizeMB คือขนาดสูงสุดของไฟล์ที่รับได้ (MB)
const maxFileSizeMB = 5
const maxFileSize = maxFileSizeMB * 1024 * 1024 // 5 MB → bytes

// allowedExtensions นามสกุลไฟล์ที่อนุญาต
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// OcrVehicleBook รับรูปภาพเล่มทะเบียนรถไทยแบบ multipart/form-data
//
//   - field "image"  : ไฟล์รูปภาพ (required)
//   - field "branch" : รหัสสาขา เช่น "CMI" (optional — ถ้าไม่ส่งจะใช้ "ไม่ระบุ")
//
// POST /api/v1/ocr/vehicle
func OcrVehicleBook(c *fiber.Ctx) error {
	// ── 0. ดึงข้อมูล MO จาก JWT (Bearer Token หรือ Cookie) ───────────
	moUsername := extractUsername(c)
	branch := strings.TrimSpace(c.FormValue("branch"))
	if branch == "" {
		branch = "ไม่ระบุ"
	}

	// ── 1. รับไฟล์จาก multipart form ─────────────────────────────────
	fileHeader, err := c.FormFile("image")
	if err != nil {
		logOCRRequest(moUsername, branch, "", 0, "ERROR: ไม่พบฟิลด์ 'image'")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ต้องแนบรูปภาพในฟิลด์ 'image' แบบ multipart/form-data",
		})
	}

	filename := fileHeader.Filename
	fileSize := fileHeader.Size

	// ── 2. ตรวจ File Extension ────────────────────────────────────────
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("REJECTED: นามสกุลไฟล์ '%s' ไม่ได้รับอนุญาต", ext))
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("นามสกุลไฟล์ '%s' ไม่ได้รับอนุญาต — ใช้ได้เฉพาะ .jpg, .jpeg, .png, .webp", ext),
		})
	}

	// ── 3. ตรวจ File Size ─────────────────────────────────────────────
	if fileSize > maxFileSize {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("REJECTED: ขนาดไฟล์ %.2f MB เกิน %d MB", float64(fileSize)/1024/1024, maxFileSizeMB))
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("ขนาดไฟล์ %.2f MB เกินกำหนด (สูงสุด %d MB)", float64(fileSize)/1024/1024, maxFileSizeMB),
		})
	}

	// ── 4. เปิดไฟล์และอ่าน bytes ──────────────────────────────────────
	file, err := fileHeader.Open()
	if err != nil {
		logOCRRequest(moUsername, branch, filename, fileSize, "ERROR: เปิดไฟล์ไม่ได้")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "เปิดไฟล์ไม่ได้",
		})
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		logOCRRequest(moUsername, branch, filename, fileSize, "ERROR: อ่านไฟล์ไม่ได้")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "อ่านไฟล์ไม่ได้",
		})
	}

	// ── 5. Detect MIME type จาก magic bytes (ปลอดภัยกว่าเชื่อ client) ─
	mimeType := http.DetectContentType(imageBytes)
	if !isAllowedMIME(mimeType) {
		logOCRRequest(moUsername, branch, filename, fileSize,
			fmt.Sprintf("REJECTED: MIME type '%s' ไม่ได้รับอนุญาต (อาจเป็นไฟล์ปลอม)", mimeType))
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"success": false,
			"message": "เนื้อไฟล์ไม่ใช่รูปภาพจริง — ตรวจพบ MIME: " + mimeType,
		})
	}

	// ── 6. Log ก่อนส่ง Gemini ─────────────────────────────────────────
	logOCRRequest(moUsername, branch, filename, fileSize, "SENDING → Gemini")

	// ── 7. เรียก Gemini OCR Service ───────────────────────────────────
	vehicleInfo, err := services.AnalyzeVehicleBook(c.Context(), imageBytes, mimeType)
	if err != nil {
		logOCRRequest(moUsername, branch, filename, fileSize, "FAILED: "+err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "วิเคราะห์รูปภาพไม่สำเร็จ",
			"error":   err.Error(),
		})
	}

	// ── 8. Log สำเร็จ ─────────────────────────────────────────────────
	logOCRRequest(moUsername, branch, filename, fileSize,
		fmt.Sprintf("SUCCESS: เลขตัวรถ=%s ทะเบียน=%s", vehicleInfo.ChassisNumber, vehicleInfo.PlateNumber))

	// ── 9. คืนผลลัพธ์ ─────────────────────────────────────────────────
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "วิเคราะห์เล่มทะเบียนสำเร็จ",
		"data":    vehicleInfo,
	})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// logOCRRequest เขียน log แบบ structured ระบุว่า MO/สาขาไหน ส่งไฟล์อะไร ขนาดเท่าไร
func logOCRRequest(mo, branch, filename string, sizeBytes int64, status string) {
	sizeMB := float64(sizeBytes) / 1024 / 1024
	log.Printf("[OCR] %s | MO=%-10s | สาขา=%-6s | ไฟล์=%-30s | %.2f MB | %s",
		time.Now().Format("2006-01-02 15:04:05"),
		mo, branch, filename, sizeMB, status,
	)
}

// extractUsername ดึง username ออกจาก JWT
// รองรับทั้ง Authorization: Bearer <token> (Flutter) และ Cookie "token" (web)
func extractUsername(c *fiber.Ctx) string {
	// 1. ลอง Authorization header ก่อน (Flutter ส่งมาแบบนี้)
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if u := parseJWTUsername(tokenStr); u != "" {
			return u
		}
	}

	// 2. Fallback: Cookie (web browser)
	tokenStr := c.Cookies("token")
	if u := parseJWTUsername(tokenStr); u != "" {
		return u
	}

	return "anonymous"
}

// parseJWTUsername parse token string ดึง claim "username"
func parseJWTUsername(tokenStr string) string {
	if tokenStr == "" {
		return ""
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("mysecret"), nil
	})
	if err != nil || !token.Valid {
		return ""
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		u, _ := claims["username"].(string)
		return u
	}
	return ""
}

// isAllowedMIME ตรวจ MIME type จาก magic bytes
func isAllowedMIME(mime string) bool {
	return map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}[mime]
}

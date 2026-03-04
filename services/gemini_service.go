// services/gemini_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/genai"

	"loan-app/config"
	"loan-app/models"
)

// modelName กำหนดโมเดลที่ใช้งาน — flash เพื่อความเร็วสูงสุด
const modelName = "gemini-2.5-flash-lite"

// ocrPrompt คือ prompt ที่บอก Gemini ให้สกัดข้อมูลจากเล่มทะเบียนรถไทยและตอบ JSON เท่านั้น
const ocrPrompt = `จงสกัดข้อมูลจากรูปภาพเล่มทะเบียนรถไทยนี้ให้ออกมาเป็น JSON ตามโครงสร้างที่กำหนด:
{
  "registration_date": "วันจดทะเบียน (string)",
  "plate_number": "เลขทะเบียน (string)",
  "province": "จังหวัด (string)",
  "vehicle_brand": "ยี่ห้อรถ (string)",
  "chassis_number": "เลขตัวรถ / Chassis Number (string)",
  "engine_number": "เลขเครื่องยนต์ (string)",
  "model_year": ปี ค.ศ. ของรุ่น (int — แปลง พ.ศ. เป็น ค.ศ. โดยลบ 543 ให้ถูกต้อง),
  "color": "สีรถ (string)",
  "engine_cc": ขนาดเครื่องยนต์ (int — หน่วยเป็น cc เช่น 1500 ไม่ใส่หน่วย),
  "car_weight": น้ำหนักรถ (int — หน่วยเป็น กิโลกรัม เช่น 1200 ไม่ใส่หน่วย)
}
ข้อกำหนด:
- ตอบกลับเฉพาะ JSON เท่านั้น ห้ามมีข้อความอื่น ห้ามใช้ markdown code block
- model_year ต้องเป็นตัวเลข ค.ศ. เท่านั้น (เช่น 2568 พ.ศ. = 2025 ค.ศ.)
- engine_cc และ car_weight ต้องเป็นตัวเลขเท่านั้น ไม่ใส่คำว่า "cc" หรือ "กก."
- ถ้าหาข้อมูลใดไม่พบให้ใส่ "" หรือ 0 ตามประเภทข้อมูล`

// AnalyzeVehicleBook รับ imageData (raw bytes ของรูปภาพเล่มทะเบียน) และ mimeType เช่น "image/jpeg"
// แล้วส่งไปให้ Gemini วิเคราะห์ คืนค่าเป็น *models.VehicleInfo ที่ผ่านการ Clean แล้ว
func AnalyzeVehicleBook(ctx context.Context, imageData []byte, mimeType string) (*models.VehicleInfo, error) {
	// ── 1. ตรวจสอบ Input ──────────────────────────────────────────────
	if len(imageData) == 0 {
		return nil, fmt.Errorf("gemini_service: imageData ต้องไม่ว่างเปล่า")
	}

	// ถ้าไม่ระบุ mimeType ให้ detect จาก magic bytes อัตโนมัติ
	if mimeType == "" {
		mimeType = detectMIMEType(imageData)
	}

	// ── 2. สร้าง Gemini Client ────────────────────────────────────────
	cfg := config.GetConfig()
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("gemini_service: GEMINI_API_KEY ยังไม่ได้ตั้งค่าใน .env")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini_service: สร้าง Gemini client ไม่ได้: %w", err)
	}

	// ── 3. สร้าง Request Parts (Text Prompt + Inline Image) ────────────
	parts := []*genai.Part{
		// prompt อธิบาย task
		{Text: ocrPrompt},
		// แนบรูปภาพแบบ inline (เหมาะกับไฟล์ < 20MB)
		{InlineData: &genai.Blob{
			Data:     imageData,
			MIMEType: mimeType,
		}},
	}

	contents := []*genai.Content{
		{Parts: parts, Role: "user"},
	}

	// ── 4. เรียก Gemini API ───────────────────────────────────────────
	result, err := client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini_service: เรียก Gemini API ล้มเหลว: %w", err)
	}

	// ── 5. ดึง Text จาก Response ─────────────────────────────────────
	rawText, err := extractText(result)
	if err != nil {
		return nil, err
	}

	// ── 6. Parse JSON → VehicleInfo ───────────────────────────────────
	info, err := parseVehicleJSON(rawText)
	if err != nil {
		return nil, err
	}

	// ── 7. Clean ข้อมูล (ตัด noise, whitespace, แปลง พ.ศ.) ──────────
	return info.Clean(), nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// extractText ดึง text ออกจาก GenerateContentResponse
// คืน error ถ้า response ว่างเปล่าหรือไม่มี candidate
func extractText(result *genai.GenerateContentResponse) (string, error) {
	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("gemini_service: Gemini ไม่คืน candidate ใดๆ")
	}

	candidate := result.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("gemini_service: Candidate ว่างเปล่า (FinishReason: %v)", candidate.FinishReason)
	}

	// รวม text ทุก part เข้าด้วยกัน
	var sb strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			sb.WriteString(part.Text)
		}
	}

	text := strings.TrimSpace(sb.String())
	if text == "" {
		return "", fmt.Errorf("gemini_service: Gemini คืน response ว่างเปล่า")
	}
	return text, nil
}

// parseVehicleJSON แปลง raw text (อาจมี markdown fence) เป็น *models.VehicleInfo
func parseVehicleJSON(raw string) (*models.VehicleInfo, error) {
	// Gemini บางครั้งอาจตอบกลับมาพร้อม ```json ... ``` ทั้งที่บอกแล้วว่าห้าม
	// ป้องกัน edge case นี้ด้วยการ strip fence ออกก่อน
	jsonStr := stripMarkdownFence(raw)

	var info models.VehicleInfo
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		// log raw เพื่อ debug แต่ truncate ไม่เกิน 200 chars
		preview := jsonStr
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("gemini_service: parse JSON ล้มเหลว: %w\nraw response: %s", err, preview)
	}
	return &info, nil
}

// stripMarkdownFence ลบ ```json ... ``` หรือ ``` ... ``` ออกจาก string
func stripMarkdownFence(s string) string {
	// ลอง trim หัวท้ายก่อน
	s = strings.TrimSpace(s)

	// กรณี ```json\n...\n``` หรือ ```\n...\n```
	for _, fence := range []string{"```json", "```"} {
		if strings.HasPrefix(s, fence) {
			s = strings.TrimPrefix(s, fence)
			// ตัด closing fence
			if idx := strings.LastIndex(s, "```"); idx >= 0 {
				s = s[:idx]
			}
			s = strings.TrimSpace(s)
			break
		}
	}
	return s
}

// detectMIMEType ใช้ magic bytes เพื่อ detect ประเภทไฟล์รูปภาพ
// fallback เป็น "image/jpeg" ซึ่งพบบ่อยที่สุดสำหรับรูปถ่าย
func detectMIMEType(data []byte) string {
	mime := http.DetectContentType(data)
	// http.DetectContentType คืน "image/jpeg", "image/png", "image/webp" ฯลฯ
	// ถ้าไม่รู้จักให้ default เป็น jpeg
	switch mime {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return mime
	default:
		return "image/jpeg"
	}
}

// ─── ID Card OCR ──────────────────────────────────────────────────────────────

// idCardPrompt สั่ง Gemini สกัดข้อมูลจากบัตรประชาชนไทย
const idCardPrompt = `จงสกัดข้อมูลจากรูปภาพบัตรประชาชนไทยนี้ให้ออกมาเป็น JSON ตามโครงสร้างที่กำหนด:
{
  "id_number": "เลขบัตรประชาชน 13 หลัก (string — เฉพาะตัวเลข ไม่มีขีด)",
  "title": "คำนำหน้า เช่น นาย นาง นางสาว (string)",
  "first_name": "ชื่อภาษาไทย (string)",
  "last_name": "นามสกุลภาษาไทย (string)",
  "date_of_birth": "วันเกิด รูปแบบ DD/MM/YYYY โดยปีเป็น ค.ศ. (string — แปลง พ.ศ. โดยลบ 543)",
  "gender": "เพศ: ชาย หรือ หญิง (string)",
  "house_no": "เลขที่บ้าน (string)",
  "moo": "หมู่ที่ เฉพาะตัวเลข (string)",
  "soi": "ซอย (string)",
  "road": "ถนน (string)",
  "subdistrict": "ตำบล/แขวง ไม่ต้องมีคำว่า ตำบล หรือ แขวง นำหน้า (string)",
  "district": "อำเภอ/เขต ไม่ต้องมีคำว่า อำเภอ หรือ เขต นำหน้า (string)",
  "province": "จังหวัด ไม่ต้องมีคำว่า จังหวัด นำหน้า (string)",
  "zipcode": "รหัสไปรษณีย์ 5 หลัก (string)",
  "issue_date": "วันออกบัตร รูปแบบ DD/MM/YYYY ค.ศ. (string)",
  "expiry_date": "วันหมดอายุ รูปแบบ DD/MM/YYYY ค.ศ. (string)",
  "religion": "ศาสนา เช่น พุทธ อิสลาม คริสต์ (string)",
  "address": "ที่อยู่แบบ raw ทั้งหมดรวมกัน (string — ใส่ถ้าอ่านแบบแยกส่วนไม่ได้)"
}
ข้อกำหนด:
- ตอบกลับเฉพาะ JSON เท่านั้น ห้ามมีข้อความอื่น ห้ามใช้ markdown code block
- id_number ต้องเป็นตัวเลข 13 หลักเท่านั้น ลบขีด (-) ออก
- วันที่ทุกช่องต้องเป็น ค.ศ. (แปลง พ.ศ. โดยลบ 543)
- ถ้าหาข้อมูลใดไม่พบให้ใส่ "" (string ว่าง)`

// AnalyzeIDCard รับ imageData (raw bytes ของรูปบัตรประชาชน) และ mimeType
// ส่งไปให้ Gemini วิเคราะห์ คืนค่าเป็น *models.IDCardInfo ที่ผ่านการ Clean แล้ว
func AnalyzeIDCard(ctx context.Context, imageData []byte, mimeType string) (*models.IDCardInfo, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("gemini_service: imageData ต้องไม่ว่างเปล่า")
	}
	if mimeType == "" {
		mimeType = detectMIMEType(imageData)
	}

	cfg := config.GetConfig()
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("gemini_service: GEMINI_API_KEY ยังไม่ได้ตั้งค่าใน .env")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini_service: สร้าง Gemini client ไม่ได้: %w", err)
	}

	parts := []*genai.Part{
		{Text: idCardPrompt},
		{InlineData: &genai.Blob{Data: imageData, MIMEType: mimeType}},
	}
	contents := []*genai.Content{{Parts: parts, Role: "user"}}

	result, err := client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini_service: เรียก Gemini API ล้มเหลว: %w", err)
	}

	rawText, err := extractText(result)
	if err != nil {
		return nil, err
	}

	jsonStr := stripMarkdownFence(rawText)
	var info models.IDCardInfo
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		preview := jsonStr
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("gemini_service: parse JSON ล้มเหลว: %w\nraw: %s", err, preview)
	}

	return info.Clean(), nil
}

// models/vehicle.go
package models

import (
	"regexp"
	"strings"
	"unicode"
)

// VehicleInfo เก็บข้อมูลที่อ่านได้จาก OCR เล่มทะเบียนรถไทย
type VehicleInfo struct {
	// วันจดทะเบียน เช่น "15 มกราคม 2568"
	RegistrationDate string `json:"registration_date"`

	// เลขทะเบียน เช่น "กข 1234"
	PlateNumber string `json:"plate_number"`

	// จังหวัด เช่น "กรุงเทพมหานคร"
	Province string `json:"province"`

	// ยี่ห้อรถ เช่น "TOYOTA"
	VehicleBrand string `json:"vehicle_brand"`

	// เลขตัวรถ (Chassis / VIN) — สำคัญมาก ใช้ยืนยันตัวรถ
	ChassisNumber string `json:"chassis_number"`

	// เลขเครื่องยนต์
	EngineNumber string `json:"engine_number"`

	// ปี ค.ศ. ของรุ่น (แปลงจาก พ.ศ. แล้ว เช่น 2568 → 2025)
	ModelYear int `json:"model_year"`

	// สีรถ เช่น "ขาว"
	Color string `json:"color"`

	// ขนาดเครื่องยนต์ (cc) เช่น 1500
	EngineCC int `json:"engine_cc"`

	// น้ำหนักรถ (กก.) เช่น 1200
	CarWeight int `json:"car_weight"`
}

// -------------------------------------------------------------------
// Clean — ล้างข้อมูลทุกฟิลด์ที่ได้จาก OCR / Gemini
// ควรเรียก v.Clean() ทันทีหลัง json.Unmarshal
// -------------------------------------------------------------------

// Clean ล้างข้อมูลทุกฟิลด์ใน VehicleInfo แล้วคืน *VehicleInfo ตัวเดิม
// เพื่อให้ chain ได้ เช่น info.Clean().ToMap()
func (v *VehicleInfo) Clean() *VehicleInfo {
	v.RegistrationDate = cleanField(v.RegistrationDate)
	v.PlateNumber = cleanField(v.PlateNumber)
	v.Province = cleanField(v.Province)
	v.VehicleBrand = cleanField(v.VehicleBrand)

	// เลขตัวรถ / เลขเครื่อง — ไม่ควรมีช่องว่างกลาง ลบ whitespace ออกทั้งหมด
	v.ChassisNumber = cleanCode(v.ChassisNumber)
	v.EngineNumber = cleanCode(v.EngineNumber)

	v.Color = cleanField(v.Color)

	// แปลง พ.ศ. → ค.ศ. อัตโนมัติ ถ้า ModelYear >= 2400
	if v.ModelYear >= 2400 {
		v.ModelYear -= 543
	}

	// EngineCC / CarWeight: ค่าลบไม่สมเหตุสมผล reset เป็น 0
	if v.EngineCC < 0 {
		v.EngineCC = 0
	}
	if v.CarWeight < 0 {
		v.CarWeight = 0
	}

	return v
}

// -------------------------------------------------------------------
// helpers (unexported)
// -------------------------------------------------------------------

// reOCRNoise จับอักขระที่ Gemini / OCR มักอ่านเกินมา
// เช่น "|", "–", "—", "*", "#", "°", backtick, control chars
var reOCRNoise = regexp.MustCompile(`[|–—*#°` + "`" + `\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

// cleanField ล้างค่าทั่วไป (ชื่อ, จังหวัด, สี, วันที่ ฯลฯ)
//  1. ตัด leading/trailing whitespace (รวม \t, \r, \n, non-breaking space U+00A0)
//  2. กำจัดอักขระ noise ที่ OCR อ่านเกิน
//  3. ยุบ whitespace กลางประโยคให้เหลือ space เดียว
func cleanField(s string) string {
	// trim unicode spaces ทุกชนิด (รวม \u00A0, \u200B ฯลฯ)
	s = strings.TrimFunc(s, unicode.IsSpace)
	// ลบ noise characters
	s = reOCRNoise.ReplaceAllString(s, "")
	// ยุบช่องว่างซ้อนกัน
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// cleanCode ใช้กับเลขตัวรถ / เลขเครื่อง
// เหมือน cleanField แต่ลบ whitespace กลางออกด้วย และ uppercase
// (VIN / Chassis Number ไม่ควรมีช่องว่างในตัวเลขเลย)
func cleanCode(s string) string {
	s = cleanField(s)
	// ลบ whitespace กลางออกทั้งหมด
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
	// uppercase เพื่อให้ format สม่ำเสมอ
	return strings.ToUpper(s)
}

// -------------------------------------------------------------------
// ToMap — แปลง VehicleInfo เป็น map[string]interface{}
// เพื่อนำไป Update ลง MySQL ได้สะดวก เช่น db.Model(&loan).Updates(info.ToMap())
// -------------------------------------------------------------------
func (v *VehicleInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"registration_date": v.RegistrationDate,
		"plate_number":      v.PlateNumber,
		"province":          v.Province,
		"vehicle_brand":     v.VehicleBrand,
		"chassis_number":    v.ChassisNumber,
		"engine_number":     v.EngineNumber,
		"model_year":        v.ModelYear,
		"color":             v.Color,
		"engine_cc":         v.EngineCC,
		"car_weight":        v.CarWeight,
	}
}

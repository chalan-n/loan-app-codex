// models/idcard.go
package models

import "strings"

// IDCardInfo เก็บข้อมูลที่อ่านได้จาก OCR บัตรประชาชนไทย
type IDCardInfo struct {
	// เลขบัตรประชาชน 13 หลัก
	IDNumber string `json:"id_number"`

	// คำนำหน้า เช่น "นาย", "นาง", "นางสาว"
	Title string `json:"title"`

	// ชื่อ (ภาษาไทย)
	FirstName string `json:"first_name"`

	// นามสกุล (ภาษาไทย)
	LastName string `json:"last_name"`

	// วันเกิด รูปแบบ DD/MM/YYYY (ค.ศ.) เช่น "15/01/1990"
	// Gemini จะแปลง พ.ศ. → ค.ศ. ให้อัตโนมัติตาม prompt
	DateOfBirth string `json:"date_of_birth"`

	// เพศ: "ชาย" หรือ "หญิง"
	Gender string `json:"gender"`

	// ที่อยู่จากบัตร — แยกเป็นส่วนๆ
	Address     string `json:"address"`     // ที่อยู่รวมแบบ raw (fallback)
	HouseNo     string `json:"house_no"`    // เลขที่บ้าน
	Moo         string `json:"moo"`         // หมู่
	Soi         string `json:"soi"`         // ซอย
	Road        string `json:"road"`        // ถนน
	Subdistrict string `json:"subdistrict"` // ตำบล/แขวง
	District    string `json:"district"`    // อำเภอ/เขต
	Province    string `json:"province"`    // จังหวัด
	Zipcode     string `json:"zipcode"`     // รหัสไปรษณีย์

	// วันออกบัตร / วันหมดอายุ (DD/MM/YYYY ค.ศ.)
	IssueDate  string `json:"issue_date"`
	ExpiryDate string `json:"expiry_date"`

	// ศาสนา เช่น "พุทธ" "อิสลาม" "คริสต์"
	Religion string `json:"religion"`
}

// Clean ล้างข้อมูลทุกฟิลด์ใน IDCardInfo แล้วคืน *IDCardInfo ตัวเดิม
func (c *IDCardInfo) Clean() *IDCardInfo {
	c.IDNumber = cleanCode(strings.ReplaceAll(c.IDNumber, "-", ""))
	c.Title = cleanField(c.Title)
	c.FirstName = cleanField(c.FirstName)
	c.LastName = cleanField(c.LastName)
	c.DateOfBirth = cleanField(c.DateOfBirth)
	c.Gender = cleanField(c.Gender)
	c.Address = cleanField(c.Address)
	c.HouseNo = cleanField(c.HouseNo)
	c.Moo = cleanField(c.Moo)
	c.Soi = cleanField(c.Soi)
	c.Road = cleanField(c.Road)
	c.Subdistrict = cleanField(c.Subdistrict)
	c.District = cleanField(c.District)
	c.Province = cleanField(c.Province)
	c.Zipcode = cleanCode(c.Zipcode)
	c.IssueDate = cleanField(c.IssueDate)
	c.ExpiryDate = cleanField(c.ExpiryDate)
	c.Religion = cleanField(c.Religion)
	return c
}

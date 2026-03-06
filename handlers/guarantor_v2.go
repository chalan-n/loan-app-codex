package handlers

import (
	"log"
	"strconv"
	"strings"

	"loan-app/config"
	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

// parseMoney แปลง string ที่มี comma เป็น float64
func parseMoney(amount string) float64 {
	amount = strings.ReplaceAll(amount, ",", "")
	val, _ := strconv.ParseFloat(amount, 64)
	return val
}

// Helper to safely dereference string pointers
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func AddGuarantorGetV2(c *fiber.Ctx) error {
	loanID := c.Query("loan_id")
	guarantorID := c.Query("guarantor_id")

	if loanID == "" {
		return c.Redirect("/")
	}

	// Create a map to hold safe values for the template
	guarantorData := make(map[string]interface{})
	var guarantor models.Guarantor

	if guarantorID != "" {
		if err := config.DB.First(&guarantor, guarantorID).Error; err == nil {
			// Manually map fields to safe values
			// General
			guarantorData["ID"] = guarantor.ID
			guarantorData["Title"] = guarantor.Title
			guarantorData["FirstName"] = guarantor.FirstName
			guarantorData["LastName"] = guarantor.LastName
			guarantorData["Gender"] = guarantor.Gender
			guarantorData["IdCard"] = guarantor.IdCard

			// Safe Dates
			guarantorData["IdCardIssueDate"] = safeStr(guarantor.IdCardIssueDate)
			guarantorData["IdCardExpiryDate"] = safeStr(guarantor.IdCardExpiryDate)
			guarantorData["DateOfBirth"] = safeStr(guarantor.DateOfBirth)

			guarantorData["Ethnicity"] = guarantor.Ethnicity
			guarantorData["Nationality"] = guarantor.Nationality
			guarantorData["Religion"] = guarantor.Religion
			guarantorData["MaritalStatus"] = guarantor.MaritalStatus
			guarantorData["MobilePhone"] = guarantor.MobilePhone

			// House Address
			guarantorData["HouseRegNo"] = guarantor.HouseRegNo
			guarantorData["HouseRegMoo"] = guarantor.HouseRegMoo
			guarantorData["HouseRegSoi"] = guarantor.HouseRegSoi
			guarantorData["HouseRegRoad"] = guarantor.HouseRegRoad
			guarantorData["HouseRegProvince"] = guarantor.HouseRegProvince
			guarantorData["HouseRegDistrict"] = guarantor.HouseRegDistrict
			guarantorData["HouseRegSubdistrict"] = guarantor.HouseRegSubdistrict
			guarantorData["HouseRegZipcode"] = guarantor.HouseRegZipcode

			guarantorData["SameAsHouseReg"] = guarantor.SameAsHouseReg

			// Current Address
			guarantorData["CurrentNo"] = guarantor.CurrentNo
			guarantorData["CurrentMoo"] = guarantor.CurrentMoo
			guarantorData["CurrentSoi"] = guarantor.CurrentSoi
			guarantorData["CurrentRoad"] = guarantor.CurrentRoad
			guarantorData["CurrentProvince"] = guarantor.CurrentProvince
			guarantorData["CurrentDistrict"] = guarantor.CurrentDistrict
			guarantorData["CurrentSubdistrict"] = guarantor.CurrentSubdistrict
			guarantorData["CurrentZipcode"] = guarantor.CurrentZipcode

			// Work
			guarantorData["CompanyName"] = guarantor.CompanyName
			guarantorData["Occupation"] = guarantor.Occupation
			guarantorData["Position"] = guarantor.Position
			guarantorData["Salary"] = guarantor.Salary
			guarantorData["OtherIncome"] = guarantor.OtherIncome
			guarantorData["IncomeSource"] = guarantor.IncomeSource

			guarantorData["WorkNo"] = guarantor.WorkNo
			guarantorData["WorkMoo"] = guarantor.WorkMoo
			guarantorData["WorkSoi"] = guarantor.WorkSoi
			guarantorData["WorkRoad"] = guarantor.WorkRoad
			guarantorData["WorkProvince"] = guarantor.WorkProvince
			guarantorData["WorkDistrict"] = guarantor.WorkDistrict
			guarantorData["WorkSubdistrict"] = guarantor.WorkSubdistrict
			guarantorData["WorkZipcode"] = guarantor.WorkZipcode
			guarantorData["WorkPhone"] = guarantor.WorkPhone

			// Other Card
			guarantorData["OtherCardType"] = guarantor.OtherCardType
			guarantorData["OtherCardNumber"] = guarantor.OtherCardNumber
			guarantorData["OtherCardIssueDate"] = safeStr(guarantor.OtherCardIssueDate)
			guarantorData["OtherCardExpiryDate"] = safeStr(guarantor.OtherCardExpiryDate)

			// Doc Delivery
			guarantorData["DocDeliveryType"] = guarantor.DocDeliveryType
			guarantorData["DocNo"] = guarantor.DocNo
			guarantorData["DocMoo"] = guarantor.DocMoo
			guarantorData["DocSoi"] = guarantor.DocSoi
			guarantorData["DocRoad"] = guarantor.DocRoad
			guarantorData["DocProvince"] = guarantor.DocProvince
			guarantorData["DocDistrict"] = guarantor.DocDistrict
			guarantorData["DocSubdistrict"] = guarantor.DocSubdistrict
			guarantorData["DocZipcode"] = guarantor.DocZipcode
		}
	}

	return c.Render("add_guarantor", fiber.Map{
		"LoanID":    loanID,
		"Guarantor": guarantorData, // Pass the map instead of the struct
		"IsEdit":    guarantorID != "",
	})
}

// AddGuarantorPostV2 - Renamed to force code refresh
func AddGuarantorPostV2(c *fiber.Ctx) error {
	// Parse loan_id directly
	loanIDStr := c.FormValue("loan_id")
	guarantorID := c.FormValue("guarantor_id")

	if guarantorID != "" {
		// UPDATE existing guarantor
		// Logic to pick the correct company name
		var companyName string
		if c.FormValue("guarantor_type") == "juristic" {
			companyName = c.FormValue("juristic_company_name")
		} else {
			companyName = c.FormValue("work_company_name")
		}

		query := `UPDATE loan_applications_guarantors SET
            updated_at = NOW(),
            guarantor_type = ?,
            trade_registration_id = ?, registration_date = NULLIF(?, ''), tax_id = ?,
            title = ?, first_name = ?, last_name = ?, gender = ?, id_card = ?,
            id_card_issue_date = NULLIF(?, ''), id_card_expiry_date = NULLIF(?, ''), date_of_birth = NULLIF(?, ''),
            ethnicity = ?, nationality = ?, religion = ?, marital_status = ?, mobile_phone = ?,
            house_reg_no = ?, house_reg_moo = ?, house_reg_soi = ?, house_reg_road = ?, house_reg_province = ?, house_reg_district = ?, house_reg_subdistrict = ?, house_reg_zipcode = ?,
            same_as_house_reg = ?,
            current_no = ?, current_moo = ?, current_soi = ?, current_road = ?, current_province = ?, current_district = ?, current_subdistrict = ?, current_zipcode = ?,
            company_name = ?, occupation = ?, position = ?, salary = ?, other_income = ?, income_source = ?,
            work_phone = ?, work_no = ?, work_moo = ?, work_soi = ?, work_road = ?, work_province = ?, work_district = ?, work_subdistrict = ?, work_zipcode = ?,
            other_card_type = ?, other_card_number = ?, other_card_issue_date = NULLIF(?, ''), other_card_expiry_date = NULLIF(?, ''),
            doc_delivery_type = ?, doc_no = ?, doc_moo = ?, doc_soi = ?, doc_road = ?, doc_province = ?, doc_district = ?, doc_subdistrict = ?, doc_zipcode = ?
            WHERE id = ?`

		err := config.DB.Exec(query,
			c.FormValue("guarantor_type"),
			c.FormValue("trade_registration_id"), c.FormValue("registration_date"), c.FormValue("tax_id"),
			c.FormValue("title"), c.FormValue("first_name"), c.FormValue("last_name"), c.FormValue("gender"), c.FormValue("id_card"),
			c.FormValue("id_card_issue_date"), c.FormValue("id_card_expiry_date"), c.FormValue("date_of_birth"),
			c.FormValue("ethnicity"), c.FormValue("nationality"), c.FormValue("religion"), c.FormValue("marital_status"), c.FormValue("mobile_phone"),
			c.FormValue("house_reg_no"), c.FormValue("house_reg_moo"), c.FormValue("house_reg_soi"), c.FormValue("house_reg_road"), c.FormValue("house_reg_province"), c.FormValue("house_reg_district"), c.FormValue("house_reg_subdistrict"), c.FormValue("house_reg_zipcode"),
			c.FormValue("same_as_house_reg") == "on",
			c.FormValue("current_no"), c.FormValue("current_moo"), c.FormValue("current_soi"), c.FormValue("current_road"), c.FormValue("current_province"), c.FormValue("current_district"), c.FormValue("current_subdistrict"), c.FormValue("current_zipcode"),
			companyName, c.FormValue("occupation"), c.FormValue("position"), parseMoney(c.FormValue("salary")), parseMoney(c.FormValue("other_income")), c.FormValue("income_source"),
			c.FormValue("work_phone"), c.FormValue("work_no"), c.FormValue("work_moo"), c.FormValue("work_soi"), c.FormValue("work_road"), c.FormValue("work_province"), c.FormValue("work_district"), c.FormValue("work_subdistrict"), c.FormValue("work_zipcode"),
			c.FormValue("other_card_type"), c.FormValue("other_card_number"), c.FormValue("other_card_issue_date"), c.FormValue("other_card_expiry_date"),
			c.FormValue("doc_delivery_type"), c.FormValue("doc_no"), c.FormValue("doc_moo"), c.FormValue("doc_soi"), c.FormValue("doc_road"), c.FormValue("doc_province"), c.FormValue("doc_district"), c.FormValue("doc_subdistrict"), c.FormValue("doc_zipcode"),
			guarantorID,
		).Error

		if err != nil {
			log.Printf("[guarantor] error updating ID %s: %v", guarantorID, err)
			return c.Status(500).SendString("Error updating guarantor (V2): " + err.Error())
		}
		return c.Redirect("/step4?id=" + loanIDStr)
	}

	// Explicitly use strict Raw SQL with NULLIF to prevent "Incorrect date value"
	query := `INSERT INTO loan_applications_guarantors (
		created_at, updated_at, loan_id,
		guarantor_type, trade_registration_id, registration_date, tax_id,
		title, first_name, last_name, gender, id_card,
		id_card_issue_date, id_card_expiry_date, date_of_birth,
		ethnicity, nationality, religion, marital_status, mobile_phone,
		house_reg_no, house_reg_moo, house_reg_soi, house_reg_road, house_reg_province, house_reg_district, house_reg_subdistrict, house_reg_zipcode,
		same_as_house_reg,
		current_no, current_moo, current_soi, current_road, current_province, current_district, current_subdistrict, current_zipcode,
		company_name, occupation, position, salary, other_income, income_source,
		work_phone, work_no, work_moo, work_soi, work_road, work_province, work_district, work_subdistrict, work_zipcode,
		other_card_type, other_card_number, other_card_issue_date, other_card_expiry_date,
		doc_delivery_type, doc_no, doc_moo, doc_soi, doc_road, doc_province, doc_district, doc_subdistrict, doc_zipcode
	) VALUES (
		NOW(), NOW(), ?,
		?, ?, NULLIF(?, ''), ?,
		?, ?, ?, ?, ?,
		NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''),
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, NULLIF(?, ''), NULLIF(?, ''),
		?, ?, ?, ?, ?, ?, ?, ?, ?
	)`

	// Logic to pick the correct company name
	// If Juristic -> uses "juristic_company_name"
	// If Individual -> uses "work_company_name" (which maps to db company_name column contextually for work info)
	var companyName string
	if c.FormValue("guarantor_type") == "juristic" {
		companyName = c.FormValue("juristic_company_name")
	} else {
		companyName = c.FormValue("work_company_name")
	}

	err := config.DB.Exec(query,
		loanIDStr, // loan_id
		c.FormValue("guarantor_type"),
		c.FormValue("trade_registration_id"), c.FormValue("registration_date"), c.FormValue("tax_id"),
		c.FormValue("title"), c.FormValue("first_name"), c.FormValue("last_name"), c.FormValue("gender"), c.FormValue("id_card"),
		c.FormValue("id_card_issue_date"), c.FormValue("id_card_expiry_date"), c.FormValue("date_of_birth"),
		c.FormValue("ethnicity"), c.FormValue("nationality"), c.FormValue("religion"), c.FormValue("marital_status"), c.FormValue("mobile_phone"),
		c.FormValue("house_reg_no"), c.FormValue("house_reg_moo"), c.FormValue("house_reg_soi"), c.FormValue("house_reg_road"), c.FormValue("house_reg_province"), c.FormValue("house_reg_district"), c.FormValue("house_reg_subdistrict"), c.FormValue("house_reg_zipcode"),
		c.FormValue("same_as_house_reg") == "on",
		c.FormValue("current_no"), c.FormValue("current_moo"), c.FormValue("current_soi"), c.FormValue("current_road"), c.FormValue("current_province"), c.FormValue("current_district"), c.FormValue("current_subdistrict"), c.FormValue("current_zipcode"),
		companyName, c.FormValue("occupation"), c.FormValue("position"), parseMoney(c.FormValue("salary")), parseMoney(c.FormValue("other_income")), c.FormValue("income_source"),
		c.FormValue("work_phone"), c.FormValue("work_no"), c.FormValue("work_moo"), c.FormValue("work_soi"), c.FormValue("work_road"), c.FormValue("work_province"), c.FormValue("work_district"), c.FormValue("work_subdistrict"), c.FormValue("work_zipcode"),
		c.FormValue("other_card_type"), c.FormValue("other_card_number"), c.FormValue("other_card_issue_date"), c.FormValue("other_card_expiry_date"),
		c.FormValue("doc_delivery_type"), c.FormValue("doc_no"), c.FormValue("doc_moo"), c.FormValue("doc_soi"), c.FormValue("doc_road"), c.FormValue("doc_province"), c.FormValue("doc_district"), c.FormValue("doc_subdistrict"), c.FormValue("doc_zipcode"),
	).Error

	if err != nil {
		log.Printf("[guarantor] error inserting: %v", err)
		return c.Status(500).SendString("Error saving guarantor (V2): " + err.Error())
	}

	// Update loan_applications to uncheck NoGuarantor
	updateLoanSQL := "UPDATE loan_applications SET no_guarantor = false, last_update_date = NOW() WHERE id = ?"
	if err := config.DB.Exec(updateLoanSQL, loanIDStr).Error; err != nil {
		log.Printf("[guarantor] error updating no_guarantor for loan %s: %v", loanIDStr, err)
	}

	// Create model just to check ID for redirect (optional, or just use loanIDStr)
	// Actually we can just redirect using loanIDStr
	return c.Redirect("/step4?id=" + loanIDStr)
}

// DeleteGuarantor handles deleting a guarantor by ID
func DeleteGuarantor(c *fiber.Ctx) error {
	id := c.FormValue("id")
	loanID := c.FormValue("loan_id")

	if id == "" || loanID == "" {
		return c.Status(400).SendString("Missing ID or Loan ID")
	}

	// Delete from database
	if err := config.DB.Exec("DELETE FROM loan_applications_guarantors WHERE id = ?", id).Error; err != nil {
		log.Printf("[guarantor] error deleting ID %s: %v", id, err)
		return c.Status(500).SendString("Error deleting guarantor")
	}

	return c.Redirect("/step4?id=" + loanID)
}

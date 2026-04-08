package handlers

import (
	"log"
	"strconv"
	"strings"

	"loan-app/config"
	"loan-app/services"

	"github.com/gofiber/fiber/v2"
)

func parseMoney(amount string) float64 {
	amount = strings.ReplaceAll(amount, ",", "")
	val, _ := strconv.ParseFloat(amount, 64)
	return val
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func guarantorInputFromRequest(c *fiber.Ctx) services.GuarantorInput {
	companyName := c.FormValue("work_company_name")
	if c.FormValue("guarantor_type") == "juristic" {
		companyName = c.FormValue("juristic_company_name")
	}

	return services.GuarantorInput{
		GuarantorType:       c.FormValue("guarantor_type"),
		TradeRegistrationID: c.FormValue("trade_registration_id"),
		RegistrationDate:    c.FormValue("registration_date"),
		TaxID:               c.FormValue("tax_id"),
		Title:               c.FormValue("title"),
		FirstName:           c.FormValue("first_name"),
		LastName:            c.FormValue("last_name"),
		Gender:              c.FormValue("gender"),
		IdCard:              c.FormValue("id_card"),
		IdCardIssueDate:     c.FormValue("id_card_issue_date"),
		IdCardExpiryDate:    c.FormValue("id_card_expiry_date"),
		DateOfBirth:         c.FormValue("date_of_birth"),
		Ethnicity:           c.FormValue("ethnicity"),
		Nationality:         c.FormValue("nationality"),
		Religion:            c.FormValue("religion"),
		MaritalStatus:       c.FormValue("marital_status"),
		MobilePhone:         c.FormValue("mobile_phone"),
		HouseRegNo:          c.FormValue("house_reg_no"),
		HouseRegMoo:         c.FormValue("house_reg_moo"),
		HouseRegSoi:         c.FormValue("house_reg_soi"),
		HouseRegRoad:        c.FormValue("house_reg_road"),
		HouseRegProvince:    c.FormValue("house_reg_province"),
		HouseRegDistrict:    c.FormValue("house_reg_district"),
		HouseRegSubdistrict: c.FormValue("house_reg_subdistrict"),
		HouseRegZipcode:     c.FormValue("house_reg_zipcode"),
		SameAsHouseReg:      c.FormValue("same_as_house_reg") == "on",
		CurrentNo:           c.FormValue("current_no"),
		CurrentMoo:          c.FormValue("current_moo"),
		CurrentSoi:          c.FormValue("current_soi"),
		CurrentRoad:         c.FormValue("current_road"),
		CurrentProvince:     c.FormValue("current_province"),
		CurrentDistrict:     c.FormValue("current_district"),
		CurrentSubdistrict:  c.FormValue("current_subdistrict"),
		CurrentZipcode:      c.FormValue("current_zipcode"),
		CompanyName:         companyName,
		Occupation:          c.FormValue("occupation"),
		Position:            c.FormValue("position"),
		Salary:              parseMoney(c.FormValue("salary")),
		OtherIncome:         parseMoney(c.FormValue("other_income")),
		IncomeSource:        c.FormValue("income_source"),
		WorkPhone:           c.FormValue("work_phone"),
		WorkNo:              c.FormValue("work_no"),
		WorkMoo:             c.FormValue("work_moo"),
		WorkSoi:             c.FormValue("work_soi"),
		WorkRoad:            c.FormValue("work_road"),
		WorkProvince:        c.FormValue("work_province"),
		WorkDistrict:        c.FormValue("work_district"),
		WorkSubdistrict:     c.FormValue("work_subdistrict"),
		WorkZipcode:         c.FormValue("work_zipcode"),
		OtherCardType:       c.FormValue("other_card_type"),
		OtherCardNumber:     c.FormValue("other_card_number"),
		OtherCardIssueDate:  c.FormValue("other_card_issue_date"),
		OtherCardExpiryDate: c.FormValue("other_card_expiry_date"),
		DocDeliveryType:     c.FormValue("doc_delivery_type"),
		DocNo:               c.FormValue("doc_no"),
		DocMoo:              c.FormValue("doc_moo"),
		DocSoi:              c.FormValue("doc_soi"),
		DocRoad:             c.FormValue("doc_road"),
		DocProvince:         c.FormValue("doc_province"),
		DocDistrict:         c.FormValue("doc_district"),
		DocSubdistrict:      c.FormValue("doc_subdistrict"),
		DocZipcode:          c.FormValue("doc_zipcode"),
	}
}

func AddGuarantorGetV2(c *fiber.Ctx) error {
	loanID := c.Query("loan_id")
	guarantorID := c.Query("guarantor_id")

	if loanID == "" {
		return c.Redirect("/")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		return c.Redirect("/main")
	}

	guarantorData := make(map[string]interface{})
	if guarantorID != "" {
		if guarantor, err := services.FindGuarantorByLoan(config.DB, loan.ID, guarantorID); err == nil {
			guarantorData["ID"] = guarantor.ID
			guarantorData["Title"] = guarantor.Title
			guarantorData["FirstName"] = guarantor.FirstName
			guarantorData["LastName"] = guarantor.LastName
			guarantorData["Gender"] = guarantor.Gender
			guarantorData["IdCard"] = guarantor.IdCard
			guarantorData["IdCardIssueDate"] = safeStr(guarantor.IdCardIssueDate)
			guarantorData["IdCardExpiryDate"] = safeStr(guarantor.IdCardExpiryDate)
			guarantorData["DateOfBirth"] = safeStr(guarantor.DateOfBirth)
			guarantorData["Ethnicity"] = guarantor.Ethnicity
			guarantorData["Nationality"] = guarantor.Nationality
			guarantorData["Religion"] = guarantor.Religion
			guarantorData["MaritalStatus"] = guarantor.MaritalStatus
			guarantorData["MobilePhone"] = guarantor.MobilePhone
			guarantorData["HouseRegNo"] = guarantor.HouseRegNo
			guarantorData["HouseRegMoo"] = guarantor.HouseRegMoo
			guarantorData["HouseRegSoi"] = guarantor.HouseRegSoi
			guarantorData["HouseRegRoad"] = guarantor.HouseRegRoad
			guarantorData["HouseRegProvince"] = guarantor.HouseRegProvince
			guarantorData["HouseRegDistrict"] = guarantor.HouseRegDistrict
			guarantorData["HouseRegSubdistrict"] = guarantor.HouseRegSubdistrict
			guarantorData["HouseRegZipcode"] = guarantor.HouseRegZipcode
			guarantorData["SameAsHouseReg"] = guarantor.SameAsHouseReg
			guarantorData["CurrentNo"] = guarantor.CurrentNo
			guarantorData["CurrentMoo"] = guarantor.CurrentMoo
			guarantorData["CurrentSoi"] = guarantor.CurrentSoi
			guarantorData["CurrentRoad"] = guarantor.CurrentRoad
			guarantorData["CurrentProvince"] = guarantor.CurrentProvince
			guarantorData["CurrentDistrict"] = guarantor.CurrentDistrict
			guarantorData["CurrentSubdistrict"] = guarantor.CurrentSubdistrict
			guarantorData["CurrentZipcode"] = guarantor.CurrentZipcode
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
			guarantorData["OtherCardType"] = guarantor.OtherCardType
			guarantorData["OtherCardNumber"] = guarantor.OtherCardNumber
			guarantorData["OtherCardIssueDate"] = safeStr(guarantor.OtherCardIssueDate)
			guarantorData["OtherCardExpiryDate"] = safeStr(guarantor.OtherCardExpiryDate)
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
		"Guarantor": guarantorData,
		"IsEdit":    guarantorID != "",
	})
}

func AddGuarantorPostV2(c *fiber.Ctx) error {
	loanIDStr := c.FormValue("loan_id")
	guarantorID := c.FormValue("guarantor_id")

	loan, err := requireLoanAccess(c, loanIDStr)
	if err != nil {
		return c.Status(403).SendString("Forbidden")
	}

	err = services.SaveGuarantor(config.DB, loan.ID, guarantorID, guarantorInputFromRequest(c))
	if err != nil {
		log.Printf("[guarantor] error saving ID %s: %v", guarantorID, err)
		return c.Status(500).SendString("Error saving guarantor (V2): " + err.Error())
	}

	return c.Redirect("/step4?id=" + loanIDStr)
}

func DeleteGuarantor(c *fiber.Ctx) error {
	id := c.FormValue("id")
	loanID := c.FormValue("loan_id")
	if id == "" || loanID == "" {
		return c.Status(400).SendString("Missing ID or Loan ID")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		return c.Status(403).SendString("Forbidden")
	}

	if err := services.DeleteGuarantor(config.DB, loan.ID, id); err != nil {
		log.Printf("[guarantor] error deleting ID %s: %v", id, err)
		return c.Status(500).SendString("Error deleting guarantor")
	}

	return c.Redirect("/step4?id=" + loanID)
}

package handlers

import (
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"loan-app/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Helper to handle optional dates
func safeDate(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func loanService() *services.LoanService {
	return services.NewDefaultLoanService(config.DB)
}

func parseMoneyValue(v string) float64 {
	return services.ParseMoney(v)
}

func step1InputFromRequest(c *fiber.Ctx, salary, otherIncome float64) services.Step1Input {
	companyName := c.FormValue("work_company_name")
	if c.FormValue("borrower_type") == "juristic" {
		companyName = c.FormValue("juristic_company_name")
	}

	return services.Step1Input{
		Title:               c.FormValue("title"),
		FirstName:           c.FormValue("first_name"),
		LastName:            c.FormValue("last_name"),
		Gender:              c.FormValue("gender"),
		BorrowerType:        c.FormValue("borrower_type"),
		TradeRegistrationID: c.FormValue("trade_registration_id"),
		RegistrationDate:    c.FormValue("registration_date"),
		TaxID:               c.FormValue("tax_id"),
		CompanyName:         companyName,
		IdCard:              c.FormValue("id_card"),
		IdCardIssueDate:     c.FormValue("id_card_issue_date"),
		IdCardExpiryDate:    c.FormValue("id_card_expiry_date"),
		DateOfBirth:         c.FormValue("date_of_birth"),
		Ethnicity:           c.FormValue("ethnicity"),
		Nationality:         c.FormValue("nationality"),
		Religion:            c.FormValue("religion"),
		MaritalStatus:       c.FormValue("marital_status"),
		MobilePhone:         c.FormValue("mobile_phone"),
		OtherCardType:       c.FormValue("other_card_type"),
		OtherCardNumber:     c.FormValue("other_card_number"),
		OtherCardIssueDate:  c.FormValue("other_card_issue_date"),
		OtherCardExpiryDate: c.FormValue("other_card_expiry_date"),
		HouseRegNo:          c.FormValue("house_reg_no"),
		HouseRegBuilding:    c.FormValue("house_reg_building"),
		HouseRegFloor:       c.FormValue("house_reg_floor"),
		HouseRegRoom:        c.FormValue("house_reg_room"),
		HouseRegMoo:         c.FormValue("house_reg_moo"),
		HouseRegSoi:         c.FormValue("house_reg_soi"),
		HouseRegRoad:        c.FormValue("house_reg_road"),
		HouseRegProvince:    c.FormValue("house_reg_province"),
		HouseRegDistrict:    c.FormValue("house_reg_district"),
		HouseRegSubdistrict: c.FormValue("house_reg_subdistrict"),
		HouseRegZipcode:     c.FormValue("house_reg_zipcode"),
		SameAsHouseReg:      c.FormValue("same_as_house_reg") == "on",
		CurrentCompany:      c.FormValue("current_company"),
		CurrentNo:           c.FormValue("current_no"),
		CurrentBuilding:     c.FormValue("current_building"),
		CurrentFloor:        c.FormValue("current_floor"),
		CurrentRoom:         c.FormValue("current_room"),
		CurrentMoo:          c.FormValue("current_moo"),
		CurrentSoi:          c.FormValue("current_soi"),
		CurrentRoad:         c.FormValue("current_road"),
		CurrentProvince:     c.FormValue("current_province"),
		CurrentDistrict:     c.FormValue("current_district"),
		CurrentSubdistrict:  c.FormValue("current_subdistrict"),
		CurrentZipcode:      c.FormValue("current_zipcode"),
		WorkNo:              c.FormValue("work_no"),
		WorkBuilding:        c.FormValue("work_building"),
		WorkFloor:           c.FormValue("work_floor"),
		WorkRoom:            c.FormValue("work_room"),
		WorkMoo:             c.FormValue("work_moo"),
		WorkSoi:             c.FormValue("work_soi"),
		WorkRoad:            c.FormValue("work_road"),
		WorkProvince:        c.FormValue("work_province"),
		WorkDistrict:        c.FormValue("work_district"),
		WorkSubdistrict:     c.FormValue("work_subdistrict"),
		WorkZipcode:         c.FormValue("work_zipcode"),
		WorkPhone:           c.FormValue("work_phone"),
		DocDeliveryType:     c.FormValue("doc_delivery_type"),
		DocNo:               c.FormValue("doc_no"),
		DocBuilding:         c.FormValue("doc_building"),
		DocFloor:            c.FormValue("doc_floor"),
		DocRoom:             c.FormValue("doc_room"),
		DocMoo:              c.FormValue("doc_moo"),
		DocSoi:              c.FormValue("doc_soi"),
		DocRoad:             c.FormValue("doc_road"),
		DocProvince:         c.FormValue("doc_province"),
		DocDistrict:         c.FormValue("doc_district"),
		DocSubdistrict:      c.FormValue("doc_subdistrict"),
		DocZipcode:          c.FormValue("doc_zipcode"),
		Occupation:          c.FormValue("occupation"),
		Position:            c.FormValue("position"),
		Salary:              salary,
		OtherIncome:         otherIncome,
		IncomeSource:        c.FormValue("income_source"),
		CreditBureauStatus:  c.FormValue("credit_bureau_status"),
	}
}

func step2InputFromRequest(c *fiber.Ctx) services.Step2Input {
	return services.Step2Input{
		CarType:         c.FormValue("car_type"),
		CarCode:         c.FormValue("car_code"),
		CarBrand:        c.FormValue("car_brand"),
		CarRegisterDate: c.FormValue("car_register_date"),
		CarModel:        c.FormValue("car_model"),
		CarYear:         c.FormValue("car_year"),
		CarColor:        c.FormValue("car_color"),
		CarWeight:       parseMoneyValue(c.FormValue("car_weight")),
		CarCC:           int(parseMoneyValue(c.FormValue("car_cc"))),
		CarMileage:      parseMoneyValue(c.FormValue("car_mileage")),
		CarChassisNo:    c.FormValue("car_chassis_no"),
		CarGear:         c.FormValue("car_gear"),
		CarEngineNo:     c.FormValue("car_engine_no"),
		CarCondition:    c.FormValue("car_condition"),
		LicensePlate:    c.FormValue("license_plate"),
		LicenseProvince: c.FormValue("license_province"),
		VatRate:         parseMoneyValue(c.FormValue("vat_rate")),
		CarPrice:        parseMoneyValue(c.FormValue("car_price")),
		IsRefinance:     c.FormValue("is_refinance") == "on",
	}
}

func step3InputFromRequest(c *fiber.Ctx) services.Step3Input {
	interestRate, _ := strconv.ParseFloat(c.FormValue("interest_rate"), 64)
	installments, _ := strconv.Atoi(c.FormValue("installments"))
	paymentDay, _ := strconv.Atoi(c.FormValue("payment_day"))
	lifeInterestRate, _ := strconv.ParseFloat(c.FormValue("life_interest_rate"), 64)
	lifeInstallments, _ := strconv.Atoi(c.FormValue("life_installments"))

	return services.Step3Input{
		ContractSignDate:         c.FormValue("contract_sign_date"),
		LoanType:                 c.FormValue("loan_type"),
		LoanAmount:               parseMoneyValue(c.FormValue("loan_amount")),
		InterestRate:             interestRate,
		Installments:             installments,
		InstallmentAmount:        parseMoneyValue(c.FormValue("installment_amount")),
		DownPayment:              parseMoneyValue(c.FormValue("down_payment")),
		ContractStartDate:        c.FormValue("contract_start_date"),
		FirstPaymentDate:         c.FormValue("first_payment_date"),
		TransferType:             c.FormValue("transfer_type"),
		TransferFee:              parseMoneyValue(c.FormValue("transfer_fee")),
		TaxFee:                   parseMoneyValue(c.FormValue("tax_fee")),
		DutyFee:                  parseMoneyValue(c.FormValue("duty_fee")),
		PaymentDay:               paymentDay,
		HasLifeInsurance:         c.FormValue("is_life_insurance") == "true",
		CarInsuranceBeginning:    parseMoneyValue(c.FormValue("beginning_amount")),
		CarInsuranceRefinanceFee: parseMoneyValue(c.FormValue("refinance_fee")),
		LifeLoanPrincipal:        parseMoneyValue(c.FormValue("life_loan_principal")),
		LifeInterestRate:         lifeInterestRate,
		LifeInstallments:         lifeInstallments,
		LifeInsuranceCompany:     c.FormValue("life_insurance_company"),
		LifeInsuranceRate:        parseMoneyValue(c.FormValue("life_premium_rate")),
		LifePremium:              parseMoneyValue(c.FormValue("insurance_premium")),
	}
}

func step4InputFromRequest(c *fiber.Ctx) services.Step4Input {
	return services.Step4Input{
		NoGuarantor:       c.FormValue("no_guarantor") == "on",
		Guarantor1Name:    c.FormValue("guarantor1_name"),
		Guarantor1Contact: c.FormValue("guarantor1_contact"),
		Guarantor2Name:    c.FormValue("guarantor2_name"),
		Guarantor2Contact: c.FormValue("guarantor2_contact"),
		Guarantor3Name:    c.FormValue("guarantor3_name"),
		Guarantor3Contact: c.FormValue("guarantor3_contact"),
		LoanOfficer:       c.FormValue("loan_officer"),
		CompanySeller:     c.FormValue("company_seller"),
		CompanySellerID:   c.FormValue("company_seller_id"),
		ShowroomStaff:     c.FormValue("showroom_staff"),
		Commission:        parseMoneyValue(c.FormValue("commission")),
		ScoreOfficer:      parseMoneyValue(c.FormValue("score_officer")),
		ScoreManager:      parseMoneyValue(c.FormValue("score_manager")),
	}
}

func step5InputFromRequest(c *fiber.Ctx) services.Step5Input {
	lifeInterestRate, _ := strconv.ParseFloat(c.FormValue("life_interest_rate"), 64)
	lifeInstallments, _ := strconv.Atoi(c.FormValue("life_installments"))

	return services.Step5Input{
		HasLifeInsurance:     c.FormValue("is_life_insurance") == "true",
		LifeInsuranceCompany: c.FormValue("life_insurance_company"),
		LifeLoanPrincipal:    parseMoneyValue(c.FormValue("life_loan_amount")),
		LifeInterestRate:     lifeInterestRate,
		LifeInstallments:     lifeInstallments,
		LifeGender:           c.FormValue("life_gender"),
		LifeDob:              c.FormValue("life_dob"),
		LifeSignDate:         c.FormValue("life_sign_date"),
		LifeInsuranceRate:    parseMoneyValue(c.FormValue("life_premium_rate")),
		LifePremium:          parseMoneyValue(c.FormValue("life_premium")),
		Beneficiary1Name:     c.FormValue("beneficiary1_name"),
		Beneficiary1Relation: c.FormValue("beneficiary1_relation"),
		Beneficiary1Address:  c.FormValue("beneficiary1_address"),
		Beneficiary2Name:     c.FormValue("beneficiary2_name"),
		Beneficiary2Relation: c.FormValue("beneficiary2_relation"),
		Beneficiary2Address:  c.FormValue("beneficiary2_address"),
		Beneficiary3Name:     c.FormValue("beneficiary3_name"),
		Beneficiary3Relation: c.FormValue("beneficiary3_relation"),
		Beneficiary3Address:  c.FormValue("beneficiary3_address"),
		InsuranceSeller:      c.FormValue("insurance_agent"),
		InsuranceAgentEmpID:  c.FormValue("insurance_agent_empid"),
		InsuranceLicenseNo:   c.FormValue("agent_license"),
	}
}

func step6InputFromRequest(c *fiber.Ctx) services.Step6Input {
	carInsurancePremium, _ := strconv.ParseFloat(c.FormValue("insurance_cost"), 64)
	carInsuranceAvoidanceFee, _ := strconv.ParseFloat(c.FormValue("avoidance_fee"), 64)

	return services.Step6Input{
		CarInsuranceType:         c.FormValue("insurance_type"),
		CarInsuranceCompany:      c.FormValue("insurance_company"),
		CarInsuranceClass:        c.FormValue("insurance_class"),
		CarInsuranceNotifyDate:   c.FormValue("notification_date"),
		CarInsuranceNotifyNo:     c.FormValue("notification_number"),
		CarInsuranceStartDate:    c.FormValue("coverage_start_date"),
		CarInsurancePremium:      carInsurancePremium,
		CarInsuranceAvoidanceFee: carInsuranceAvoidanceFee,
	}
}

func step7InputFromRequest(c *fiber.Ctx) services.Step7Input {
	return services.Step7Input{
		TaxPayerType:   c.FormValue("tax_payer_type"),
		TaxIdCard:      c.FormValue("tax_id_card"),
		TaxPrefix:      c.FormValue("tax_prefix"),
		TaxFirstName:   c.FormValue("tax_first_name"),
		TaxLastName:    c.FormValue("tax_last_name"),
		TaxHouseNo:     c.FormValue("tax_house_no"),
		TaxBuilding:    c.FormValue("tax_building"),
		TaxFloor:       c.FormValue("tax_floor"),
		TaxRoom:        c.FormValue("tax_room"),
		TaxVillage:     c.FormValue("tax_village"),
		TaxMoo:         c.FormValue("tax_moo"),
		TaxSoi:         c.FormValue("tax_soi"),
		TaxRoad:        c.FormValue("tax_road"),
		TaxProvince:    c.FormValue("tax_province"),
		TaxDistrict:    c.FormValue("tax_district"),
		TaxSubdistrict: c.FormValue("tax_sub_district"),
		TaxZipcode:     c.FormValue("tax_zipcode"),
	}
}

func Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", nil)
}

// API: Get loan list as JSON for the mobile app.
func GetLoanList(c *fiber.Ctx) error {
	staffID := parseJWTUsername(c.Cookies("token"))
	loans, err := loanService().ListByStaff(staffID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load loans"})
	}

	return c.JSON(fiber.Map{
		"loans": loans,
		"total": len(loans),
	})
}

func MainPage(c *fiber.Ctx) error {
	staffID := parseJWTUsername(c.Cookies("token"))
	loans, err := loanService().ListByStaff(staffID)
	if err != nil {
		return c.Status(500).SendString("Failed to load loans")
	}

	return c.Render("main", fiber.Map{
		"title":       "Main - CMO APP",
		"Loans":       loans,
		"StaffID":     staffID,
		"CurrentRole": getUserRole(staffID),
	})
}

/* Step 1: borrower profile and contact details */
func Step1(c *fiber.Ctx) error {
	id := c.Query("id")
	var loan models.LoanApplication

	// If ID is provided (Edit Mode)
	if id != "" {
		if accessibleLoan, err := requireLoanAccess(c, id); err == nil {
			loan = *accessibleLoan
			// Set cookie
			c.Cookie(&fiber.Cookie{
				Name:  "loan_id",
				Value: fmt.Sprintf("%d", loan.ID),
			})
			return c.Render("step1", fiber.Map{
				"title": "Step 1",
				"Loan":  loan,
			})
		}
	}

	// If no ID (Add Mode), clear cookie
	c.ClearCookie("loan_id")
	return c.Render("step1", fiber.Map{
		"title": "Step 1",
		"Loan":  models.LoanApplication{},
	})
}

func Step1Post(c *fiber.Ctx) error {
	salary := parseMoneyValue(c.FormValue("salary"))
	otherIncome := parseMoneyValue(c.FormValue("other_income"))

	// Server-side Validation
	borrowerType := c.FormValue("borrower_type")
	if borrowerType == "juristic" {
		if c.FormValue("company_name") == "" || c.FormValue("trade_registration_id") == "" {
			return c.Status(400).SendString("Missing required fields for Juristic Person: company_name or trade_registration_id")
		}
	} else {
		// Individual
		if c.FormValue("first_name") == "" || c.FormValue("last_name") == "" || c.FormValue("id_card") == "" {
			return c.Status(400).SendString("Missing required fields: first_name, last_name, or id_card")
		}
	}

	// Parse Staff ID from token
	staffID := parseJWTUsername(c.Cookies("token"))

	// Check if updating existing loan
	cookieID := c.Cookies("loan_id")
	if cookieID != "" {
		existingLoan, err := requireLoanAccess(c, cookieID)
		if err != nil {
			clearAuthCookie(c)
			return c.Redirect("/step1")
		}
		if existingLoan != nil {
			if err := loanService().UpdateStep1(existingLoan, staffID, step1InputFromRequest(c, salary, otherIncome)); err != nil {
				return c.Status(500).SendString(err.Error())
			}
			return c.Redirect("/step2")
		}
	}

	loan, err := loanService().CreateStep1(staffID, step1InputFromRequest(c, salary, otherIncome))
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	WriteAudit(c, "create_loan", loan.RefCode, loan.FirstName+" "+loan.LastName)

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:  "loan_id",
		Value: fmt.Sprintf("%d", loan.ID),
	})

	return c.Redirect("/step2")
}

/* Step 2: loan objective and requested amount */
func Step2(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		loan = *accessibleLoan
	}

	var carKinds []models.CarKind
	config.DB.Find(&carKinds)

	var carBrands []models.CarBrand
	config.DB.Find(&carBrands)

	return c.Render("step2", fiber.Map{
		"title":     "Step 2",
		"Loan":      loan,
		"CarKinds":  carKinds,
		"CarBrands": carBrands,
	})
}

func Step2Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep2(loan, step2InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Redirect("/step3")
}

/* Step 3: collateral and repayment details */
func Step3(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		loan = *accessibleLoan
	}
	return c.Render("step3", fiber.Map{
		"title": "Step 3",
		"Loan":  loan,
	})
}

func Step3Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep3(loan, step3InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Redirect("/step4")
}

/* Step 4: guarantor and reference information */
func Step4(c *fiber.Ctx) error {
	// Support ?id= query param (e.g. redirect back from add_guarantor) or cookie
	loanID := c.Query("id")
	if loanID == "" {
		loanID = c.Cookies("loan_id")
	} else {
		// Update cookie so subsequent navigation stays consistent
		c.Cookie(&fiber.Cookie{
			Name:  "loan_id",
			Value: loanID,
		})
	}

	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		if loadedLoan, err := loanService().FindWithGuarantors(accessibleLoan.ID); err == nil {
			loan = *loadedLoan
		}
	}
	return c.Render("step4", fiber.Map{
		"title": "Step 4",
		"Loan":  loan,
	})
}

func Step4Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep4(loan, step4InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Redirect("/step5")
}

/* Step 5: insurance and protection options */
func Step5(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		loan = *accessibleLoan
	}
	return c.Render("step5", fiber.Map{
		"title": "Step 5",
		"Loan":  loan,
	})
}

func Step5Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep5(loan, step5InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Redirect("/step6")
}

/* Step 6: document checklist and upload confirmation */
func Step6(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		loan = *accessibleLoan
	}
	return c.Render("step6", fiber.Map{
		"title": "Step 6",
		"Loan":  loan,
	})
}

func Step6Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep6(loan, step6InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.Redirect("/step7")
}

/* Step 7: tax invoice and final submission */
func Step7(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		accessibleLoan, err := requireLoanAccess(c, loanID)
		if err != nil {
			c.ClearCookie("loan_id")
			return c.Redirect("/step1")
		}
		loan = *accessibleLoan
	}
	return c.Render("step7", fiber.Map{
		"title": "Step 7",
		"Loan":  loan,
	})
}

func Step7Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		clearAuthCookie(c)
		return c.Redirect("/step1")
	}

	if err := loanService().UpdateStep7(loan, step7InputFromRequest(c)); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	WriteAudit(c, "submit_loan", loan.RefCode, "Completed step 7")

	// Clear cookie after finish
	c.ClearCookie("loan_id")

	return c.Redirect("/main")
}

func UpdateStatus(c *fiber.Ctx) error {
	type Request struct {
		RefCode string `json:"ref_code"`
		Status  string `json:"status"` // A=Approved, R=Rejected, C=Conditional
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	loan, err := loanService().UpdateStatus(req.RefCode, req.Status)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	// Notify the assigned staff member once the status changes.
	borrowerName := loan.FirstName + " " + loan.LastName
	if borrowerName == " " {
		borrowerName = loan.RefCode
	}
	switch loan.Status {
	case "A":
		BroadcastToStaff(loan.StaffID, "Loan approved",
			fmt.Sprintf("Ref %s - %s has been approved", loan.RefCode, borrowerName))
	case "R":
		BroadcastToStaff(loan.StaffID, "Loan rejected",
			fmt.Sprintf("Ref %s - %s has been rejected", loan.RefCode, borrowerName))
	case "C":
		BroadcastToStaff(loan.StaffID, "Additional conditions required",
			fmt.Sprintf("Ref %s - %s requires additional conditions", loan.RefCode, borrowerName))
	}

	return c.JSON(fiber.Map{"success": true})
}

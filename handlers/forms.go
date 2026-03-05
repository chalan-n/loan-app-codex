package handlers

import (
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"strconv"
	"strings"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Helper to handle optional dates
func safeDate(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", nil)
}

// 📱 API: Get Loan List as JSON (สำหรับ Mobile App)
func GetLoanList(c *fiber.Ctx) error {
	// 1. Get User from Token (same as MainPage)
	tokenStr := c.Cookies("token")
	var staffID string
	if tokenStr != "" {
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("mysecret"), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			staffID, _ = claims["username"].(string)
		}
	}

	// 2. Query loans for this user
	var loans []models.LoanApplication
	if staffID != "" {
		config.DB.Where("staff_id = ?", staffID).Order("id desc").Find(&loans)
	} else {
		loans = []models.LoanApplication{}
	}

	// 3. Return JSON
	return c.JSON(fiber.Map{
		"loans": loans,
		"total": len(loans),
	})
}

func MainPage(c *fiber.Ctx) error {
	// 1. Get User from Token
	tokenStr := c.Cookies("token")
	var staffID string
	if tokenStr != "" {
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("mysecret"), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			staffID, _ = claims["username"].(string)
		}
	}

	// 2. Filter query
	var loans []models.LoanApplication
	if staffID != "" {
		config.DB.Where("staff_id = ?", staffID).Order("id desc").Find(&loans)
	} else {
		loans = []models.LoanApplication{}
	}

	return c.Render("main", fiber.Map{
		"title":   "หน้าหลัก - CMO APP",
		"Loans":   loans,
		"StaffID": staffID,
	})
}

/* ข้อมูลผู้เช่าซื้อ */
func Step1(c *fiber.Ctx) error {
	id := c.Query("id")
	var loan models.LoanApplication

	// If ID is provided (Edit Mode)
	if id != "" {
		if err := config.DB.First(&loan, id).Error; err == nil {
			// Set cookie
			c.Cookie(&fiber.Cookie{
				Name:  "loan_id",
				Value: fmt.Sprintf("%d", loan.ID),
			})
			return c.Render("step1", fiber.Map{
				"title": "ขั้นตอน 1",
				"Loan":  loan,
			})
		}
	}

	// If no ID (Add Mode), clear cookie
	c.ClearCookie("loan_id")
	return c.Render("step1", fiber.Map{
		"title": "ขั้นตอน 1",
		"Loan":  models.LoanApplication{},
	})
}

func Step1Post(c *fiber.Ctx) error {
	// Helper to parse currency (strip commas)

	// Helper to parse currency (strip commas)
	parseMoney := func(v string) float64 {
		v = strings.ReplaceAll(v, ",", "")
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	salary := parseMoney(c.FormValue("salary"))
	otherIncome := parseMoney(c.FormValue("other_income"))

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
	tokenStr := c.Cookies("token")
	var staffID string
	if tokenStr != "" {
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("mysecret"), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			staffID, _ = claims["username"].(string)
		}
	}

	// Check if updating existing loan
	cookieID := c.Cookies("loan_id")
	if cookieID != "" {
		// Update Mode
		var existingLoan models.LoanApplication
		if err := config.DB.First(&existingLoan, cookieID).Error; err == nil {
			// Update only Step 1 fields
			// GORM Updates with struct ignores zero values, which is mostly fine here.
			// However, RefCode and Status should usually NOT be changed by Step 1 Edit unless intended.
			// We construct a specific update map or struct.
			// Helper: use the 'loan' struct we built, but ensure RefCode is NOT overwritten if we don't want to.
			// actually 'loan' above has a NEW generated RefCode. We should NOT use that for update.

			// Let's rebuild the loan struct WITHOUT RefCode/Status for update,
			// OR just overwrite them in 'loan' with existing values before Update?
			// But 'loan' is a fresh struct.

			// Better Strategy: Define updates map
			updates := map[string]interface{}{
				"LastUpdateDate":      time.Now().Format("2006-01-02 15:04:05"),
				"StaffID":             staffID,
				"Title":               c.FormValue("title"),
				"FirstName":           c.FormValue("first_name"),
				"LastName":            c.FormValue("last_name"),
				"Gender":              c.FormValue("gender"),
				"BorrowerType":        c.FormValue("borrower_type"),
				"TradeRegistrationID": c.FormValue("trade_registration_id"),
				"RegistrationDate":    safeDate(c.FormValue("registration_date")),
				"TaxID":               c.FormValue("tax_id"),
				"IdCard":              c.FormValue("id_card"),
				"IdCardIssueDate":     safeDate(c.FormValue("id_card_issue_date")),
				"IdCardExpiryDate":    safeDate(c.FormValue("id_card_expiry_date")),
				"DateOfBirth":         safeDate(c.FormValue("date_of_birth")),
				"MobilePhone":         c.FormValue("mobile_phone"),
				"CompanyName": func() string {
					if c.FormValue("borrower_type") == "juristic" {
						return c.FormValue("juristic_company_name")
					}
					return c.FormValue("work_company_name")
				}(),
				// ... skipping irrelevant lines if contiguous ...
				// Actually, I cannot skip in ReplacementContent without providing the full block.
				// Let's try to target specific lines if possible or blocks.
				// Step 1 Updates Map is large.
				// I will target the specific dates in Step 1 NEW creation first, then Updates map.

				// Address - House Registration
				"HouseRegNo":          c.FormValue("house_reg_no"),
				"HouseRegBuilding":    c.FormValue("house_reg_building"),
				"HouseRegFloor":       c.FormValue("house_reg_floor"),
				"HouseRegRoom":        c.FormValue("house_reg_room"),
				"HouseRegMoo":         c.FormValue("house_reg_moo"),
				"HouseRegSoi":         c.FormValue("house_reg_soi"),
				"HouseRegRoad":        c.FormValue("house_reg_road"),
				"HouseRegProvince":    c.FormValue("house_reg_province"),
				"HouseRegDistrict":    c.FormValue("house_reg_district"),
				"HouseRegSubdistrict": c.FormValue("house_reg_subdistrict"),
				"HouseRegZipcode":     c.FormValue("house_reg_zipcode"),
				"SameAsHouseReg":      c.FormValue("same_as_house_reg") == "on",

				// Current Address
				"CurrentCompany":     c.FormValue("current_company"),
				"CurrentNo":          c.FormValue("current_no"),
				"CurrentBuilding":    c.FormValue("current_building"),
				"CurrentFloor":       c.FormValue("current_floor"),
				"CurrentRoom":        c.FormValue("current_room"),
				"CurrentMoo":         c.FormValue("current_moo"),
				"CurrentSoi":         c.FormValue("current_soi"),
				"CurrentRoad":        c.FormValue("current_road"),
				"CurrentProvince":    c.FormValue("current_province"),
				"CurrentDistrict":    c.FormValue("current_district"),
				"CurrentSubdistrict": c.FormValue("current_subdistrict"),
				"CurrentZipcode":     c.FormValue("current_zipcode"),

				// Work Address
				"WorkNo":          c.FormValue("work_no"),
				"WorkBuilding":    c.FormValue("work_building"),
				"WorkFloor":       c.FormValue("work_floor"),
				"WorkRoom":        c.FormValue("work_room"),
				"WorkMoo":         c.FormValue("work_moo"),
				"WorkSoi":         c.FormValue("work_soi"),
				"WorkRoad":        c.FormValue("work_road"),
				"WorkProvince":    c.FormValue("work_province"),
				"WorkDistrict":    c.FormValue("work_district"),
				"WorkSubdistrict": c.FormValue("work_subdistrict"),
				"WorkZipcode":     c.FormValue("work_zipcode"),
				"WorkPhone":       c.FormValue("work_phone"),

				// Doc Delivery
				"DocDeliveryType": c.FormValue("doc_delivery_type"),
				"DocNo":           c.FormValue("doc_no"),
				"DocBuilding":     c.FormValue("doc_building"),
				"DocFloor":        c.FormValue("doc_floor"),
				"DocRoom":         c.FormValue("doc_room"),
				"DocMoo":          c.FormValue("doc_moo"),
				"DocSoi":          c.FormValue("doc_soi"),
				"DocRoad":         c.FormValue("doc_road"),
				"DocProvince":     c.FormValue("doc_province"),
				"DocDistrict":     c.FormValue("doc_district"),
				"DocSubdistrict":  c.FormValue("doc_subdistrict"),
				"DocZipcode":      c.FormValue("doc_zipcode"),
				// "CompanyName":        c.FormValue("company_name"), // Removed duplicate
				"Occupation":         c.FormValue("occupation"),
				"Position":           c.FormValue("position"),
				"Salary":             salary,
				"OtherIncome":        otherIncome,
				"IncomeSource":       c.FormValue("income_source"),
				"CreditBureauStatus": c.FormValue("credit_bureau_status"),
			}

			config.DB.Model(&existingLoan).Updates(updates)

			// Redirect
			return c.Redirect("/step2")
		}
	}

	// Create New Mode
	// Generate RefCode from RefRunning
	currentYear := time.Now().Format("2006")

	// Use global max to determine next running number safely
	var maxRunning int = 0
	var lastLoan models.LoanApplication
	// Find the strict highest RefCode in the actual table
	if err := config.DB.Where("ref_code LIKE ?", currentYear+"%").Order("ref_code desc").First(&lastLoan).Error; err == nil {
		if len(lastLoan.RefCode) >= 8 {
			if lastRun, err := strconv.Atoi(lastLoan.RefCode[4:]); err == nil {
				maxRunning = lastRun
			}
		}
	}

	var refRunning models.RefRunning
	// Ensure table exists
	config.DB.AutoMigrate(&models.RefRunning{})

	// Check/Create RefRunning record for THIS staffID (User Requirement)
	if err := config.DB.Where("ref_year = ? AND emp_id = ?", currentYear, staffID).First(&refRunning).Error; err != nil {
		// Not found, create new record for this user
		refRunning = models.RefRunning{
			RefYear: currentYear,
			EmpID:   staffID,
			Running: maxRunning, // Init with current global max (will increment below)
		}
		config.DB.Create(&refRunning)
	}

	// Always increment from the GLOBAL max to ensure uniqueness
	// We ignore refRunning.Running for calculation to prevent collision if multiple users are active
	nextRunning := maxRunning + 1

	// Update the user's RefRunning record to reflect the number they are about to use
	refRunning.Running = nextRunning
	config.DB.Save(&refRunning)

	// Format: YYYY + Running (4 digits)
	refCode := fmt.Sprintf("%s%04d", currentYear, nextRunning)

	loan := models.LoanApplication{
		// New Fields
		RefCode:        refCode,
		Status:         "D",
		LastUpdateDate: time.Now().Format("2006-01-02 15:04:05"),
		StaffID:        staffID,

		// Step 1
		Title:               c.FormValue("title"),
		FirstName:           c.FormValue("first_name"),
		LastName:            c.FormValue("last_name"),
		Gender:              c.FormValue("gender"),
		BorrowerType:        c.FormValue("borrower_type"),
		TradeRegistrationID: c.FormValue("trade_registration_id"),
		RegistrationDate:    safeDate(c.FormValue("registration_date")),
		TaxID:               c.FormValue("tax_id"),
		CompanyName: func() string {
			if c.FormValue("borrower_type") == "juristic" {
				return c.FormValue("juristic_company_name")
			}
			return c.FormValue("work_company_name")
		}(),
		IdCard:              c.FormValue("id_card"),
		IdCardIssueDate:     safeDate(c.FormValue("id_card_issue_date")),
		IdCardExpiryDate:    safeDate(c.FormValue("id_card_expiry_date")),
		DateOfBirth:         safeDate(c.FormValue("date_of_birth")),
		Ethnicity:           c.FormValue("ethnicity"),
		Nationality:         c.FormValue("nationality"),
		Religion:            c.FormValue("religion"),
		MaritalStatus:       c.FormValue("marital_status"),
		MobilePhone:         c.FormValue("mobile_phone"),
		OtherCardType:       c.FormValue("other_card_type"),
		OtherCardNumber:     c.FormValue("other_card_number"),
		OtherCardIssueDate:  safeDate(c.FormValue("other_card_issue_date")),
		OtherCardExpiryDate: safeDate(c.FormValue("other_card_expiry_date")),

		// Address - House Reg
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

		// Address - Current
		CurrentCompany:     c.FormValue("current_company"),
		CurrentNo:          c.FormValue("current_no"),
		CurrentBuilding:    c.FormValue("current_building"),
		CurrentFloor:       c.FormValue("current_floor"),
		CurrentRoom:        c.FormValue("current_room"),
		CurrentMoo:         c.FormValue("current_moo"),
		CurrentSoi:         c.FormValue("current_soi"),
		CurrentRoad:        c.FormValue("current_road"),
		CurrentProvince:    c.FormValue("current_province"),
		CurrentDistrict:    c.FormValue("current_district"),
		CurrentSubdistrict: c.FormValue("current_subdistrict"),
		CurrentZipcode:     c.FormValue("current_zipcode"),

		// Address - Work
		WorkNo:          c.FormValue("work_no"),
		WorkBuilding:    c.FormValue("work_building"),
		WorkFloor:       c.FormValue("work_floor"),
		WorkRoom:        c.FormValue("work_room"),
		WorkMoo:         c.FormValue("work_moo"),
		WorkSoi:         c.FormValue("work_soi"),
		WorkRoad:        c.FormValue("work_road"),
		WorkProvince:    c.FormValue("work_province"),
		WorkDistrict:    c.FormValue("work_district"),
		WorkSubdistrict: c.FormValue("work_subdistrict"),
		WorkZipcode:     c.FormValue("work_zipcode"),
		WorkPhone:       c.FormValue("work_phone"),

		// Address - Doc Delivery
		DocDeliveryType: c.FormValue("doc_delivery_type"),
		DocNo:           c.FormValue("doc_no"),
		DocBuilding:     c.FormValue("doc_building"),
		DocFloor:        c.FormValue("doc_floor"),
		DocRoom:         c.FormValue("doc_room"),
		DocMoo:          c.FormValue("doc_moo"),
		DocSoi:          c.FormValue("doc_soi"),
		DocRoad:         c.FormValue("doc_road"),
		DocProvince:     c.FormValue("doc_province"),
		DocDistrict:     c.FormValue("doc_district"),
		DocSubdistrict:  c.FormValue("doc_subdistrict"),
		DocZipcode:      c.FormValue("doc_zipcode"),

		// CompanyName:        c.FormValue("company_name"), // Moved up
		Occupation:         c.FormValue("occupation"),
		Position:           c.FormValue("position"),
		Salary:             salary,
		OtherIncome:        otherIncome,
		IncomeSource:       c.FormValue("income_source"),
		CreditBureauStatus: c.FormValue("credit_bureau_status"),
	}

	if err := config.DB.Create(&loan).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:  "loan_id",
		Value: fmt.Sprintf("%d", loan.ID),
	})

	return c.Redirect("/step2")
}

/* ข้อมูลรถ */
func Step2(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		config.DB.First(&loan, loanID)
	}

	var carKinds []models.CarKind
	config.DB.Find(&carKinds)

	var carBrands []models.CarBrand
	config.DB.Find(&carBrands)

	return c.Render("step2", fiber.Map{
		"title":     "ขั้นตอน 2",
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

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	// Helper to parse currency (strip commas)
	parseMoney := func(v string) float64 {
		v = strings.ReplaceAll(v, ",", "")
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	carWeight := parseMoney(c.FormValue("car_weight"))
	carCC := int(parseMoney(c.FormValue("car_cc")))
	carMileage := parseMoney(c.FormValue("car_mileage"))
	vatRate := parseMoney(c.FormValue("vat_rate"))
	carPrice := parseMoney(c.FormValue("car_price"))
	isRefinance := c.FormValue("is_refinance") == "on"

	// Update fields
	loan.CarType = c.FormValue("car_type")
	loan.CarCode = c.FormValue("car_code")
	loan.CarBrand = c.FormValue("car_brand")
	loan.CarRegisterDate = safeDate(c.FormValue("car_register_date"))
	loan.CarModel = c.FormValue("car_model")
	loan.CarYear = c.FormValue("car_year")
	loan.CarColor = c.FormValue("car_color")
	loan.CarWeight = carWeight
	loan.CarCC = carCC
	loan.CarMileage = carMileage
	loan.CarChassisNo = c.FormValue("car_chassis_no")
	loan.CarGear = c.FormValue("car_gear")
	loan.CarEngineNo = c.FormValue("car_engine_no")
	loan.CarCondition = c.FormValue("car_condition")
	loan.LicensePlate = c.FormValue("license_plate")
	loan.LicenseProvince = c.FormValue("license_province")
	loan.VatRate = vatRate
	loan.CarPrice = carPrice
	loan.IsRefinance = isRefinance

	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)
	return c.Redirect("/step3")
}

/* ข้อมูลสัญญา */
func Step3(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		config.DB.First(&loan, loanID)
	}
	return c.Render("step3", fiber.Map{
		"title": "ขั้นตอน 3",
		"Loan":  loan,
	})
}

func Step3Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	// Helper to parse currency (strip commas)
	parseMoney := func(v string) float64 {
		// Remove commas
		v = strings.ReplaceAll(v, ",", "")
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	loanAmount := parseMoney(c.FormValue("loan_amount"))
	interestRate, _ := strconv.ParseFloat(c.FormValue("interest_rate"), 64)
	installments, _ := strconv.Atoi(c.FormValue("installments"))
	installmentAmount := parseMoney(c.FormValue("installment_amount"))

	downPayment := parseMoney(c.FormValue("down_payment"))
	transferFee := parseMoney(c.FormValue("transfer_fee"))
	taxFee := parseMoney(c.FormValue("tax_fee"))
	dutyFee := parseMoney(c.FormValue("duty_fee"))

	// Update fields
	loan.ContractSignDate = safeDate(c.FormValue("contract_sign_date"))
	loan.LoanType = c.FormValue("loan_type")
	loan.LoanAmount = loanAmount
	loan.InterestRate = interestRate
	loan.Installments = installments
	loan.InstallmentAmount = installmentAmount
	loan.DownPayment = downPayment
	loan.ContractStartDate = safeDate(c.FormValue("contract_start_date"))
	loan.FirstPaymentDate = safeDate(c.FormValue("first_payment_date"))
	loan.TransferType = c.FormValue("transfer_type")
	loan.TransferFee = transferFee
	loan.TaxFee = taxFee
	loan.DutyFee = dutyFee
	loan.PaymentDay, _ = strconv.Atoi(c.FormValue("payment_day"))
	loan.HasLifeInsurance = c.FormValue("is_life_insurance") == "true"
	loan.CarInsuranceBeginning = parseMoney(c.FormValue("beginning_amount"))
	loan.CarInsuranceRefinanceFee = parseMoney(c.FormValue("refinance_fee"))

	// Parse Life Insurance Details
	if loan.HasLifeInsurance {
		loan.LifeLoanPrincipal = parseMoney(c.FormValue("life_loan_principal"))
		loan.LifeInterestRate, _ = strconv.ParseFloat(c.FormValue("life_interest_rate"), 64)
		loan.LifeInstallments, _ = strconv.Atoi(c.FormValue("life_installments"))
		loan.LifeInsuranceCompany = c.FormValue("life_insurance_company")
		loan.LifeInsuranceRate, _ = strconv.ParseFloat(c.FormValue("life_premium_rate"), 64)
		loan.LifePremium = parseMoney(c.FormValue("insurance_premium"))
	} else {
		// Reset if unchecked
		loan.LifeLoanPrincipal = 0
		loan.LifeInterestRate = 0
		loan.LifeInstallments = 0
		loan.LifeInsuranceCompany = ""
		loan.LifeInsuranceRate = 0
		loan.LifePremium = 0
	}

	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)
	return c.Redirect("/step4")
}

/* ค้ำประกัน/อื่นๆ */
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
		config.DB.Preload("Guarantors", "deleted_at IS NULL").First(&loan, loanID)
	}
	return c.Render("step4", fiber.Map{
		"title": "ขั้นตอน 4",
		"Loan":  loan,
	})
}

func Step4Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	// Helper to parse currency (strip commas)
	parseMoney := func(v string) float64 {
		v = strings.ReplaceAll(v, ",", "")
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	commission := parseMoney(c.FormValue("commission"))
	scoreOfficer := parseMoney(c.FormValue("score_officer"))
	scoreManager := parseMoney(c.FormValue("score_manager"))

	// Update fields
	loan.NoGuarantor = c.FormValue("no_guarantor") == "on"
	loan.Guarantor1Name = c.FormValue("guarantor1_name")
	loan.Guarantor1Contact = c.FormValue("guarantor1_contact")
	loan.Guarantor2Name = c.FormValue("guarantor2_name")
	loan.Guarantor2Contact = c.FormValue("guarantor2_contact")
	loan.Guarantor3Name = c.FormValue("guarantor3_name")
	loan.Guarantor3Contact = c.FormValue("guarantor3_contact")
	loan.LoanOfficer = c.FormValue("loan_officer")
	loan.CompanySeller = c.FormValue("company_seller")
	loan.CompanySellerID = c.FormValue("company_seller_id") // Save New Field
	loan.ShowroomStaff = c.FormValue("showroom_staff")
	loan.Commission = commission
	loan.ScoreOfficer = scoreOfficer
	loan.ScoreManager = scoreManager

	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)
	return c.Redirect("/step5")
}

/* ประกันชีวิต */
func Step5(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		config.DB.First(&loan, loanID)
	}
	return c.Render("step5", fiber.Map{
		"title": "ขั้นตอน 5",
		"Loan":  loan,
	})
}

func Step5Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	// Helper to parse currency (strip commas)
	parseMoney := func(v string) float64 {
		v = strings.ReplaceAll(v, ",", "")
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	hasLifeInsurance := c.FormValue("is_life_insurance") == "true"
	lifeLoanPrincipal := parseMoney(c.FormValue("life_loan_amount"))
	lifeInterestRate, _ := strconv.ParseFloat(c.FormValue("life_interest_rate"), 64)
	lifeInstallments, _ := strconv.Atoi(c.FormValue("life_installments"))
	lifeInsuranceRate := parseMoney(c.FormValue("life_premium_rate"))
	lifePremium := parseMoney(c.FormValue("life_premium"))

	// Update fields
	loan.HasLifeInsurance = hasLifeInsurance
	loan.LifeInsuranceCompany = c.FormValue("life_insurance_company")
	loan.LifeLoanPrincipal = lifeLoanPrincipal
	loan.LifeInterestRate = lifeInterestRate
	loan.LifeInstallments = lifeInstallments
	loan.LifeGender = c.FormValue("life_gender")
	loan.LifeDob = safeDate(c.FormValue("life_dob"))
	loan.LifeSignDate = safeDate(c.FormValue("life_sign_date"))
	loan.LifeInsuranceRate = lifeInsuranceRate
	loan.LifePremium = lifePremium
	loan.Beneficiary1Name = c.FormValue("beneficiary1_name")
	loan.Beneficiary1Relation = c.FormValue("beneficiary1_relation")
	loan.Beneficiary1Address = c.FormValue("beneficiary1_address")
	loan.Beneficiary2Name = c.FormValue("beneficiary2_name")
	loan.Beneficiary2Relation = c.FormValue("beneficiary2_relation")
	loan.Beneficiary2Address = c.FormValue("beneficiary2_address")
	loan.Beneficiary3Name = c.FormValue("beneficiary3_name")
	loan.Beneficiary3Relation = c.FormValue("beneficiary3_relation")
	loan.Beneficiary3Address = c.FormValue("beneficiary3_address")
	loan.InsuranceSeller = c.FormValue("insurance_agent")
	loan.InsuranceAgentEmpId = c.FormValue("insurance_agent_empid")
	loan.InsuranceLicenseNo = c.FormValue("agent_license")

	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)
	return c.Redirect("/step6")
}

/* ประกันภัย */
func Step6(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		config.DB.First(&loan, loanID)
	}
	return c.Render("step6", fiber.Map{
		"title": "ขั้นตอน 6",
		"Loan":  loan,
	})
}

func Step6Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	carInsurancePremium, _ := strconv.ParseFloat(c.FormValue("insurance_cost"), 64)
	carInsuranceAvoidanceFee, _ := strconv.ParseFloat(c.FormValue("avoidance_fee"), 64)

	// Update fields
	loan.CarInsuranceType = c.FormValue("insurance_type")
	loan.CarInsuranceCompany = c.FormValue("insurance_company")
	loan.CarInsuranceClass = c.FormValue("insurance_class")
	loan.CarInsuranceNotifyDate = safeDate(c.FormValue("notification_date"))
	loan.CarInsuranceNotifyNo = c.FormValue("notification_number")
	loan.CarInsuranceStartDate = safeDate(c.FormValue("coverage_start_date"))
	loan.CarInsurancePremium = carInsurancePremium
	loan.CarInsuranceAvoidanceFee = carInsuranceAvoidanceFee
	// loan.CarInsuranceFile = c.FormValue("car_insurance_file") // Skip file for now

	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)
	return c.Redirect("/step7")
}

/* หักภาษี */
func Step7(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	var loan models.LoanApplication
	if loanID != "" {
		config.DB.First(&loan, loanID)
	}
	return c.Render("step7", fiber.Map{
		"title": "ขั้นตอน 7",
		"Loan":  loan,
	})
}

func Step7Post(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Redirect("/step1")
	}

	var loan models.LoanApplication
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		return c.Redirect("/step1")
	}

	// Update fields
	loan.TaxPayerType = c.FormValue("tax_payer_type")
	loan.TaxIdCard = c.FormValue("tax_id_card")
	loan.TaxPrefix = c.FormValue("tax_prefix")
	loan.TaxFirstName = c.FormValue("tax_first_name")
	loan.TaxLastName = c.FormValue("tax_last_name")
	loan.TaxHouseNo = c.FormValue("tax_house_no")
	loan.TaxBuilding = c.FormValue("tax_building")
	loan.TaxFloor = c.FormValue("tax_floor")
	loan.TaxRoom = c.FormValue("tax_room")
	loan.TaxVillage = c.FormValue("tax_village")
	loan.TaxMoo = c.FormValue("tax_moo")
	loan.TaxSoi = c.FormValue("tax_soi")
	loan.TaxRoad = c.FormValue("tax_road")
	loan.TaxProvince = c.FormValue("tax_province")
	loan.TaxDistrict = c.FormValue("tax_district")
	loan.TaxSubdistrict = c.FormValue("tax_sub_district")
	loan.TaxZipcode = c.FormValue("tax_zipcode")

	// Finalize Application
	loan.Status = "D"
	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")

	config.DB.Save(&loan)

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

	var loan models.LoanApplication
	if err := config.DB.Where("ref_code = ?", req.RefCode).First(&loan).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
	}

	// ถ้าไม่ส่ง status มา ใช้ P (Pending) เหมือนเดิม
	newStatus := req.Status
	if newStatus == "" {
		newStatus = "P"
	}

	loan.Status = newStatus
	loan.LastUpdateDate = time.Now().Format("2006-01-02 15:04:05")
	if newStatus == "P" {
		loan.SubmittedDate = time.Now().Format("2006-01-02 15:04:05")
	}

	config.DB.Save(&loan)

	// แจ้งเตือนเฉพาะเมื่อมีผลการอนุมัติ และส่งเฉพาะ staff เจ้าของเคส
	borrowerName := loan.FirstName + " " + loan.LastName
	if borrowerName == " " {
		borrowerName = loan.RefCode
	}
	switch newStatus {
	case "A":
		BroadcastToStaff(loan.StaffID, "✅ อนุมัติสินเชื่อแล้ว",
			fmt.Sprintf("เลขที่ %s - %s ได้รับการอนุมัติ", loan.RefCode, borrowerName))
	case "R":
		BroadcastToStaff(loan.StaffID, "❌ ไม่อนุมัติสินเชื่อ",
			fmt.Sprintf("เลขที่ %s - %s ไม่ผ่านการอนุมัติ", loan.RefCode, borrowerName))
	case "C":
		BroadcastToStaff(loan.StaffID, "⚠️ อนุมัติแบบมีเงื่อนไข",
			fmt.Sprintf("เลขที่ %s - %s อนุมัติแบบมีเงื่อนไข", loan.RefCode, borrowerName))
	}

	return c.JSON(fiber.Map{"success": true})
}

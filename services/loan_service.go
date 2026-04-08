package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"loan-app/models"
	"loan-app/repositories"

	"gorm.io/gorm"
)

type LoanService struct {
	repo LoanRepository
	now  func() time.Time
}

type LoanRepository interface {
	ListByStaff(staffID string) ([]models.LoanApplication, error)
	SaveLoan(loan *models.LoanApplication) error
	CreateLoan(loan *models.LoanApplication) error
	FindLoanByRefCode(refCode string) (*models.LoanApplication, error)
	FindLoanWithGuarantors(loanID int) (*models.LoanApplication, error)
	FindLatestRefCodeByYear(year string) (string, error)
	FindRefRunning(year, empID string) (*models.RefRunning, error)
	CreateRefRunning(refRunning *models.RefRunning) error
	SaveRefRunning(refRunning *models.RefRunning) error
}

type Step1Input struct {
	Title               string
	FirstName           string
	LastName            string
	Gender              string
	BorrowerType        string
	TradeRegistrationID string
	RegistrationDate    string
	TaxID               string
	CompanyName         string
	IdCard              string
	IdCardIssueDate     string
	IdCardExpiryDate    string
	DateOfBirth         string
	Ethnicity           string
	Nationality         string
	Religion            string
	MaritalStatus       string
	MobilePhone         string
	OtherCardType       string
	OtherCardNumber     string
	OtherCardIssueDate  string
	OtherCardExpiryDate string
	HouseRegNo          string
	HouseRegBuilding    string
	HouseRegFloor       string
	HouseRegRoom        string
	HouseRegMoo         string
	HouseRegSoi         string
	HouseRegRoad        string
	HouseRegProvince    string
	HouseRegDistrict    string
	HouseRegSubdistrict string
	HouseRegZipcode     string
	SameAsHouseReg      bool
	CurrentCompany      string
	CurrentNo           string
	CurrentBuilding     string
	CurrentFloor        string
	CurrentRoom         string
	CurrentMoo          string
	CurrentSoi          string
	CurrentRoad         string
	CurrentProvince     string
	CurrentDistrict     string
	CurrentSubdistrict  string
	CurrentZipcode      string
	WorkNo              string
	WorkBuilding        string
	WorkFloor           string
	WorkRoom            string
	WorkMoo             string
	WorkSoi             string
	WorkRoad            string
	WorkProvince        string
	WorkDistrict        string
	WorkSubdistrict     string
	WorkZipcode         string
	WorkPhone           string
	DocDeliveryType     string
	DocNo               string
	DocBuilding         string
	DocFloor            string
	DocRoom             string
	DocMoo              string
	DocSoi              string
	DocRoad             string
	DocProvince         string
	DocDistrict         string
	DocSubdistrict      string
	DocZipcode          string
	Occupation          string
	Position            string
	Salary              float64
	OtherIncome         float64
	IncomeSource        string
	CreditBureauStatus  string
}

type Step2Input struct {
	CarType         string
	CarCode         string
	CarBrand        string
	CarRegisterDate string
	CarModel        string
	CarYear         string
	CarColor        string
	CarWeight       float64
	CarCC           int
	CarMileage      float64
	CarChassisNo    string
	CarGear         string
	CarEngineNo     string
	CarCondition    string
	LicensePlate    string
	LicenseProvince string
	VatRate         float64
	CarPrice        float64
	IsRefinance     bool
}

type Step3Input struct {
	ContractSignDate         string
	LoanType                 string
	LoanAmount               float64
	InterestRate             float64
	Installments             int
	InstallmentAmount        float64
	DownPayment              float64
	ContractStartDate        string
	FirstPaymentDate         string
	TransferType             string
	TransferFee              float64
	TaxFee                   float64
	DutyFee                  float64
	PaymentDay               int
	HasLifeInsurance         bool
	CarInsuranceBeginning    float64
	CarInsuranceRefinanceFee float64
	LifeLoanPrincipal        float64
	LifeInterestRate         float64
	LifeInstallments         int
	LifeInsuranceCompany     string
	LifeInsuranceRate        float64
	LifePremium              float64
}

type Step4Input struct {
	NoGuarantor       bool
	Guarantor1Name    string
	Guarantor1Contact string
	Guarantor2Name    string
	Guarantor2Contact string
	Guarantor3Name    string
	Guarantor3Contact string
	LoanOfficer       string
	CompanySeller     string
	CompanySellerID   string
	ShowroomStaff     string
	Commission        float64
	ScoreOfficer      float64
	ScoreManager      float64
}

type Step5Input struct {
	HasLifeInsurance     bool
	LifeInsuranceCompany string
	LifeLoanPrincipal    float64
	LifeInterestRate     float64
	LifeInstallments     int
	LifeGender           string
	LifeDob              string
	LifeSignDate         string
	LifeInsuranceRate    float64
	LifePremium          float64
	Beneficiary1Name     string
	Beneficiary1Relation string
	Beneficiary1Address  string
	Beneficiary2Name     string
	Beneficiary2Relation string
	Beneficiary2Address  string
	Beneficiary3Name     string
	Beneficiary3Relation string
	Beneficiary3Address  string
	InsuranceSeller      string
	InsuranceAgentEmpID  string
	InsuranceLicenseNo   string
}

type Step6Input struct {
	CarInsuranceType         string
	CarInsuranceCompany      string
	CarInsuranceClass        string
	CarInsuranceNotifyDate   string
	CarInsuranceNotifyNo     string
	CarInsuranceStartDate    string
	CarInsurancePremium      float64
	CarInsuranceAvoidanceFee float64
}

type Step7Input struct {
	TaxPayerType   string
	TaxIdCard      string
	TaxPrefix      string
	TaxFirstName   string
	TaxLastName    string
	TaxHouseNo     string
	TaxBuilding    string
	TaxFloor       string
	TaxRoom        string
	TaxVillage     string
	TaxMoo         string
	TaxSoi         string
	TaxRoad        string
	TaxProvince    string
	TaxDistrict    string
	TaxSubdistrict string
	TaxZipcode     string
}

func NewLoanService(repo LoanRepository) *LoanService {
	return &LoanService{repo: repo, now: time.Now}
}

func NewDefaultLoanService(db *gorm.DB) *LoanService {
	return NewLoanService(repositories.NewGormLoanRepository(db))
}

func optionalDate(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func timestampString(now time.Time) string {
	return now.Format("2006-01-02 15:04:05")
}

func ParseMoney(value string) float64 {
	value = strings.ReplaceAll(value, ",", "")
	if value == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(value, 64)
	return f
}

func (s *LoanService) ListByStaff(staffID string) ([]models.LoanApplication, error) {
	if staffID == "" {
		return []models.LoanApplication{}, nil
	}
	return s.repo.ListByStaff(staffID)
}

func (s *LoanService) FindWithGuarantors(loanID int) (*models.LoanApplication, error) {
	return s.repo.FindLoanWithGuarantors(loanID)
}

func (s *LoanService) UpdateStep1(loan *models.LoanApplication, staffID string, input Step1Input) error {
	loan.LastUpdateDate = timestampString(s.now())
	loan.StaffID = staffID
	loan.Title = input.Title
	loan.FirstName = input.FirstName
	loan.LastName = input.LastName
	loan.Gender = input.Gender
	loan.BorrowerType = input.BorrowerType
	loan.TradeRegistrationID = input.TradeRegistrationID
	loan.RegistrationDate = optionalDate(input.RegistrationDate)
	loan.TaxID = input.TaxID
	loan.CompanyName = input.CompanyName
	loan.IdCard = input.IdCard
	loan.IdCardIssueDate = optionalDate(input.IdCardIssueDate)
	loan.IdCardExpiryDate = optionalDate(input.IdCardExpiryDate)
	loan.DateOfBirth = optionalDate(input.DateOfBirth)
	loan.Ethnicity = input.Ethnicity
	loan.Nationality = input.Nationality
	loan.Religion = input.Religion
	loan.MaritalStatus = input.MaritalStatus
	loan.MobilePhone = input.MobilePhone
	loan.OtherCardType = input.OtherCardType
	loan.OtherCardNumber = input.OtherCardNumber
	loan.OtherCardIssueDate = optionalDate(input.OtherCardIssueDate)
	loan.OtherCardExpiryDate = optionalDate(input.OtherCardExpiryDate)
	loan.HouseRegNo = input.HouseRegNo
	loan.HouseRegBuilding = input.HouseRegBuilding
	loan.HouseRegFloor = input.HouseRegFloor
	loan.HouseRegRoom = input.HouseRegRoom
	loan.HouseRegMoo = input.HouseRegMoo
	loan.HouseRegSoi = input.HouseRegSoi
	loan.HouseRegRoad = input.HouseRegRoad
	loan.HouseRegProvince = input.HouseRegProvince
	loan.HouseRegDistrict = input.HouseRegDistrict
	loan.HouseRegSubdistrict = input.HouseRegSubdistrict
	loan.HouseRegZipcode = input.HouseRegZipcode
	loan.SameAsHouseReg = input.SameAsHouseReg
	loan.CurrentCompany = input.CurrentCompany
	loan.CurrentNo = input.CurrentNo
	loan.CurrentBuilding = input.CurrentBuilding
	loan.CurrentFloor = input.CurrentFloor
	loan.CurrentRoom = input.CurrentRoom
	loan.CurrentMoo = input.CurrentMoo
	loan.CurrentSoi = input.CurrentSoi
	loan.CurrentRoad = input.CurrentRoad
	loan.CurrentProvince = input.CurrentProvince
	loan.CurrentDistrict = input.CurrentDistrict
	loan.CurrentSubdistrict = input.CurrentSubdistrict
	loan.CurrentZipcode = input.CurrentZipcode
	loan.WorkNo = input.WorkNo
	loan.WorkBuilding = input.WorkBuilding
	loan.WorkFloor = input.WorkFloor
	loan.WorkRoom = input.WorkRoom
	loan.WorkMoo = input.WorkMoo
	loan.WorkSoi = input.WorkSoi
	loan.WorkRoad = input.WorkRoad
	loan.WorkProvince = input.WorkProvince
	loan.WorkDistrict = input.WorkDistrict
	loan.WorkSubdistrict = input.WorkSubdistrict
	loan.WorkZipcode = input.WorkZipcode
	loan.WorkPhone = input.WorkPhone
	loan.DocDeliveryType = input.DocDeliveryType
	loan.DocNo = input.DocNo
	loan.DocBuilding = input.DocBuilding
	loan.DocFloor = input.DocFloor
	loan.DocRoom = input.DocRoom
	loan.DocMoo = input.DocMoo
	loan.DocSoi = input.DocSoi
	loan.DocRoad = input.DocRoad
	loan.DocProvince = input.DocProvince
	loan.DocDistrict = input.DocDistrict
	loan.DocSubdistrict = input.DocSubdistrict
	loan.DocZipcode = input.DocZipcode
	loan.Occupation = input.Occupation
	loan.Position = input.Position
	loan.Salary = input.Salary
	loan.OtherIncome = input.OtherIncome
	loan.IncomeSource = input.IncomeSource
	loan.CreditBureauStatus = input.CreditBureauStatus
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) CreateStep1(staffID string, input Step1Input) (*models.LoanApplication, error) {
	refCode, err := s.nextRefCode(staffID)
	if err != nil {
		return nil, err
	}

	loan := &models.LoanApplication{
		RefCode: refCode,
		Status:  "D",
	}
	if err := s.UpdateStep1(loan, staffID, input); err != nil {
		return nil, err
	}
	if err := s.repo.CreateLoan(loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *LoanService) UpdateStep2(loan *models.LoanApplication, input Step2Input) error {
	loan.CarType = input.CarType
	loan.CarCode = input.CarCode
	loan.CarBrand = input.CarBrand
	loan.CarRegisterDate = optionalDate(input.CarRegisterDate)
	loan.CarModel = input.CarModel
	loan.CarYear = input.CarYear
	loan.CarColor = input.CarColor
	loan.CarWeight = input.CarWeight
	loan.CarCC = input.CarCC
	loan.CarMileage = input.CarMileage
	loan.CarChassisNo = input.CarChassisNo
	loan.CarGear = input.CarGear
	loan.CarEngineNo = input.CarEngineNo
	loan.CarCondition = input.CarCondition
	loan.LicensePlate = input.LicensePlate
	loan.LicenseProvince = input.LicenseProvince
	loan.VatRate = input.VatRate
	loan.CarPrice = input.CarPrice
	loan.IsRefinance = input.IsRefinance
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStep3(loan *models.LoanApplication, input Step3Input) error {
	loan.ContractSignDate = optionalDate(input.ContractSignDate)
	loan.LoanType = input.LoanType
	loan.LoanAmount = input.LoanAmount
	loan.InterestRate = input.InterestRate
	loan.Installments = input.Installments
	loan.InstallmentAmount = input.InstallmentAmount
	loan.DownPayment = input.DownPayment
	loan.ContractStartDate = optionalDate(input.ContractStartDate)
	loan.FirstPaymentDate = optionalDate(input.FirstPaymentDate)
	loan.TransferType = input.TransferType
	loan.TransferFee = input.TransferFee
	loan.TaxFee = input.TaxFee
	loan.DutyFee = input.DutyFee
	loan.PaymentDay = input.PaymentDay
	loan.HasLifeInsurance = input.HasLifeInsurance
	loan.CarInsuranceBeginning = input.CarInsuranceBeginning
	loan.CarInsuranceRefinanceFee = input.CarInsuranceRefinanceFee
	if input.HasLifeInsurance {
		loan.LifeLoanPrincipal = input.LifeLoanPrincipal
		loan.LifeInterestRate = input.LifeInterestRate
		loan.LifeInstallments = input.LifeInstallments
		loan.LifeInsuranceCompany = input.LifeInsuranceCompany
		loan.LifeInsuranceRate = input.LifeInsuranceRate
		loan.LifePremium = input.LifePremium
	} else {
		loan.LifeLoanPrincipal = 0
		loan.LifeInterestRate = 0
		loan.LifeInstallments = 0
		loan.LifeInsuranceCompany = ""
		loan.LifeInsuranceRate = 0
		loan.LifePremium = 0
	}
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStep4(loan *models.LoanApplication, input Step4Input) error {
	loan.NoGuarantor = input.NoGuarantor
	loan.Guarantor1Name = input.Guarantor1Name
	loan.Guarantor1Contact = input.Guarantor1Contact
	loan.Guarantor2Name = input.Guarantor2Name
	loan.Guarantor2Contact = input.Guarantor2Contact
	loan.Guarantor3Name = input.Guarantor3Name
	loan.Guarantor3Contact = input.Guarantor3Contact
	loan.LoanOfficer = input.LoanOfficer
	loan.CompanySeller = input.CompanySeller
	loan.CompanySellerID = input.CompanySellerID
	loan.ShowroomStaff = input.ShowroomStaff
	loan.Commission = input.Commission
	loan.ScoreOfficer = input.ScoreOfficer
	loan.ScoreManager = input.ScoreManager
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStep5(loan *models.LoanApplication, input Step5Input) error {
	loan.HasLifeInsurance = input.HasLifeInsurance
	loan.LifeInsuranceCompany = input.LifeInsuranceCompany
	loan.LifeLoanPrincipal = input.LifeLoanPrincipal
	loan.LifeInterestRate = input.LifeInterestRate
	loan.LifeInstallments = input.LifeInstallments
	loan.LifeGender = input.LifeGender
	loan.LifeDob = optionalDate(input.LifeDob)
	loan.LifeSignDate = optionalDate(input.LifeSignDate)
	loan.LifeInsuranceRate = input.LifeInsuranceRate
	loan.LifePremium = input.LifePremium
	loan.Beneficiary1Name = input.Beneficiary1Name
	loan.Beneficiary1Relation = input.Beneficiary1Relation
	loan.Beneficiary1Address = input.Beneficiary1Address
	loan.Beneficiary2Name = input.Beneficiary2Name
	loan.Beneficiary2Relation = input.Beneficiary2Relation
	loan.Beneficiary2Address = input.Beneficiary2Address
	loan.Beneficiary3Name = input.Beneficiary3Name
	loan.Beneficiary3Relation = input.Beneficiary3Relation
	loan.Beneficiary3Address = input.Beneficiary3Address
	loan.InsuranceSeller = input.InsuranceSeller
	loan.InsuranceAgentEmpId = input.InsuranceAgentEmpID
	loan.InsuranceLicenseNo = input.InsuranceLicenseNo
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStep6(loan *models.LoanApplication, input Step6Input) error {
	loan.CarInsuranceType = input.CarInsuranceType
	loan.CarInsuranceCompany = input.CarInsuranceCompany
	loan.CarInsuranceClass = input.CarInsuranceClass
	loan.CarInsuranceNotifyDate = optionalDate(input.CarInsuranceNotifyDate)
	loan.CarInsuranceNotifyNo = input.CarInsuranceNotifyNo
	loan.CarInsuranceStartDate = optionalDate(input.CarInsuranceStartDate)
	loan.CarInsurancePremium = input.CarInsurancePremium
	loan.CarInsuranceAvoidanceFee = input.CarInsuranceAvoidanceFee
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStep7(loan *models.LoanApplication, input Step7Input) error {
	loan.TaxPayerType = input.TaxPayerType
	loan.TaxIdCard = input.TaxIdCard
	loan.TaxPrefix = input.TaxPrefix
	loan.TaxFirstName = input.TaxFirstName
	loan.TaxLastName = input.TaxLastName
	loan.TaxHouseNo = input.TaxHouseNo
	loan.TaxBuilding = input.TaxBuilding
	loan.TaxFloor = input.TaxFloor
	loan.TaxRoom = input.TaxRoom
	loan.TaxVillage = input.TaxVillage
	loan.TaxMoo = input.TaxMoo
	loan.TaxSoi = input.TaxSoi
	loan.TaxRoad = input.TaxRoad
	loan.TaxProvince = input.TaxProvince
	loan.TaxDistrict = input.TaxDistrict
	loan.TaxSubdistrict = input.TaxSubdistrict
	loan.TaxZipcode = input.TaxZipcode
	loan.Status = "D"
	loan.LastUpdateDate = timestampString(s.now())
	return s.repo.SaveLoan(loan)
}

func (s *LoanService) UpdateStatus(refCode, status string) (*models.LoanApplication, error) {
	loan, err := s.repo.FindLoanByRefCode(refCode)
	if err != nil {
		return nil, err
	}
	if status == "" {
		status = "P"
	}
	loan.Status = status
	loan.LastUpdateDate = timestampString(s.now())
	if status == "P" {
		loan.SubmittedDate = timestampString(s.now())
	}
	if err := s.repo.SaveLoan(loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *LoanService) nextRefCode(staffID string) (string, error) {
	currentYear := s.now().Format("2006")
	maxRunning := 0
	if latestRefCode, err := s.repo.FindLatestRefCodeByYear(currentYear); err == nil && len(latestRefCode) >= 8 {
		if lastRun, convErr := strconv.Atoi(latestRefCode[4:]); convErr == nil {
			maxRunning = lastRun
		}
	}

	refRunning, err := s.repo.FindRefRunning(currentYear, staffID)
	if err != nil {
		refRunning = &models.RefRunning{
			RefYear: currentYear,
			EmpID:   staffID,
			Running: maxRunning,
		}
		if createErr := s.repo.CreateRefRunning(refRunning); createErr != nil {
			return "", createErr
		}
	}

	nextRunning := maxRunning + 1
	refRunning.Running = nextRunning
	if err := s.repo.SaveRefRunning(refRunning); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", currentYear, nextRunning), nil
}

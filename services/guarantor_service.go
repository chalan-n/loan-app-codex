package services

import (
	"time"

	"loan-app/models"

	"gorm.io/gorm"
)

type GuarantorInput struct {
	GuarantorType       string
	TradeRegistrationID string
	RegistrationDate    string
	TaxID               string
	Title               string
	FirstName           string
	LastName            string
	Gender              string
	IdCard              string
	IdCardIssueDate     string
	IdCardExpiryDate    string
	DateOfBirth         string
	Ethnicity           string
	Nationality         string
	Religion            string
	MaritalStatus       string
	MobilePhone         string
	HouseRegNo          string
	HouseRegMoo         string
	HouseRegSoi         string
	HouseRegRoad        string
	HouseRegProvince    string
	HouseRegDistrict    string
	HouseRegSubdistrict string
	HouseRegZipcode     string
	SameAsHouseReg      bool
	CurrentNo           string
	CurrentMoo          string
	CurrentSoi          string
	CurrentRoad         string
	CurrentProvince     string
	CurrentDistrict     string
	CurrentSubdistrict  string
	CurrentZipcode      string
	CompanyName         string
	Occupation          string
	Position            string
	Salary              float64
	OtherIncome         float64
	IncomeSource        string
	WorkPhone           string
	WorkNo              string
	WorkMoo             string
	WorkSoi             string
	WorkRoad            string
	WorkProvince        string
	WorkDistrict        string
	WorkSubdistrict     string
	WorkZipcode         string
	OtherCardType       string
	OtherCardNumber     string
	OtherCardIssueDate  string
	OtherCardExpiryDate string
	DocDeliveryType     string
	DocNo               string
	DocMoo              string
	DocSoi              string
	DocRoad             string
	DocProvince         string
	DocDistrict         string
	DocSubdistrict      string
	DocZipcode          string
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func ApplyGuarantorInput(guarantor *models.Guarantor, input GuarantorInput) {
	guarantor.GuarantorType = input.GuarantorType
	guarantor.TradeRegistrationID = input.TradeRegistrationID
	guarantor.RegistrationDate = stringPtrOrNil(input.RegistrationDate)
	guarantor.TaxID = input.TaxID
	guarantor.Title = input.Title
	guarantor.FirstName = input.FirstName
	guarantor.LastName = input.LastName
	guarantor.Gender = input.Gender
	guarantor.IdCard = input.IdCard
	guarantor.IdCardIssueDate = stringPtrOrNil(input.IdCardIssueDate)
	guarantor.IdCardExpiryDate = stringPtrOrNil(input.IdCardExpiryDate)
	guarantor.DateOfBirth = stringPtrOrNil(input.DateOfBirth)
	guarantor.Ethnicity = input.Ethnicity
	guarantor.Nationality = input.Nationality
	guarantor.Religion = input.Religion
	guarantor.MaritalStatus = input.MaritalStatus
	guarantor.MobilePhone = input.MobilePhone
	guarantor.HouseRegNo = input.HouseRegNo
	guarantor.HouseRegMoo = input.HouseRegMoo
	guarantor.HouseRegSoi = input.HouseRegSoi
	guarantor.HouseRegRoad = input.HouseRegRoad
	guarantor.HouseRegProvince = input.HouseRegProvince
	guarantor.HouseRegDistrict = input.HouseRegDistrict
	guarantor.HouseRegSubdistrict = input.HouseRegSubdistrict
	guarantor.HouseRegZipcode = input.HouseRegZipcode
	guarantor.SameAsHouseReg = input.SameAsHouseReg
	guarantor.CurrentNo = input.CurrentNo
	guarantor.CurrentMoo = input.CurrentMoo
	guarantor.CurrentSoi = input.CurrentSoi
	guarantor.CurrentRoad = input.CurrentRoad
	guarantor.CurrentProvince = input.CurrentProvince
	guarantor.CurrentDistrict = input.CurrentDistrict
	guarantor.CurrentSubdistrict = input.CurrentSubdistrict
	guarantor.CurrentZipcode = input.CurrentZipcode
	guarantor.CompanyName = input.CompanyName
	guarantor.Occupation = input.Occupation
	guarantor.Position = input.Position
	guarantor.Salary = input.Salary
	guarantor.OtherIncome = input.OtherIncome
	guarantor.IncomeSource = input.IncomeSource
	guarantor.WorkPhone = input.WorkPhone
	guarantor.WorkNo = input.WorkNo
	guarantor.WorkMoo = input.WorkMoo
	guarantor.WorkSoi = input.WorkSoi
	guarantor.WorkRoad = input.WorkRoad
	guarantor.WorkProvince = input.WorkProvince
	guarantor.WorkDistrict = input.WorkDistrict
	guarantor.WorkSubdistrict = input.WorkSubdistrict
	guarantor.WorkZipcode = input.WorkZipcode
	guarantor.OtherCardType = input.OtherCardType
	guarantor.OtherCardNumber = input.OtherCardNumber
	guarantor.OtherCardIssueDate = stringPtrOrNil(input.OtherCardIssueDate)
	guarantor.OtherCardExpiryDate = stringPtrOrNil(input.OtherCardExpiryDate)
	guarantor.DocDeliveryType = input.DocDeliveryType
	guarantor.DocNo = input.DocNo
	guarantor.DocMoo = input.DocMoo
	guarantor.DocSoi = input.DocSoi
	guarantor.DocRoad = input.DocRoad
	guarantor.DocProvince = input.DocProvince
	guarantor.DocDistrict = input.DocDistrict
	guarantor.DocSubdistrict = input.DocSubdistrict
	guarantor.DocZipcode = input.DocZipcode
}

func FindGuarantorByLoan(db *gorm.DB, loanID int, guarantorID string) (*models.Guarantor, error) {
	var guarantor models.Guarantor
	if err := db.Where("id = ? AND loan_id = ?", guarantorID, loanID).First(&guarantor).Error; err != nil {
		return nil, err
	}
	return &guarantor, nil
}

func SaveGuarantor(db *gorm.DB, loanID int, guarantorID string, input GuarantorInput) error {
	if guarantorID != "" {
		guarantor, err := FindGuarantorByLoan(db, loanID, guarantorID)
		if err != nil {
			return err
		}
		ApplyGuarantorInput(guarantor, input)
		return db.Save(guarantor).Error
	}

	guarantor := models.Guarantor{
		LoanID: int32(loanID),
	}
	ApplyGuarantorInput(&guarantor, input)
	if err := db.Create(&guarantor).Error; err != nil {
		return err
	}

	return db.Model(&models.LoanApplication{}).
		Where("id = ?", loanID).
		Updates(map[string]interface{}{
			"no_guarantor":     false,
			"last_update_date": gorm.Expr("NOW()"),
		}).Error
}

func DeleteGuarantor(db *gorm.DB, loanID int, guarantorID string) error {
	return db.Where("id = ? AND loan_id = ?", guarantorID, loanID).Delete(&models.Guarantor{}).Error
}

func TouchLoanGuarantorUpdate(db *gorm.DB, loanID int) error {
	return db.Model(&models.LoanApplication{}).
		Where("id = ?", loanID).
		Update("last_update_date", time.Now().Format("2006-01-02 15:04:05")).Error
}

package repositories

import (
	"loan-app/models"

	"gorm.io/gorm"
)

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

type GormLoanRepository struct {
	db *gorm.DB
}

func NewGormLoanRepository(db *gorm.DB) *GormLoanRepository {
	return &GormLoanRepository{db: db}
}

func (r *GormLoanRepository) ListByStaff(staffID string) ([]models.LoanApplication, error) {
	var loans []models.LoanApplication
	if err := r.db.Where("staff_id = ?", staffID).Order("id desc").Find(&loans).Error; err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *GormLoanRepository) SaveLoan(loan *models.LoanApplication) error {
	return r.db.Save(loan).Error
}

func (r *GormLoanRepository) CreateLoan(loan *models.LoanApplication) error {
	return r.db.Create(loan).Error
}

func (r *GormLoanRepository) FindLoanByRefCode(refCode string) (*models.LoanApplication, error) {
	var loan models.LoanApplication
	if err := r.db.Where("ref_code = ?", refCode).First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *GormLoanRepository) FindLoanWithGuarantors(loanID int) (*models.LoanApplication, error) {
	var loan models.LoanApplication
	if err := r.db.Preload("Guarantors", "deleted_at IS NULL").First(&loan, loanID).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *GormLoanRepository) FindLatestRefCodeByYear(year string) (string, error) {
	var loan models.LoanApplication
	if err := r.db.Where("ref_code LIKE ?", year+"%").Order("ref_code desc").First(&loan).Error; err != nil {
		return "", err
	}
	return loan.RefCode, nil
}

func (r *GormLoanRepository) FindRefRunning(year, empID string) (*models.RefRunning, error) {
	var refRunning models.RefRunning
	if err := r.db.Where("ref_year = ? AND emp_id = ?", year, empID).First(&refRunning).Error; err != nil {
		return nil, err
	}
	return &refRunning, nil
}

func (r *GormLoanRepository) CreateRefRunning(refRunning *models.RefRunning) error {
	return r.db.Create(refRunning).Error
}

func (r *GormLoanRepository) SaveRefRunning(refRunning *models.RefRunning) error {
	return r.db.Save(refRunning).Error
}

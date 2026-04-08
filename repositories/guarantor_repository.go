package repositories

import (
	"loan-app/models"

	"gorm.io/gorm"
)

type GuarantorRepository interface {
	FindByLoan(loanID int, guarantorID string) (*models.Guarantor, error)
	Save(guarantor *models.Guarantor) error
	Create(guarantor *models.Guarantor) error
	MarkLoanHasGuarantor(loanID int) error
	DeleteByLoan(loanID int, guarantorID string) error
}

type GormGuarantorRepository struct {
	db *gorm.DB
}

func NewGormGuarantorRepository(db *gorm.DB) *GormGuarantorRepository {
	return &GormGuarantorRepository{db: db}
}

func (r *GormGuarantorRepository) FindByLoan(loanID int, guarantorID string) (*models.Guarantor, error) {
	var guarantor models.Guarantor
	if err := r.db.Where("id = ? AND loan_id = ?", guarantorID, loanID).First(&guarantor).Error; err != nil {
		return nil, err
	}
	return &guarantor, nil
}

func (r *GormGuarantorRepository) Save(guarantor *models.Guarantor) error {
	return r.db.Save(guarantor).Error
}

func (r *GormGuarantorRepository) Create(guarantor *models.Guarantor) error {
	return r.db.Create(guarantor).Error
}

func (r *GormGuarantorRepository) MarkLoanHasGuarantor(loanID int) error {
	return r.db.Model(&models.LoanApplication{}).
		Where("id = ?", loanID).
		Updates(map[string]interface{}{
			"no_guarantor":     false,
			"last_update_date": gorm.Expr("NOW()"),
		}).Error
}

func (r *GormGuarantorRepository) DeleteByLoan(loanID int, guarantorID string) error {
	return r.db.Where("id = ? AND loan_id = ?", guarantorID, loanID).Delete(&models.Guarantor{}).Error
}

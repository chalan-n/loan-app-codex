package repositories

import (
	"loan-app/models"

	"gorm.io/gorm"
)

type LoanFileRepository interface {
	FindByStorageKey(storageKey string) (*models.LoanFile, error)
	Create(file *models.LoanFile) error
	DeleteByLoanAndStorageKey(loanID int, storageKey string) error
}

type GormLoanFileRepository struct {
	db *gorm.DB
}

func NewGormLoanFileRepository(db *gorm.DB) *GormLoanFileRepository {
	return &GormLoanFileRepository{db: db}
}

func (r *GormLoanFileRepository) FindByStorageKey(storageKey string) (*models.LoanFile, error) {
	var file models.LoanFile
	if err := r.db.Where("storage_key = ?", storageKey).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormLoanFileRepository) Create(file *models.LoanFile) error {
	return r.db.Create(file).Error
}

func (r *GormLoanFileRepository) DeleteByLoanAndStorageKey(loanID int, storageKey string) error {
	return r.db.Where("loan_id = ? AND storage_key = ?", loanID, storageKey).Delete(&models.LoanFile{}).Error
}

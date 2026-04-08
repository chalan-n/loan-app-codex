package models

import "time"

const (
	LoanFileCategoryCarInsurance = "car_insurance"
)

type LoanFile struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	LoanID       int       `gorm:"index;not null" json:"loan_id"`
	StorageKey   string    `gorm:"size:255;uniqueIndex;not null" json:"storage_key"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	Category     string    `gorm:"size:50;index" json:"category"`
	UploadedBy   string    `gorm:"size:100;index" json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func (LoanFile) TableName() string {
	return "loan_application_files"
}

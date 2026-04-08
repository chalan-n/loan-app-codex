package services

import (
	"errors"
	"loan-app/models"
	"loan-app/repositories"
)

var ErrLoanFileMetadataNotFound = errors.New("loan file metadata not found")

type LoanFileService struct {
	repo repositories.LoanFileRepository
}

func NewLoanFileService(repo repositories.LoanFileRepository) *LoanFileService {
	return &LoanFileService{repo: repo}
}

func (s *LoanFileService) FindByStorageKey(storageKey string) (*models.LoanFile, error) {
	file, err := s.repo.FindByStorageKey(storageKey)
	if err != nil {
		return nil, ErrLoanFileMetadataNotFound
	}
	return file, nil
}

func (s *LoanFileService) Create(loanID int, storageKey, originalName, category, uploadedBy string) error {
	return s.repo.Create(&models.LoanFile{
		LoanID:       loanID,
		StorageKey:   storageKey,
		OriginalName: originalName,
		Category:     category,
		UploadedBy:   uploadedBy,
	})
}

func (s *LoanFileService) Delete(loanID int, storageKey string) error {
	return s.repo.DeleteByLoanAndStorageKey(loanID, storageKey)
}

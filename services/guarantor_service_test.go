package services

import (
	"errors"
	"testing"

	"loan-app/models"
)

type guarantorRepoStub struct {
	findResult       *models.Guarantor
	findErr          error
	savedGuarantor   *models.Guarantor
	createdGuarantor *models.Guarantor
	deletedLoanID    int
	deletedID        string
	markedLoanID     int
}

func (s *guarantorRepoStub) FindByLoan(loanID int, guarantorID string) (*models.Guarantor, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.findResult, nil
}

func (s *guarantorRepoStub) Save(guarantor *models.Guarantor) error {
	copyGuarantor := *guarantor
	s.savedGuarantor = &copyGuarantor
	return nil
}

func (s *guarantorRepoStub) Create(guarantor *models.Guarantor) error {
	copyGuarantor := *guarantor
	s.createdGuarantor = &copyGuarantor
	return nil
}

func (s *guarantorRepoStub) MarkLoanHasGuarantor(loanID int) error {
	s.markedLoanID = loanID
	return nil
}

func (s *guarantorRepoStub) DeleteByLoan(loanID int, guarantorID string) error {
	s.deletedLoanID = loanID
	s.deletedID = guarantorID
	return nil
}

func TestGuarantorServiceSaveUpdatesExistingGuarantor(t *testing.T) {
	repo := &guarantorRepoStub{
		findResult: &models.Guarantor{ID: 5, LoanID: 12, FirstName: "Old"},
	}
	service := NewGuarantorService(repo)

	err := service.Save(12, "5", GuarantorInput{
		FirstName:        "Jane",
		LastName:         "Doe",
		RegistrationDate: "",
	})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if repo.savedGuarantor == nil {
		t.Fatalf("Save() did not persist updated guarantor")
	}
	if repo.savedGuarantor.FirstName != "Jane" || repo.savedGuarantor.LastName != "Doe" {
		t.Fatalf("saved guarantor = %+v, want updated names", repo.savedGuarantor)
	}
	if repo.markedLoanID != 0 {
		t.Fatalf("MarkLoanHasGuarantor called for update, got %d", repo.markedLoanID)
	}
}

func TestGuarantorServiceSaveCreatesGuarantorAndMarksLoan(t *testing.T) {
	repo := &guarantorRepoStub{}
	service := NewGuarantorService(repo)

	err := service.Save(12, "", GuarantorInput{
		FirstName:   "Jane",
		LastName:    "Doe",
		Salary:      12000,
		CompanyName: "ACME",
	})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if repo.createdGuarantor == nil {
		t.Fatalf("Save() did not create guarantor")
	}
	if repo.createdGuarantor.LoanID != 12 {
		t.Fatalf("created loan_id = %d, want %d", repo.createdGuarantor.LoanID, 12)
	}
	if repo.createdGuarantor.CompanyName != "ACME" {
		t.Fatalf("created guarantor company = %q, want %q", repo.createdGuarantor.CompanyName, "ACME")
	}
	if repo.markedLoanID != 12 {
		t.Fatalf("marked loan id = %d, want %d", repo.markedLoanID, 12)
	}
}

func TestGuarantorServiceSaveReturnsFindError(t *testing.T) {
	service := NewGuarantorService(&guarantorRepoStub{findErr: errors.New("not found")})

	if err := service.Save(12, "4", GuarantorInput{}); err == nil {
		t.Fatalf("Save() error = nil, want non-nil")
	}
}

func TestGuarantorServiceDeleteUsesRepository(t *testing.T) {
	repo := &guarantorRepoStub{}
	service := NewGuarantorService(repo)

	if err := service.Delete(12, "4"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if repo.deletedLoanID != 12 || repo.deletedID != "4" {
		t.Fatalf("DeleteByLoan called with (%d, %q), want (12, %q)", repo.deletedLoanID, repo.deletedID, "4")
	}
}

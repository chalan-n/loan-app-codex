package services

import (
	"errors"
	"testing"
	"time"

	"loan-app/models"
)

type loanRepoStub struct {
	listResult        []models.LoanApplication
	savedLoan         *models.LoanApplication
	createdLoan       *models.LoanApplication
	loanByRef         *models.LoanApplication
	loanWithGuarantor *models.LoanApplication
	latestRefCode     string
	refRunning        *models.RefRunning
	findRefErr        error
}

func (s *loanRepoStub) ListByStaff(staffID string) ([]models.LoanApplication, error) {
	return s.listResult, nil
}

func (s *loanRepoStub) SaveLoan(loan *models.LoanApplication) error {
	copyLoan := *loan
	s.savedLoan = &copyLoan
	return nil
}

func (s *loanRepoStub) CreateLoan(loan *models.LoanApplication) error {
	copyLoan := *loan
	s.createdLoan = &copyLoan
	return nil
}

func (s *loanRepoStub) FindLoanByRefCode(refCode string) (*models.LoanApplication, error) {
	if s.loanByRef == nil {
		return nil, errors.New("not found")
	}
	return s.loanByRef, nil
}

func (s *loanRepoStub) FindLoanWithGuarantors(loanID int) (*models.LoanApplication, error) {
	if s.loanWithGuarantor == nil {
		return nil, errors.New("not found")
	}
	return s.loanWithGuarantor, nil
}

func (s *loanRepoStub) FindLatestRefCodeByYear(year string) (string, error) {
	if s.latestRefCode == "" {
		return "", errors.New("not found")
	}
	return s.latestRefCode, nil
}

func (s *loanRepoStub) FindRefRunning(year, empID string) (*models.RefRunning, error) {
	if s.findRefErr != nil {
		return nil, s.findRefErr
	}
	if s.refRunning == nil {
		return nil, errors.New("not found")
	}
	return s.refRunning, nil
}

func (s *loanRepoStub) CreateRefRunning(refRunning *models.RefRunning) error {
	copyRefRunning := *refRunning
	s.refRunning = &copyRefRunning
	return nil
}

func (s *loanRepoStub) SaveRefRunning(refRunning *models.RefRunning) error {
	copyRefRunning := *refRunning
	s.refRunning = &copyRefRunning
	return nil
}

func TestLoanServiceCreateStep1CreatesDraftWithRefCode(t *testing.T) {
	repo := &loanRepoStub{latestRefCode: "20260012", findRefErr: errors.New("not found")}
	service := NewLoanService(repo)
	service.now = func() time.Time { return time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC) }

	loan, err := service.CreateStep1("570639", Step1Input{
		FirstName: "Jane",
		LastName:  "Doe",
		IdCard:    "123",
	})
	if err != nil {
		t.Fatalf("CreateStep1() error: %v", err)
	}
	if loan.RefCode != "20260013" {
		t.Fatalf("ref_code = %q, want %q", loan.RefCode, "20260013")
	}
	if loan.Status != models.LoanStatusDraft {
		t.Fatalf("status = %q, want %q", loan.Status, models.LoanStatusDraft)
	}
	if repo.createdLoan == nil || repo.createdLoan.StaffID != "570639" {
		t.Fatalf("created loan = %+v, want saved draft for staff", repo.createdLoan)
	}
}

func TestLoanServiceUpdateStep3ClearsLifeInsuranceWhenDisabled(t *testing.T) {
	repo := &loanRepoStub{}
	service := NewLoanService(repo)
	service.now = func() time.Time { return time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC) }
	loan := &models.LoanApplication{
		LifeLoanPrincipal:    100,
		LifeInterestRate:     1.5,
		LifeInstallments:     12,
		LifeInsuranceCompany: "Old",
		LifeInsuranceRate:    2.5,
		LifePremium:          300,
	}

	err := service.UpdateStep3(loan, Step3Input{
		HasLifeInsurance: false,
		LoanAmount:       5000,
	})
	if err != nil {
		t.Fatalf("UpdateStep3() error: %v", err)
	}
	if repo.savedLoan == nil {
		t.Fatal("UpdateStep3() did not save loan")
	}
	if repo.savedLoan.LifeLoanPrincipal != 0 || repo.savedLoan.LifeInsuranceCompany != "" || repo.savedLoan.LifePremium != 0 {
		t.Fatalf("life insurance fields not cleared: %+v", repo.savedLoan)
	}
}

func TestLoanServiceUpdateStatusDefaultsToPending(t *testing.T) {
	repo := &loanRepoStub{
		loanByRef: &models.LoanApplication{RefCode: "20260001", StaffID: "570639"},
	}
	service := NewLoanService(repo)
	service.now = func() time.Time { return time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC) }

	loan, err := service.UpdateStatus("20260001", "")
	if err != nil {
		t.Fatalf("UpdateStatus() error: %v", err)
	}
	if loan.Status != models.LoanStatusPending {
		t.Fatalf("status = %q, want %q", loan.Status, models.LoanStatusPending)
	}
	if loan.SubmittedDate == "" {
		t.Fatal("submitted date should be set for pending status")
	}
}

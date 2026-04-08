package services

import (
	"testing"

	"loan-app/models"
)

func TestCanAccessLoan(t *testing.T) {
	loan := &models.LoanApplication{StaffID: "owner"}

	if !CanAccessLoan(models.RoleOfficer, "owner", loan) {
		t.Fatal("expected owner to access own loan")
	}
	if !CanAccessLoan(models.RoleAdmin, "admin", loan) {
		t.Fatal("expected admin to access other user's loan")
	}
	if CanAccessLoan(models.RoleOfficer, "other", loan) {
		t.Fatal("expected non-owner officer to be denied")
	}
}

func TestLoanIDFromFilename(t *testing.T) {
	if loanID, ok := LoanIDFromFilename("12_file.pdf"); !ok || loanID != 12 {
		t.Fatalf("LoanIDFromFilename() = %d, %v; want 12, true", loanID, ok)
	}
	if _, ok := LoanIDFromFilename("badfile.pdf"); ok {
		t.Fatal("expected malformed filename to fail")
	}
}

func TestLoanHasFile(t *testing.T) {
	loan := &models.LoanApplication{CarInsuranceFile: "12_a.pdf,12_b.pdf"}
	if !LoanHasFile(loan, "12_b.pdf") {
		t.Fatal("expected file to be linked to loan")
	}
	if LoanHasFile(loan, "12_c.pdf") {
		t.Fatal("expected missing file to be rejected")
	}
}

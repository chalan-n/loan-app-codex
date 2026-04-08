package models

import "testing"

func TestNormalizeRole(t *testing.T) {
	if got := NormalizeRole(RoleAdmin); got != RoleAdmin {
		t.Fatalf("NormalizeRole(admin) = %q, want %q", got, RoleAdmin)
	}
	if got := NormalizeRole("unknown"); got != RoleOfficer {
		t.Fatalf("NormalizeRole(unknown) = %q, want %q", got, RoleOfficer)
	}
	if !IsManagerOrAbove(RoleManager) || !IsManagerOrAbove(RoleAdmin) {
		t.Fatal("expected manager/admin to be manager-or-above")
	}
	if IsManagerOrAbove(RoleOfficer) {
		t.Fatal("expected officer to not be manager-or-above")
	}
}

func TestNormalizeLoanStatus(t *testing.T) {
	if got := NormalizeLoanStatus(LoanStatusApproved); got != LoanStatusApproved {
		t.Fatalf("NormalizeLoanStatus(approved) = %q, want %q", got, LoanStatusApproved)
	}
	if got := NormalizeLoanStatus(""); got != LoanStatusPending {
		t.Fatalf("NormalizeLoanStatus(empty) = %q, want %q", got, LoanStatusPending)
	}
	if got := NormalizeLoanStatus("X"); got != LoanStatusPending {
		t.Fatalf("NormalizeLoanStatus(invalid) = %q, want %q", got, LoanStatusPending)
	}
}

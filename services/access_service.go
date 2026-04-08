package services

import (
	"loan-app/models"
	"strconv"
	"strings"
)

// CanAccessLoan centralizes loan ownership and role-based access rules.
func CanAccessLoan(role, username string, loan *models.LoanApplication) bool {
	if loan == nil {
		return false
	}

	return role == models.RoleAdmin || role == models.RoleManager || loan.StaffID == username
}

// LoanIDFromFilename extracts the loan ID prefix from stored filenames like "<loanID>_...".
func LoanIDFromFilename(filename string) (int, bool) {
	prefix, _, found := strings.Cut(filename, "_")
	if !found || prefix == "" {
		return 0, false
	}

	loanID, err := strconv.Atoi(prefix)
	if err != nil || loanID <= 0 {
		return 0, false
	}

	return loanID, true
}

// LoanHasFile checks whether a filename is linked to the given loan record.
func LoanHasFile(loan *models.LoanApplication, filename string) bool {
	if loan == nil {
		return false
	}

	for _, file := range strings.Split(loan.CarInsuranceFile, ",") {
		if strings.TrimSpace(file) == filename {
			return true
		}
	}
	return false
}

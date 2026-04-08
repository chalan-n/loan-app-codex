package models

const (
	LoanStatusDraft       = "D"
	LoanStatusPending     = "P"
	LoanStatusApproved    = "A"
	LoanStatusRejected    = "R"
	LoanStatusConditional = "C"
)

var validLoanStatuses = map[string]struct{}{
	LoanStatusDraft:       {},
	LoanStatusPending:     {},
	LoanStatusApproved:    {},
	LoanStatusRejected:    {},
	LoanStatusConditional: {},
}

func IsValidLoanStatus(status string) bool {
	_, ok := validLoanStatuses[status]
	return ok
}

func NormalizeLoanStatus(status string) string {
	if status == "" {
		return LoanStatusPending
	}
	if !IsValidLoanStatus(status) {
		return LoanStatusPending
	}
	return status
}

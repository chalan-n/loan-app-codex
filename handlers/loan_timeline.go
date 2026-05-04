package handlers

import (
	"loan-app/config"
	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

// LoanTimelinePage shows audit activity tied to a loan for admins only.
func LoanTimelinePage(c *fiber.Ctx) error {
	var loan models.LoanApplication
	if config.DB == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Database unavailable")
	}
	if err := config.DB.First(&loan, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Loan not found")
	}

	var logs []models.AuditLog
	if loan.RefCode != "" {
		config.DB.Where("ref_code = ?", loan.RefCode).
			Order("created_at desc").
			Limit(100).
			Find(&logs)
	}

	return c.Render("loan_timeline", fiber.Map{
		"Loan":        &loan,
		"Logs":        logs,
		"CurrentRole": GetCurrentUserRole(c),
	})
}

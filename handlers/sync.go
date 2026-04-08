package handlers

import (
	"loan-app/config"
	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

// UpdateSyncStatus handles the request to update sync status to PENDING
func UpdateSyncStatus(c *fiber.Ctx) error {
	type Request struct {
		RefCode string `json:"ref_code"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.RefCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "RefCode is required",
		})
	}

	// 1. Find Loan Application
	var loan models.LoanApplication
	if err := config.DB.Where("ref_code = ?", req.RefCode).First(&loan).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Loan application not found",
		})
	}

	// 2. Update Loan Application Status
	// SyncStatus -> PENDING
	// Status -> P
	if err := config.DB.Model(&loan).Updates(map[string]interface{}{
		"sync_status": "PENDING",
		"status":      models.LoanStatusPending,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update loan status",
		})
	}

	// 3. Update Guarantors Status
	// SyncStatus -> PENDING
	if err := config.DB.Model(&models.Guarantor{}).Where("loan_id = ?", loan.ID).Update("sync_status", "PENDING").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update guarantor status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Sync status updated successfully",
	})
}

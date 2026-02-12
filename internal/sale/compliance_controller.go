package sale

import (
	"fmt"
	"pos-fiber-app/internal/subscription"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ExportTaxReportHandler returns a CSV file for tax reporting
// @Summary Export tax report as CSV
// @Tags Compliance
// @Security BearerAuth
// @Param start_date query string false "StartDate (YYYY-MM-DD)"
// @Param end_date query string false "EndDate (YYYY-MM-DD)"
// @Success 200 {file} file "tax_report.csv"
// @Router /compliance/tax-report [get]
func ExportTaxReportHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)

		// Module Check (Redundant if guarded by route group, but safe)
		if !subscription.HasModule(db, bizID, subscription.ModuleCompliance) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Compliance module not active"})
		}

		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		if startDate == "" || endDate == "" {
			// default to current month
			now := time.Now()
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
			endDate = now.Format("2006-01-02")
		}

		csvBytes, err := GenerateTaxReport(db, bizID, startDate, endDate)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate report: "+err.Error())
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=tax_report_%s.csv", startDate))
		return c.Send(csvBytes)
	}
}

// AuditTrailHandler returns detailed activity logs for audit purposes
// @Summary Get audit trail
// @Tags Compliance
// @Security BearerAuth
// @Param date query string false "Date (YYYY-MM-DD)"
// @Param action_type query string false "Filter by action type"
// @Success 200 {array} SaleActivityLogWithUser
// @Router /compliance/audit-trail [get]
func AuditTrailHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)

		// Module Check
		if !subscription.HasModule(db, bizID, subscription.ModuleCompliance) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Compliance module not active"})
		}

		date := c.Query("date")
		actionType := c.Query("action_type") // optional filtering

		logs, err := GetAuditTrail(db, bizID, date, actionType)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch audit trail: "+err.Error())
		}

		// Enhance with stats for the day if requested?
		// For now just return logs.

		return c.JSON(logs)
	}
}

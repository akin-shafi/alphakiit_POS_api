// internal/report/internal_controller.go
package report

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type InternalReportController struct {
	db      *gorm.DB
	service *ReportService
}

func NewInternalReportController(db *gorm.DB, service *ReportService) *InternalReportController {
	return &InternalReportController{
		db:      db,
		service: service,
	}
}

// AuthMiddleware for internal scheduler calls
func (c *InternalReportController) AuthMiddleware(ctx *fiber.Ctx) error {
	token := ctx.Get("X-Internal-Token")
	secret := os.Getenv("INTERNAL_API_KEY")

	if secret == "" || token != secret {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized internal request",
		})
	}
	return ctx.Next()
}

func (c *InternalReportController) DailyReportHandler(ctx *fiber.Ctx) error {
	log.Println("Internal API: Triggering all enabled daily reports")
	// Using the RunScheduledReports logic from scheduler.go
	RunScheduledReports(c.db, c.service)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Daily reports triggered successfully",
	})
}

func (c *InternalReportController) WeeklyAuditHandler(ctx *fiber.Ctx) error {
	log.Println("Internal API: Triggering weekly audit reminders")
	if err := c.service.TriggerWeeklyAuditReminders(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Audit reminders triggered successfully",
	})
}

func (c *InternalReportController) MonthlyReportHandler(ctx *fiber.Ctx) error {
	log.Println("Internal API: Triggering monthly financial reports")
	if err := c.service.TriggerMonthlyFinancialReports(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Monthly reports triggered successfully",
	})
}

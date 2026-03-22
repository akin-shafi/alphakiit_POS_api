package reconciliation

import (
	// "pos-fiber-app/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(api fiber.Router, db *gorm.DB) {
	ctrl := NewReconciliationController(db)

	reconciliation := api.Group("/reconciliation")
	
	// 1. Public Webhooks (No Auth)
	reconciliation.Post("/webhooks/:provider", ctrl.HandleWebhook)
	
	// 2. Status check (Public, but uses internal reference)
	reconciliation.Get("/status/:reference", ctrl.GetStatus)

	// 3. Admin Tools (Must be inside business-scoped /api/v1/:id/ group or protected by Auth)
	// Usually RegisterRoutes is called from a main router that already has /api/v1/:id
	// But let's check if we can just define them here.
}

func RegisterAdminRoutes(r fiber.Router, db *gorm.DB) {
	ctrl := NewReconciliationController(db)
	
	admin := r.Group("/reconciliation")
	admin.Get("/summary", ctrl.GetSummary)
	admin.Get("/payments", ctrl.ListPayments)
	admin.Get("/logs", ctrl.ListLogs)
	admin.Get("/settlement", ctrl.GetSettlement)
	admin.Post("/manual-verify", ctrl.ManuallyVerify)
}

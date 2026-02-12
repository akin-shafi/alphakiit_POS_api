package subscription

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(router fiber.Router, db *gorm.DB) {
	ac := NewAdminController(db)

	adminGroup := router.Group("/admin")

	// Subscriptions
	adminGroup.Get("/subscriptions", ac.GetAllSubscriptions)

	// Modules
	adminGroup.Get("/modules", ac.GetAllModules)
	adminGroup.Post("/modules", ac.CreateModule)
	adminGroup.Put("/modules/:id", ac.UpdateModule)
	adminGroup.Delete("/modules/:id", ac.DeleteModule)

	// Promo Codes
	adminGroup.Get("/promo-codes", ac.GetAllPromoCodes)
	adminGroup.Post("/promo-codes", ac.CreatePromoCode)
	adminGroup.Put("/promo-codes/:id", ac.UpdatePromoCode)
	adminGroup.Delete("/promo-codes/:id", ac.DeletePromoCode)

	// Referral Commissions
	adminGroup.Get("/commissions", AdminListAllCommissionsHandler(db))
	adminGroup.Patch("/commissions/:id/status", AdminUpdateCommissionStatusHandler(db))
	adminGroup.Get("/commissions/settings", GetCommissionSettingsHandler(db))
	adminGroup.Put("/commissions/settings", UpdateCommissionSettingsHandler(db))
	// Training Resources
	adminGroup.Get("/training-resources", AdminListTrainingResourcesHandler(db))
	adminGroup.Post("/training-resources", AdminCreateTrainingResourceHandler(db))
	adminGroup.Put("/training-resources/:id", AdminUpdateTrainingResourceHandler(db))
	adminGroup.Delete("/training-resources/:id", AdminDeleteTrainingResourceHandler(db))
}

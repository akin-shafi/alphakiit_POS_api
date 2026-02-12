package subscription

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB) {
	sc := NewSubscriptionController(db)

	router.Get("/plans", sc.GetPlans)
	router.Get("/pricing", sc.GetPricing)
	router.Get("/status", sc.GetStatus)
	router.Get("/promo/validate", sc.ValidatePromoCode)
	router.Post("/subscribe", sc.Subscribe)
}

func RegisterReferralRoutes(router fiber.Router, db *gorm.DB) {
	router.Post("/codes", CreateReferralCodeHandler(db))
	router.Get("/codes", GetMyReferralCodesHandler(db))
	router.Get("/commissions", GetMyCommissionsHandler(db))
	router.Get("/training-resources", GetTrainingResourcesHandler(db))
	router.Post("/payouts", RequestPayoutHandler(db))
	router.Get("/payouts", GetPayoutRequestsHandler(db))
}

func RegisterPublicRoutes(router fiber.Router, db *gorm.DB) {
	router.Get("/referrals/settings", GetCommissionSettingsHandler(db))
	sc := NewSubscriptionController(db)
	router.Get("/pricing", sc.GetPricing)
}

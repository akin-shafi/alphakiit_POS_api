package subscription

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB) {
	sc := NewSubscriptionController(db)

	subGroup := router.Group("/subscription")

	subGroup.Get("/plans", sc.GetPlans)
	subGroup.Get("/status", sc.GetStatus)
	subGroup.Post("/subscribe", sc.Subscribe)
}

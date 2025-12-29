package onboarding

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers onboarding-related routes
func RegisterRoutes(router fiber.Router, db *gorm.DB) {
	group := router.Group("/onboarding")

	group.Post("/register", RegisterHandler(db))
}

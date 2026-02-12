// internal/shift/register.go
package shift

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterShiftRoutes registers all shift-related endpoints
func RegisterShiftRoutes(r fiber.Router, db *gorm.DB) {
	// Initialize services
	shiftService := NewShiftService(db)
	shiftController := NewShiftController(shiftService)

	// Register routes
	shiftController.RegisterRoutes(r)
}

// internal/middleware/shift_guard.go
package middleware

import (
	"pos-fiber-app/internal/shift"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ShiftGuard middleware ensures the user has an active shift before processing sales
func ShiftGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract user and business info from context (set by JWT middleware)
		businessID, ok := c.Locals("business_id").(uint)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "business context not found")
		}

		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "user context not found")
		}

		// Check if user has an active shift
		shiftService := shift.NewShiftService(db)
		activeShift, err := shiftService.ValidateActiveShift(businessID, userID)

		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "You must start a shift before processing sales")
		}

		// Store shift ID in context for use in handlers
		c.Locals("shift_id", activeShift.ID)

		return c.Next()
	}
}

// OptionalShiftGuard checks for active shift but doesn't block if not found
// Useful for endpoints that benefit from shift tracking but don't require it
func OptionalShiftGuard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		businessID, ok := c.Locals("business_id").(uint)
		if !ok {
			return c.Next()
		}

		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			return c.Next()
		}

		// Try to get active shift
		shiftService := shift.NewShiftService(db)
		activeShift, err := shiftService.GetActiveShift(businessID, userID)

		if err == nil && activeShift != nil {
			// Store shift ID if found
			c.Locals("shift_id", activeShift.ID)
		}

		return c.Next()
	}
}

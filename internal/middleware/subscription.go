package middleware

import (
	"log"
	"pos-fiber-app/internal/subscription"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SubscriptionMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip subscription check for subscription management routes
		path := c.Path()
		if strings.Contains(path, "/subscription") {
			return c.Next()
		}

		businessID := c.Locals("business_id").(uint)
		if businessID == 0 {
			return c.Next() // Should be handled by Tenant/Business middleware
		}

		active, status, err := subscription.CheckSubscriptionAccess(db, businessID)
		if err != nil {
			log.Printf("Subscription check error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error during subscription check"})
		}

		if !active {
			return c.Status(403).JSON(fiber.Map{
				"error":   "Subscription required",
				"status":  status,
				"message": "Your subscription has expired or is inactive. Please renew to continue.",
			})
		}

		return c.Next()
	}
}

// SubscriptionGuard is a more strict middleware that only allows specific roles if needed
// For now, we use the general one

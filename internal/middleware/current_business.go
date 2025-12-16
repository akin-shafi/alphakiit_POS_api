// internal/middleware/current_business.go
package middleware

import (
	"pos-fiber-app/internal/types"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func CurrentBusinessMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)

		businessIDStr := c.Get("X-Current-Business-ID")
		if businessIDStr == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing X-Current-Business-ID header"})
		}

		businessID, err := strconv.ParseUint(businessIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid business ID"})
		}

		// Optional: verify user has access to this business (query user_business_access table if needed)
		// For now, we trust the frontend has validated from /businesses list

		c.Locals("current_business_id", uint(businessID))
		c.Locals("tenant_id", claims.TenantID) // keep for backward compat

		return c.Next()
	}
}

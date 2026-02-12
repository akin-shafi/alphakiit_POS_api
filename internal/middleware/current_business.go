// internal/middleware/current_business.go
package middleware

import (
	"fmt"
	"pos-fiber-app/internal/types"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CurrentBusinessMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)

		if claims != nil {
			fmt.Printf("[DEBUG] CurrentBusinessMiddleware - User: %s, Role: %s\n", claims.UserName, claims.Role)
		}

		// Bypass for super_admin and installers
		role := strings.TrimSpace(claims.Role)
		if strings.EqualFold(role, "super_admin") || strings.EqualFold(role, "installer") {
			fmt.Printf("[DEBUG] CurrentBusinessMiddleware Bypassing for %s\n", role)
			return c.Next()
		}

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

		c.Locals("business_id", uint(businessID))
		c.Locals("current_business_id", uint(businessID))
		c.Locals("tenant_id", claims.TenantID)

		return c.Next()
	}
}

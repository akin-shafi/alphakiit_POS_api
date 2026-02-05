// package middleware

// import "github.com/gofiber/fiber/v2"

// func TenantMiddleware() fiber.Handler {
//     return func(c *fiber.Ctx) error {
//         tenantID := c.Get("X-Tenant-ID")
//         if tenantID == "" {
//             return c.Status(400).JSON(fiber.Map{"error": "Missing X-Tenant-ID"})
//         }
//         c.Locals("tenant_id", tenantID)
//         return c.Next()
//     }
// }

package middleware

import (
	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
)

func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims, ok := c.Locals("user").(*types.UserClaims)
		if !ok {
			return fiber.ErrUnauthorized
		}

		tenantID := userClaims.TenantID

		// If header is provided, verify it matches the token (extra security)
		headerTenant := c.Get("X-Tenant-ID")
		if headerTenant != "" && headerTenant != tenantID {
			return fiber.NewError(fiber.StatusForbidden, "tenant mismatch")
		}

		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}

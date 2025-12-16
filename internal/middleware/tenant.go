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
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return fiber.ErrUnauthorized
		}

		claims := token.Claims.(jwt.MapClaims)
		tenantID := claims["tenant_id"].(string)

		headerTenant := c.Get("X-Tenant-ID")
		if headerTenant == "" || headerTenant != tenantID {
			return fiber.NewError(fiber.StatusForbidden, "tenant mismatch")
		}

		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}

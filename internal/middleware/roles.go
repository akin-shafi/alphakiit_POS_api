package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		role := claims["role"].(string)
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return fiber.ErrForbidden
	}
}

// middleware/jwt.go
package middleware

import (
	"os"

	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		ContextKey: "user", // this will store jwt.MapClaims by default
		SuccessHandler: func(c *fiber.Ctx) error {
			// The default jwt middleware puts jwt.MapClaims in c.Locals("user")
			rawClaims := c.Locals("user").(jwt.MapClaims)

			// Convert to our structured claims
			userClaims := &types.UserClaims{
				UserID:   uint(rawClaims["user_id"].(float64)),
				TenantID: rawClaims["tenant_id"].(string),
				Role:     rawClaims["role"].(string),
			}

			// Handle optional OutletID
			if outletIDFloat, ok := rawClaims["outlet_id"].(float64); ok {
				uid := uint(outletIDFloat)
				userClaims.OutletID = &uid
			}

			// Override the locals with our clean struct for easier use downstream
			c.Locals("user", userClaims)

			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "jwt file says: Unauthorized",
			})
		},
	})
}

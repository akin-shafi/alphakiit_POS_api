// middleware/jwt.go
package middleware

import (
	"log"
	"os"

	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

func JWTProtected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		ContextKey: "user", // this will store jwt.MapClaims by default
		SuccessHandler: func(c *fiber.Ctx) error {
			// In gofiber/jwt v3, the token is stored as *jwt.Token in the local context
			token := c.Locals("user").(*jwt.Token)
			rawClaims := token.Claims.(jwt.MapClaims)

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
			if err != nil {
				// Log the error for debugging
				log.Printf("JWT Error: %v", err)
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "jwt file says: Unauthorized",
			})
		},
	})
}

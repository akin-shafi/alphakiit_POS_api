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

			// Safely extract claims
			var userID uint
			if uid, ok := rawClaims["user_id"].(float64); ok {
				userID = uint(uid)
			}

			tenantID, _ := rawClaims["tenant_id"].(string) // Safe: returns "" if missing/wrong type
			role, _ := rawClaims["role"].(string)

			userClaims := &types.UserClaims{
				UserID:   userID,
				TenantID: tenantID,
				Role:     role,
			}

			if name, ok := rawClaims["user_name"].(string); ok {
				userClaims.UserName = name
			} else {
				userClaims.UserName = "User"
			}

			// Handle optional OutletID
			if outletIDFloat, ok := rawClaims["outlet_id"].(float64); ok {
				uid := uint(outletIDFloat)
				userClaims.OutletID = &uid
			}

			// Override the locals with our clean struct for easier use downstream
			c.Locals("user", userClaims)
			c.Locals("user_id", userClaims.UserID)
			c.Locals("user_name", userClaims.UserName)
			c.Locals("role", userClaims.Role)

			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			errMsg := "Unauthorized"
			if err != nil {
				log.Printf("JWT Error: %v", err)
				errMsg = err.Error()
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": errMsg,
			})
		},
	})
}

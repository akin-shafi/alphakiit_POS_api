package onboarding

import (
	"log"
	"pos-fiber-app/internal/auth"
	"pos-fiber-app/internal/validation"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterHandler godoc
// @Summary      Complete business onboarding
// @Description  Creates a tenant, business, and owner user in a single atomic operation
// @Tags         Onboarding
// @Accept       json
// @Produce      json
// @Param        payload body OnboardingRequest true "Onboarding payload"
// @Success      201 {object} map[string]interface{} "Business onboarded successfully"
// @Failure      400 {object} map[string]string "Invalid request body"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /onboarding/register [post]
func RegisterHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload OnboardingRequest

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"errors": map[string][]string{
					"body": {"Invalid JSON payload"},
				},
			})
		}

		if err := payload.Validate(); err != nil {
			return c.Status(422).JSON(fiber.Map{
				"errors": validation.FormatValidationErrors(err),
			})
		}

		business, user, err := OnboardBusiness(db, &payload)
		if err != nil {
			if err.Error() == "email already exists" {
				return c.Status(409).JSON(fiber.Map{
					"error":   "Conflict",
					"message": "This email is already registered",
				})
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// Generate tokens for auto-login
		claims := auth.Claims{
			UserID:   user.ID,
			TenantID: user.TenantID,
			Role:     user.Role,
			OutletID: user.OutletID,
		}

		accessToken, err := auth.GenerateAccessToken(claims)
		if err != nil {
			log.Printf("Failed to generate access token: %v", err)
			// Return success but let frontend handle re-login
			return c.Status(201).JSON(fiber.Map{
				"message": "Business onboarded successfully. Please login.",
			})
		}

		refreshToken, err := auth.GenerateRefreshToken(claims)
		if err != nil {
			log.Printf("Failed to generate refresh token: %v", err)
			return c.Status(201).JSON(fiber.Map{
				"message": "Business onboarded successfully. Please login.",
			})
		}

		// Save refresh token
		db.Create(&auth.RefreshToken{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})

		return c.Status(201).JSON(fiber.Map{
			"message":       "Business onboarded successfully",
			"user":          user,
			"business":      business,
			"tenant":        map[string]string{"id": business.TenantID}, // Mimic tenant object
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
}

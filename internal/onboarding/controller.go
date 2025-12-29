package onboarding

import (
	"pos-fiber-app/internal/validation"

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
// @Success      201 {object} map[string]string "Business onboarded successfully"
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

		_, _, err := OnboardBusiness(db, &payload)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Business onboarded successfully",
		})
	}
}

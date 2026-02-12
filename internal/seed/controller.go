package seed

import (
	"pos-fiber-app/internal/common"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SeedHandler seeds sample products and categories for a business
// @Summary Seed business data
// @Description Seeds sample products and categories based on business type
// @Tags Seed
// @Accept json
// @Produce json
// @Param X-Current-Business-ID header int true "Current Business ID"
// @Param business_type query string false "Business Type (RETAIL, RESTAURANT, etc.)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /seed [post]
func SeedHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID, ok := c.Locals("current_business_id").(uint)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "missing current_business_id",
			})
		}

		bizTypeStr := c.Query("business_type")
		bizType := common.BusinessType(bizTypeStr)

		if err := SeedSampleData(db, bizID, bizType); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "sample data seeded successfully",
		})
	}
}

// SeedInstallerHandler seeds global installer and commission data
// @Summary Seed installer data
// @Description Seeds global commission settings, training resources, and mock installer commissions
// @Tags Seed
// @Accept json
// @Produce json
// @Param X-Current-Business-ID header int false "Current Business ID (Optional for this endpoint)"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /seed/installers [post]
func SeedInstallerHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := SeedInstallerData(db); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "installer and commission seed data created successfully",
		})
	}
}

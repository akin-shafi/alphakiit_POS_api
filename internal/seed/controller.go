package seed

import (
	"pos-fiber-app/internal/common"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

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

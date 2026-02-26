package expense

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		var req CreateExpenseRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		exp, err := Create(db, bizID, req)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.Status(fiber.StatusCreated).JSON(exp)
	}
}

func ListHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		from := c.Query("from")
		to := c.Query("to")

		expenses, err := List(db, bizID, from, to)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(expenses)
	}
}

func DeleteHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		id, _ := c.ParamsInt("id")

		if err := Delete(db, uint(id), bizID); err != nil {
			return fiber.ErrInternalServerError
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

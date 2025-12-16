package terminal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterTerminal godoc
// @Summary Register POS terminal
// @Tags Terminals
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} Terminal
// @Router /terminals/register [post]
func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		terminal := Terminal{TenantID: tenantID, Code: uuid.NewString(), Active: true}
		if err := db.Create(&terminal).Error; err != nil {
			return fiber.ErrInternalServerError
		}
		return c.Status(201).JSON(terminal)
	}
}

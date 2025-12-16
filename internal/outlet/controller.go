package outlet

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateOutlet godoc
// @Summary Create outlet
// @Tags Outlets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateOutletRequest true "Outlet payload"
// @Success 201 {object} Outlet
// @Failure 400 {object} common.ErrorResponse
// @Router /outlets [post]
func Create(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		var body CreateOutletRequest
		if err := c.BodyParser(&body); err != nil || body.Name == "" {
			return fiber.ErrBadRequest
		}
		outlet := Outlet{TenantID: tenantID, Name: body.Name, Address: body.Address}
		if err := db.Create(&outlet).Error; err != nil {
			return fiber.ErrInternalServerError
		}
		return c.Status(201).JSON(outlet)
	}
}

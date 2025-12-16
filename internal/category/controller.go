// internal/category/controller.go
package category

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListCategories godoc
// @Summary List all categories for current business
// @Tags Category
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Category
// @Router /categories [get]
func ListHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		cats, err := ListByBusiness(db, bizID)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(cats)
	}
}

// CreateCategory godoc
// @Summary Create a new category
// @Tags Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body object{name=string,description=string} true "Category info"
// @Success 201 {object} Category
// @Router /categories [post]
func CreateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Name        string `json:"name" validate:"required"`
			Description string `json:"description"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(400, "invalid payload")
		}
		bizID := c.Locals("current_business_id").(uint)
		cat, err := Create(db, bizID, req.Name, req.Description)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.Status(201).JSON(cat)
	}
}

// GetCategory godoc
// @Summary Get a specific category
// @Tags Category
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Category ID"
// @Success 200 {object} Category
// @Router /categories/{id} [get]
func GetHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.NewError(400, "invalid category ID")
		}
		bizID := c.Locals("current_business_id").(uint)
		cat, err := Get(db, uint(id), bizID)
		if err != nil {
			return fiber.ErrNotFound
		}
		return c.JSON(cat)
	}
}

// UpdateCategory godoc
// @Summary Update a specific category
// @Tags Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Category ID"
// @Param body body object{name=string,description=string} true "Updated category info"
// @Success 200 {object} Category
// @Router /categories/{id} [put]
func UpdateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.NewError(400, "invalid category ID")
		}
		var req struct {
			Name        string `json:"name" validate:"required"`
			Description string `json:"description"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(400, "invalid payload")
		}
		bizID := c.Locals("current_business_id").(uint)
		cat, err := Update(db, uint(id), bizID, req.Name, req.Description)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(cat)
	}
}

// DeleteCategory godoc
// @Summary Delete a specific category
// @Tags Category
// @Security BearerAuth
// @Param id path uint true "Category ID"
// @Success 204 "No Content"
// @Router /categories/{id} [delete]
func DeleteHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.NewError(400, "invalid category ID")
		}
		bizID := c.Locals("current_business_id").(uint)
		if err := Delete(db, uint(id), bizID); err != nil {
			return fiber.ErrInternalServerError
		}
		return c.SendStatus(204)
	}
}

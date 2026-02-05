// internal/product/controller.go
package product

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListProducts godoc
// @Summary List products for the current business
// @Description Retrieve a list of products belonging to the selected business. Optionally filter by category.
// @Tags Product
// @Security BearerAuth
// @Produce json
// @Param category_id query uint false "Filter by Category ID"
// @Param active query bool false "Filter by active status (default: true)"
// @Success 200 {array} Product
// @Failure 500 {object} map[string]string
// @Router /products [get]
func ListHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)

		var filters []func(*gorm.DB) *gorm.DB

		if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
			var categoryID uint
			fmt.Sscanf(categoryIDStr, "%d", &categoryID)
			if categoryID > 0 {
				filters = append(filters, WithCategory(categoryID))
			}
		}

		if activeStr := c.Query("active"); activeStr != "" {
			active := activeStr != "false"
			filters = append(filters, func(q *gorm.DB) *gorm.DB {
				return q.Where("products.active = ?", active)
			})
		}

		products, err := ListByBusiness(db, bizID, filters...)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(products)
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Add a new product to the current business
// @Tags Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateProductRequest true "Product information"
// @Success 201 {object} Product
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 500 {object} map[string]string
// @Router /products [post]
func CreateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateProductRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		bizID := c.Locals("current_business_id").(uint)

		product, err := Create(db, bizID, req)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.Status(fiber.StatusCreated).JSON(product)
	}
}

// GetProduct godoc
// @Summary Get a specific product
// @Description Retrieve details of a single product by ID
// @Tags Product
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Product ID"
// @Success 200 {object} Product
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string
// @Router /products/{id} [get]
func GetHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid product ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		product, err := Get(db, uint(id), bizID)
		if err != nil {
			if err.Error() == "product not found" {
				return fiber.NewError(fiber.StatusNotFound, "product not found")
			}
			return fiber.ErrInternalServerError
		}

		return c.JSON(product)
	}
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update fields of an existing product (partial updates allowed)
// @Tags Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Product ID"
// @Param body body UpdateProductRequest true "Fields to update"
// @Success 200 {object} Product
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/{id} [put]
func UpdateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid product ID")
		}

		var req UpdateProductRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		bizID := c.Locals("current_business_id").(uint)

		product, err := Update(db, uint(id), bizID, req)
		if err != nil {
			if err.Error() == "product not found" {
				return fiber.NewError(fiber.StatusNotFound, "product not found")
			}
			return fiber.ErrInternalServerError
		}

		return c.JSON(product)
	}
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Soft or hard delete a product (recommended: soft delete via Active = false)
// @Tags Product
// @Security BearerAuth
// @Param id path uint true "Product ID"
// @Success 204 "Product deleted successfully"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/{id} [delete]
func DeleteHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid product ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		if err := Delete(db, uint(id), bizID); err != nil {
			if err.Error() == "product not found" {
				return fiber.NewError(fiber.StatusNotFound, "product not found")
			}
			return fiber.ErrInternalServerError
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

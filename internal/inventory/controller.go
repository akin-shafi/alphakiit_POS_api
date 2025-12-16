package inventory

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RestockProduct godoc
// @Summary Restock or adjust product inventory
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param product_id path uint true "Product ID"
// @Param body body object{quantity=int} true "Stock adjustment (positive = add, negative = deduct)"
// @Success 200 {object} Inventory
// @Router /products/{product_id}/stock [post]
func RestockHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID, _ := c.ParamsInt("product_id")
		var req struct {
			Quantity int `json:"quantity"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(400, "invalid quantity")
		}
		bizID := c.Locals("current_business_id").(uint)

		if err := AdjustStock(db, uint(productID), bizID, req.Quantity); err != nil {
			return fiber.ErrInternalServerError
		}

		inv, _ := GetStock(db, uint(productID), bizID)
		return c.JSON(inv)
	}
}

// LowStockHandler godoc
// @Summary List products with low stock
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Inventory
// @Router /inventory/low-stock [get]
func LowStockHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		threshold := 10 // Default threshold
		lowStockItems, err := ListLowStockItems(db, bizID, threshold)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(lowStockItems)
	}
}

// AllInventoryHandler godoc
// @Summary List all inventory items for current business
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Inventory
// @Router /inventory [get]
func AllInventoryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		inventoryItems, err := ListInventoryByBusiness(db, bizID)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(inventoryItems)
	}
}



// GetProductStock godoc
// @Summary Get product stock level
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param product_id path uint true "Product ID"
// @Success 200 {object} Inventory
// @Router /products/{product_id}/stock [get]
func GetStockHandler(db *gorm.DB) fiber.Handler { 
	return func(c *fiber.Ctx) error {
		productID, _ := c.ParamsInt("product_id")
		bizID := c.Locals("current_business_id").(uint)
		inv, err := GetStock(db, uint(productID), bizID)
		if err != nil {
			return fiber.ErrNotFound
		}
		return c.JSON(inv)
	}
}


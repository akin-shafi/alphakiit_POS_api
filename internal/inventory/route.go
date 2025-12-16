// internal/inventory/route.go
package inventory

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterInventoryRoutes(r fiber.Router, db *gorm.DB) {
	r.Post("/products/:product_id/stock", RestockHandler(db))
	r.Get("/inventory/low-stock", LowStockHandler(db))
	r.Get("/inventory", AllInventoryHandler(db))
}

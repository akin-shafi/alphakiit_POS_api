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

	// Bulk Stock Rounds
	r.Post("/inventory/rounds", StartRoundHandler(db))
	r.Post("/inventory/rounds/:id/close", CloseRoundHandler(db))
	r.Get("/inventory/rounds/active", GetAllActiveRoundsHandler(db))
	r.Get("/inventory/rounds/active/:product_id", GetActiveRoundHandler(db))
}

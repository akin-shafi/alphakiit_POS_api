// internal/table/register.go
package table

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterTableRoutes registers all table-related endpoints
func RegisterTableRoutes(r fiber.Router, db *gorm.DB) {
	// Initialize services
	tableService := NewTableService(db)
	tableController := NewTableController(tableService)

	// Register routes
	tableController.RegisterRoutes(r)
}

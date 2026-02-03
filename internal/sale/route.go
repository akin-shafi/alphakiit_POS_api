// internal/sale/route.go
package sale

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterSaleRoutes registers all sales-related endpoints under the business-scoped group
func RegisterSaleRoutes(r fiber.Router, db *gorm.DB) {
	// Draft & Cart Management
	r.Post("/sales", CreateSaleHandler(db))                           // One-shot sale
	r.Post("/sales/draft", CreateDraftHandler(db))                    // Start new sale
	r.Post("/sales/:sale_id/items", AddItemHandler(db))               // Add item to sale
	r.Delete("/sales/:sale_id/items/:item_id", RemoveItemHandler(db)) // Optional: remove specific item

	// Sale Actions
	r.Post("/sales/:sale_id/complete", CompleteSaleHandler(db)) // Finalize payment
	r.Post("/sales/:sale_id/hold", HoldSaleHandler(db))         // Park sale
	r.Post("/sales/:sale_id/void", VoidSaleHandler(db))         // Void completed sale

	// Retrieval & Reporting
	r.Get("/sales", ListSalesHandler(db))                 // List with filters
	r.Get("/sales/held", ListHeldSalesHandler(db))        // List parked sales
	r.Get("/sales/:sale_id", GetSaleHandler(db))          // Get sale + items
	r.Get("/sales/reports/daily", DailyReportHandler(db)) // Daily summary

	r.Get("/sales/reports/range", SalesReportHandler(db)) // Custom date range report

	r.Delete("/sales/:sale_id/items/:item_id", RemoveItemHandler(db)) // Optional: remove specific item
}

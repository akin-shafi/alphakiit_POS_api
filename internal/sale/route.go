// internal/sale/route.go
package sale

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterSaleRoutes registers all sales-related endpoints under the business-scoped group
func RegisterSaleRoutes(r fiber.Router, db *gorm.DB) {
	// Draft & Cart Management (Original)
	r.Post("/sales", CreateSaleHandler(db))                           // One-shot sale
	r.Post("/sales/draft", CreateDraftHandler(db))                    // Start new sale (basic)
	r.Post("/sales/:sale_id/items", AddItemHandler(db))               // Add item to sale (basic)
	r.Delete("/sales/:sale_id/items/:item_id", RemoveItemHandler(db)) // Remove specific item

	// NEW: Enhanced Draft Management with Reservations
	r.Post("/sales/draft/new", CreateDraftWithTableHandler(db))                // Create draft with table
	r.Post("/sales/:sale_id/items/reserve", AddItemWithReservationHandler(db)) // Add item with reservation
	r.Post("/sales/:sale_id/resume", ResumeDraftHandler(db))                   // Resume draft order
	r.Delete("/sales/:sale_id/draft", DeleteDraftHandler(db))                  // Delete draft order
	r.Get("/sales/drafts", ListDraftsHandler(db))                              // List all drafts

	// Sale Actions (Original)
	r.Post("/sales/:sale_id/complete", CompleteSaleHandler(db)) // Finalize payment (basic)
	r.Post("/sales/:sale_id/hold", HoldSaleHandler(db))         // Park sale
	r.Post("/sales/:sale_id/void", VoidSaleHandler(db))         // Void completed sale (basic)

	// NEW: Enhanced Sale Actions with Reservations
	r.Post("/sales/:sale_id/complete/reserve", CompleteSaleWithReservationHandler(db)) // Complete with reservation release
	r.Post("/sales/:sale_id/void/reserve", VoidSaleWithReservationHandler(db))         // Void with reservation handling

	// NEW: Bill Management
	r.Post("/sales/:sale_id/transfer", TransferBillHandler(db)) // Transfer bill to another table
	r.Post("/sales/:sale_id/merge", MergeBillsHandler(db))      // Merge multiple bills

	// NEW: Activity Logs
	r.Get("/sales/:sale_id/history", GetSaleHistoryHandler(db)) // Get sale activity history

	// Retrieval & Reporting (Original)
	r.Get("/sales", ListSalesHandler(db))                 // List with filters
	r.Get("/sales/held", ListHeldSalesHandler(db))        // List parked sales
	r.Get("/sales/:sale_id", GetSaleHandler(db))          // Get sale + items
	r.Get("/sales/reports/daily", DailyReportHandler(db)) // Daily summary
	r.Get("/sales/reports/range", SalesReportHandler(db)) // Custom date range report
}

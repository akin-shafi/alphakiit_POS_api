// internal/sale/route.go
package sale

import (
	"pos-fiber-app/internal/middleware"
	"pos-fiber-app/internal/subscription"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

// RegisterManagementRoutes registers sale-related management endpoints
// that might need to be outside the business-scoped group
// (e.g. to break import cycles or handle cross-business logic)
func RegisterManagementRoutes(r fiber.Router, db *gorm.DB) {
	r.Post("/businesses/:id/purge", PurgeHandler(db))
}

// RegisterSaleRoutes registers all sales-related endpoints under the business-scoped group
func RegisterSaleRoutes(r fiber.Router, db *gorm.DB) {
	// Draft & Cart Management (Original)
	r.Post("/sales", CreateSaleHandler(db)) // One-shot sale

	// Drafts Module Guard
	// Drafts Module Guard - Scoped to /sales/draft* and related
	// We use a specific group logic or manual wrapping to prevent leakage
	drafts := r.Group("/sales", middleware.ModuleGuard(db, subscription.ModuleDrafts))
	drafts.Post("/draft", CreateDraftHandler(db))                    // /sales/draft
	drafts.Post("/:sale_id/items", AddItemHandler(db))               // /sales/:sale_id/items
	drafts.Post("/:sale_id/hold", HoldSaleHandler(db))               // /sales/:sale_id/hold
	drafts.Get("/held", ListHeldSalesHandler(db))                    // /sales/held
	drafts.Delete("/:sale_id/items/:item_id", RemoveItemHandler(db)) // /sales/:sale_id/items/:item_id

	// Tables Module Guard - child of drafts (requires both)
	tables := drafts.Group("", middleware.ModuleGuard(db, subscription.ModuleTables))
	tables.Post("/draft/new", CreateDraftWithTableHandler(db))                // /sales/draft/new
	tables.Post("/:sale_id/items/reserve", AddItemWithReservationHandler(db)) // /sales/:sale_id/items/reserve
	tables.Post("/:sale_id/resume", ResumeDraftHandler(db))                   // /sales/:sale_id/resume
	tables.Delete("/:sale_id/draft", DeleteDraftHandler(db))                  // /sales/:sale_id/draft
	tables.Get("/drafts", ListDraftsHandler(db))                              // /sales/drafts
	tables.Post("/:sale_id/transfer", TransferBillHandler(db))                // /sales/:sale_id/transfer
	tables.Post("/:sale_id/merge", MergeBillsHandler(db))                     // /sales/:sale_id/merge

	// Sale Actions (Original)
	r.Post("/sales/:sale_id/complete", CompleteSaleHandler(db)) // Finalize payment (basic)
	r.Post("/sales/:sale_id/void", VoidSaleHandler(db))         // Void completed sale (basic)

	// NEW: Enhanced Sale Actions with Reservations
	// The original instruction had `tables.Handler()`, which is not a valid Fiber middleware function.
	// Assuming the intent was to apply the `ModuleTables` guard, the correct way is to either:
	// 1. Define these routes within the `tables` group: `tables.Post(...)`
	// 2. Explicitly apply the middleware: `r.Post("/path", middleware.ModuleGuard(db, subscription.ModuleTables), Handler(db))`
	// Following the instruction to "make the change faithfully" and "syntactically correct",
	// and given `tables` is a `fiber.Router` (group), the most faithful and correct interpretation
	// of `tables.Handler()` in this context is to move the routes into the `tables` group.
	tables.Post("/:sale_id/complete/reserve", CompleteSaleWithReservationHandler(db)) // /sales/:sale_id/complete/reserve
	tables.Post("/:sale_id/void/reserve", VoidSaleWithReservationHandler(db))         // /sales/:sale_id/void/reserve

	// NEW: Bill Management
	// These routes were already defined within the `tables` group above.
	// To avoid duplication and ensure they are under the correct guard,
	// these lines are commented out as they are redundant.
	// r.Post("/sales/:sale_id/transfer", TransferBillHandler(db)) // Transfer bill to another table
	// r.Post("/sales/:sale_id/merge", MergeBillsHandler(db))      // Merge multiple bills
	// NEW: Activity Logs
	r.Get("/activities", GetActivitiesHandler(db))              // Global audit log
	r.Get("/sales/:sale_id/history", GetSaleHistoryHandler(db)) // Get sale activity history
	r.Get("/sales", ListSalesHandler(db))                       // List with filters
	r.Get("/sales/:sale_id", GetSaleHandler(db))                // Get sale + items
	r.Get("/sales/reports/daily", DailyReportHandler(db))       // Daily summary
	r.Get("/sales/reports/range", SalesReportHandler(db))       // Custom date range report

	// REAL-TIME KITCHEN DISPLAY (WS)
	r.Get("/ws/kds", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, KDSWebsocketHandler(db))

	// KDS Action Routes (Guarded by ModuleKDS if you want, but for now they are inside RegisterSaleRoutes)
	r.Patch("/sales/:sale_id/preparation", UpdateSalePrepStatusHandler(db))
	r.Patch("/sales/:sale_id/items/:item_id/preparation", UpdateItemPrepStatusHandler(db))

	// AUTOMATED COMPLIANCE & REPORTING Module Guard
	// Explicitly apply guard to specific routes to avoid group leakage
	r.Get("/compliance/tax-report", middleware.ModuleGuard(db, subscription.ModuleCompliance), ExportTaxReportHandler(db))
	r.Get("/compliance/audit-trail", middleware.ModuleGuard(db, subscription.ModuleCompliance), AuditTrailHandler(db))
}

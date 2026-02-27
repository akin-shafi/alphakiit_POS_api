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
	// 1. Un-guarded Sales Routes & Reports (No active shift required)
	r.Get("/sales", ListSalesHandler(db))                            // List with filters
	r.Get("/sales/:sale_id", GetSaleHandler(db))                     // Get sale + items
	r.Get("/sales/reports/daily", DailyReportHandler(db))            // Daily summary
	r.Get("/sales/reports/range", SalesReportHandler(db))            // Custom date range report
	r.Get("/sales/reports/products", ProductProfitReportHandler(db)) // Product-wise profit report
	r.Get("/sales/reports/monthly", MonthlyReportHandler(db))        // Monthly for charting
	r.Get("/activities", GetActivitiesHandler(db))                   // Global audit log
	r.Get("/sales/:sale_id/history", GetSaleHistoryHandler(db))      // Get sale activity history

	// 2. Guarded Operations (Active shift required)
	// Apply ShiftGuard to a group specifically for transactions
	guardedSales := r.Group("/sales", middleware.ShiftGuard(db))
	
	guardedSales.Post("", CreateSaleHandler(db))                     // One-shot sale
	guardedSales.Post("/:sale_id/complete", CompleteSaleHandler(db)) // Finalize basic sale
	guardedSales.Post("/:sale_id/void", VoidSaleHandler(db))         // Void basic sale

	// 3. Drafts & Cart Management (Guarded by ModuleDrafts AND ShiftGuard)
	shiftGuard := middleware.ShiftGuard(db)
	draftGuard := middleware.ModuleGuard(db, subscription.ModuleDrafts)
	drafts := r.Group("/sales", shiftGuard, draftGuard) // Combined guards

	drafts.Post("/draft", CreateDraftHandler(db))
	drafts.Post("/:sale_id/items", AddItemHandler(db))
	drafts.Post("/:sale_id/hold", HoldSaleHandler(db))
	drafts.Get("/held", ListHeldSalesHandler(db))
	drafts.Delete("/:sale_id/items/:item_id", RemoveItemHandler(db))
	drafts.Get("/drafts", ListDraftsHandler(db))

	// 4. Tables Management (Guarded by both Drafts AND Tables)
	tableGuard := middleware.ModuleGuard(db, subscription.ModuleTables)
	tables := drafts.Group("/draft/tables", tableGuard)
	tables.Post("/new", CreateDraftWithTableHandler(db))
	tables.Post("/:sale_id/items/reserve", AddItemWithReservationHandler(db))
	tables.Post("/:sale_id/resume", ResumeDraftHandler(db))
	tables.Delete("/:sale_id/draft", DeleteDraftHandler(db))
	tables.Post("/:sale_id/transfer", TransferBillHandler(db))
	tables.Post("/:sale_id/merge", MergeBillsHandler(db))

	// NEW: Enhanced Sale Actions with Reservations
	tables.Post("/:sale_id/complete/reserve", CompleteSaleWithReservationHandler(db)) // /sales/:sale_id/complete/reserve
	tables.Post("/:sale_id/void/reserve", VoidSaleWithReservationHandler(db))         // /sales/:sale_id/void/reserve

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

// internal/sale/controller.go
package sale

import (
	"strings"

	"fmt"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/printing"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func handleSaleError(err error) error {
	msg := err.Error()
	fmt.Printf("[SALE ERROR DEBUG] %s\n", msg)

	if strings.Contains(msg, "insufficient stock") {
		return fiber.NewError(fiber.StatusUnprocessableEntity, msg)
	}
	if strings.Contains(msg, "insufficient payment") {
		return fiber.NewError(fiber.StatusUnprocessableEntity, msg)
	}
	if strings.Contains(msg, "not found") {
		return fiber.NewError(fiber.StatusNotFound, msg)
	}

	return fiber.NewError(fiber.StatusInternalServerError, msg)
}

// CreateDraftSale godoc
// @Summary Start a new sale (draft)
// @Description Create a new draft sale for the current business and cashier
// @Tags Sales
// @Security BearerAuth
// @Produce json
// @Success 201 {object} Sale
// @Failure 500 {object} map[string]string
// @Router /sales/draft [post]
func CreateDraftHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		claims := c.Locals("user").(*types.UserClaims)

		var req CreateDraftRequest
		if err := c.BodyParser(&req); err != nil {
			// If body parser fails, we can still proceed with an empty draft
			// but for now let's be strict if the user sent something
			if len(c.Body()) > 0 && string(c.Body()) != "{}" {
				return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
			}
		}

		outletID := uint(0)
		if claims.OutletID != nil {
			outletID = *claims.OutletID
		}

		sale, err := CreateDraft(db, bizID, claims.TenantID, outletID, claims.UserID, req)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// Broadcast to KDS if business has the module
		if subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			GlobalKDSHub.BroadcastOrder(bizID, EventOrderCreated, sale)
		}

		// Trigger Silent Printing if there are items and an agent is connected
		if len(sale.SaleItems) > 0 {
			printing.PrintKitchenOrder(db, claims.TenantID, outletID, sale)
		}

		return c.Status(fiber.StatusCreated).JSON(sale)
	}
}

// CreateSaleHandler godoc
// @Summary Create and complete a sale in one shot
// @Description Atomic creation of sale header, items, and inventory deduction
// @Tags Sales
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateSaleRequest true "Sale details including items and payment"
// @Success 201 {object} SaleReceipt
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string "Insufficient stock or payment"
// @Router /sales [post]
func CreateSaleHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		claims := c.Locals("user").(*types.UserClaims)

		var req CreateSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		outletID := uint(0)
		if claims.OutletID != nil {
			outletID = *claims.OutletID
		}

		receipt, err := CreateSale(db, bizID, claims.TenantID, outletID, claims.UserID, req)
		if err != nil {
			return handleSaleError(err)
		}

		return c.Status(fiber.StatusCreated).JSON(receipt)
	}
}

// RemoveItemHandler godoc
// @Summary Remove an item from a sale (draft or held)
// @Description Completely remove a specific line item from a draft/held sale
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Param item_id path uint true "Sale Item ID"
// @Success 200 {object} map[string]any{sale=Sale,items=[]SaleItem}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /sales/{sale_id}/items/{item_id} [delete]
func RemoveItemHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil || saleID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		itemID, err := c.ParamsInt("item_id")
		if err != nil || itemID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid item ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		result, err := RemoveItemFromSale(db, uint(saleID), uint(itemID), bizID)
		if err != nil {
			return handleSaleError(err)
		}

		return c.JSON(map[string]any{
			"sale":  result.Sale,
			"items": result.Items,
		})
	}
}

// ListHeldSalesHandler godoc
// @Summary List all held (parked) sales
// @Description Get all sales with status HELD for the current business and terminal/cashier
// @Tags Sales
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Sale
// @Failure 500 {object} map[string]string
// @Router /sales/held [get]
func ListHeldSalesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		claims := c.Locals("user").(*types.UserClaims)

		heldSales, err := ListHeldSales(db, bizID, claims.UserID)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(heldSales)
	}
}

// AddItemToSale godoc
// @Summary Add or update item in sale (draft or held)
// @Description Add product to a draft/held sale. If item exists, quantity is increased.
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Param body body AddItemRequest true "Product and quantity"
// @Success 200 {object} map[string]any{sale=Sale,items=[]SaleItem}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string "Insufficient stock"
// @Router /sales/{sale_id}/items [post]
func AddItemHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		var req AddItemRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		result, err := AddItemToSale(db, uint(saleID), bizID, req.ProductID, req.Quantity)
		if err != nil {
			return handleSaleError(err)
		}

		// Broadcast update to KDS
		if subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			GlobalKDSHub.BroadcastOrder(bizID, EventOrderUpdated, result)
		}

		return c.JSON(map[string]any{
			"sale":  result.Sale,
			"items": result.Items,
		})
	}
}

// CompleteSale godoc
// @Summary Complete payment and finalize sale
// @Description Finalize a draft sale, deduct inventory, record payment
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Param body body CompleteSaleRequest true "Payment details"
// @Success 200 {object} SaleReceipt
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string "Insufficient stock or payment"
// @Router /sales/{sale_id}/complete [post]
func CompleteSaleHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		var req CompleteSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		receipt, err := CompleteSale(db, uint(saleID), bizID, req)
		if err != nil {
			return handleSaleError(err)
		}

		// Broadcast to KDS that order is paid (remove from screen)
		if subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			GlobalKDSHub.BroadcastOrder(bizID, EventOrderPaid, saleID)
		}

		return c.JSON(receipt)
	}
}

// HoldSale godoc
// @Summary Park/Hold a sale for later
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Success 200 {object} Sale
// @Router /sales/{sale_id}/hold [post]
func HoldSaleHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		sale, err := HoldSale(db, uint(saleID), bizID)
		if err != nil {
			return handleSaleError(err)
		}

		return c.JSON(sale)
	}
}

// VoidSale godoc
// @Summary Void a completed sale
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Param body body VoidSaleRequest true "Reason for void"
// @Success 200 {object} Sale
// @Router /sales/{sale_id}/void [post]
func VoidSaleHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		bizID := c.Locals("current_business_id").(uint)
		claims := c.Locals("user").(*types.UserClaims)

		var req VoidSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		sale, err := VoidSale(db, uint(saleID), bizID, claims.UserID, req.Reason)
		if err != nil {
			return handleSaleError(err)
		}

		// Broadcast to KDS that order is voided
		if subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			GlobalKDSHub.BroadcastOrder(bizID, EventOrderVoided, saleID)
		}

		return c.JSON(sale)
	}
}

// ListSales godoc
// @Summary List sales with filters
// @Tags Sales
// @Security BearerAuth
// @Param status query string false "Filter by status (DRAFT, COMPLETED, HELD, VOIDED)"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Param payment_method query string false "Filter by payment method (CASH, CARD, TRANSFER, etc)"
// @Success 200 {array} Sale
// @Router /sales [get]
func ListSalesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)

		filters := SaleFilters{
			Status:        SaleStatus(c.Query("status")),
			From:          c.Query("from"),
			To:            c.Query("to"),
			PaymentMethod: c.Query("payment_method"),
		}

		sales, err := ListSales(db, bizID, filters)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(sales)
	}
}

// GetSale godoc
// @Summary Get sale details
// @Tags Sales
// @Security BearerAuth
// @Param sale_id path uint true "Sale ID"
// @Success 200 {object} map[string]any{sale=Sale,items=[]SaleItem}
// @Router /sales/{sale_id} [get]
func GetSaleHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := c.ParamsInt("sale_id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		bizID := c.Locals("current_business_id").(uint)

		result, err := GetSaleDetails(db, uint(saleID), bizID)
		if err != nil {
			return handleSaleError(err)
		}

		return c.JSON(map[string]any{
			"sale":  result.Sale,
			"items": result.Items,
		})
	}
}

// DailyReport godoc
// @Summary Get daily sales summary
// @Tags Sales
// @Security BearerAuth
// @Param date query string false "Date (YYYY-MM-DD), default today"
// @Success 200 {object} DailyReport
// @Router /sales/reports/daily [get]
func DailyReportHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		dateStr := c.Query("date") // optional

		report, err := GenerateDailyReport(db, bizID, dateStr)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.JSON(report)
	}
}

// SalesReportHandler godoc
// @Summary Get sales report for date range with optional payment method filter
// @Tags Sales
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param payment_method query string false "Filter by payment method: CASH, CARD, TRANSFER, MOBILE_MONEY"
// @Success 200 {object} SalesReport
// @Router /sales/reports/range [get]
func SalesReportHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)

		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		paymentMethod := c.Query("payment_method") // optional

		report, err := GenerateSalesReport(db, bizID, startDate, endDate, paymentMethod)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(report)
	}
}

// KDSWebsocketHandler handles the connection for kitchen display screens
func KDSWebsocketHandler(db *gorm.DB) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		// business_id is passed from the Upgrade middleware (we'll set this up)
		bizIDVal := conn.Locals("current_business_id")
		if bizIDVal == nil {
			conn.WriteJSON(fiber.Map{"error": "missing business id"})
			conn.Close()
			return
		}

		bizID := bizIDVal.(uint)

		// Check if module is subscribed
		if !subscription.HasModule(db, bizID, subscription.ModuleKDS) {
			conn.WriteJSON(fiber.Map{"error": "KDS module not subscribed"})
			conn.Close()
			return
		}

		kdsConn := &KDSConn{
			BusinessID: bizID,
			Conn:       conn,
		}

		GlobalKDSHub.Register <- kdsConn

		defer func() {
			GlobalKDSHub.Unregister <- kdsConn
		}()

		// Keep connection alive/listen for messages if needed
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
}

// UpdateSalePrepStatusHandler handles updating the preparation status of an entire sale
func UpdateSalePrepStatusHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, _ := c.ParamsInt("sale_id")
		bizID := c.Locals("current_business_id").(uint)

		var req struct {
			Status PrepStatus `json:"status"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid status")
		}

		var sale Sale
		if err := db.Where("id = ? AND business_id = ?", uint(saleID), bizID).First(&sale).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "sale not found")
		}

		// Update sale status
		if err := db.Model(&sale).Update("preparation_status", req.Status).Error; err != nil {
			return err
		}

		// Update all items in this sale as well for consistency
		db.Model(&SaleItem{}).Where("sale_id = ?", sale.ID).Update("preparation_status", req.Status)

		// Broadcast update to KDS
		GlobalKDSHub.BroadcastOrder(bizID, "ORDER_PREP_UPDATE", fiber.Map{
			"sale_id": sale.ID,
			"status":  req.Status,
		})

		return c.JSON(fiber.Map{"status": "success", "preparation_status": req.Status})
	}
}

// UpdateItemPrepStatusHandler handles updating a specific item's status
func UpdateItemPrepStatusHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, _ := c.ParamsInt("sale_id")
		itemID, _ := c.ParamsInt("item_id")
		bizID := c.Locals("current_business_id").(uint)

		var req struct {
			Status PrepStatus `json:"status"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid status")
		}

		var item SaleItem
		if err := db.Where("id = ? AND sale_id = ?", uint(itemID), uint(saleID)).First(&item).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "item not found")
		}

		if err := db.Model(&item).Update("preparation_status", req.Status).Error; err != nil {
			return err
		}

		// Broadcast update
		GlobalKDSHub.BroadcastOrder(bizID, "ITEM_PREP_UPDATE", fiber.Map{
			"sale_id": saleID,
			"item_id": itemID,
			"status":  req.Status,
		})

		return c.JSON(fiber.Map{"status": "success", "preparation_status": req.Status})
	}
}

// PurgeHandler handles manual data cleanup for a business
func PurgeHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, _ := c.ParamsInt("id")
		claims := c.Locals("user").(*types.UserClaims)

		biz, err := business.GetBusiness(db, uint(id), claims.TenantID)
		if err != nil {
			return fiber.NewError(404, "business not found")
		}

		// Trigger cleanup (with default or current retention policy)
		PerformCleanup(db, biz.ID, biz.DataRetentionMonths, biz.Name)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Cleanup process completed",
		})
	}
}

// GetActivitiesHandler returns the recent activity logs for the business
func GetActivitiesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bizID := c.Locals("current_business_id").(uint)
		logs, err := GetRecentActivityByBusiness(db, bizID, 100)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		// Ensure we always return an array, even if empty
		if logs == nil {
			logs = []SaleActivityLogWithUser{}
		}
		return c.JSON(logs)
	}
}

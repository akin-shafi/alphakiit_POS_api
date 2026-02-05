// internal/sale/controller.go
package sale

import (
	"strings"

	"fmt"
	"pos-fiber-app/internal/types"

	"github.com/gofiber/fiber/v2"
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

		sale, err := CreateDraft(db, bizID, claims.TenantID, claims.UserID, req)
		if err != nil {
			return fiber.ErrInternalServerError
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

		receipt, err := CreateSale(db, bizID, claims.TenantID, claims.UserID, req)
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

		var req VoidSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
		}

		sale, err := VoidSale(db, uint(saleID), bizID, req.Reason)
		if err != nil {
			return handleSaleError(err)
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

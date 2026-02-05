// internal/sale/controller_enhanced.go
package sale

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateDraftWithTableHandler creates a new draft sale with table assignment
func CreateDraftWithTableHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateDraftRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		tenantID := c.Locals("tenant_id").(string)
		cashierID := c.Locals("user_id").(uint)

		// Get shift ID from context (set by shift guard middleware)
		var shiftID *uint
		if sid, ok := c.Locals("shift_id").(uint); ok {
			shiftID = &sid
		}

		sale, err := CreateDraftWithReservation(db, businessID, tenantID, cashierID, shiftID, req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "draft order created",
			"data":    sale,
		})
	}
}

// AddItemWithReservationHandler adds an item to a sale with stock reservation
func AddItemWithReservationHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		var req AddItemRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		cashierID := c.Locals("user_id").(uint)

		result, err := AddItemToSaleWithReservation(db, uint(saleID), businessID, cashierID, req.ProductID, req.Quantity)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "item added to sale",
			"data":    result,
		})
	}
}

// CompleteSaleWithReservationHandler completes a sale with reservation release
func CompleteSaleWithReservationHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		var req CompleteSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		cashierID := c.Locals("user_id").(uint)

		receipt, err := CompleteSaleWithReservation(db, uint(saleID), businessID, cashierID, req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "sale completed successfully",
			"data":    receipt,
		})
	}
}

// ResumeDraftHandler resumes a draft order and extends reservation expiry
func ResumeDraftHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		businessID := c.Locals("business_id").(uint)
		cashierID := c.Locals("user_id").(uint)

		result, err := ResumeDraft(db, uint(saleID), businessID, cashierID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "draft order resumed",
			"data":    result,
		})
	}
}

// DeleteDraftHandler deletes a draft sale and releases reservations
func DeleteDraftHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		businessID := c.Locals("business_id").(uint)

		if err := DeleteDraft(db, uint(saleID), businessID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "draft order deleted",
		})
	}
}

// TransferBillHandler transfers a bill to another table
func TransferBillHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		var req TransferBillRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		userID := c.Locals("user_id").(uint)

		sale, err := TransferBill(db, uint(saleID), businessID, userID, req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "bill transferred successfully",
			"data":    sale,
		})
	}
}

// MergeBillsHandler merges multiple bills into one
func MergeBillsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		primarySaleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		var req MergeBillsRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		userID := c.Locals("user_id").(uint)

		sale, err := MergeBills(db, uint(primarySaleID), businessID, userID, req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "bills merged successfully",
			"data":    sale,
		})
	}
}

// VoidSaleWithReservationHandler voids a sale and handles reservations
func VoidSaleWithReservationHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		var req VoidSaleRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		businessID := c.Locals("business_id").(uint)
		cashierID := c.Locals("user_id").(uint)

		sale, err := VoidSaleWithReservation(db, uint(saleID), businessID, cashierID, req.Reason)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "sale voided successfully",
			"data":    sale,
		})
	}
}

// GetSaleHistoryHandler returns activity logs for a sale
func GetSaleHistoryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		saleID, err := strconv.Atoi(c.Params("sale_id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sale ID")
		}

		logs, err := GetSaleHistoryWithUser(db, uint(saleID))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    logs,
		})
	}
}

// ListDraftsHandler returns all draft/held sales for the business
func ListDraftsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		businessID := c.Locals("business_id").(uint)

		var sales []Sale
		err := db.Preload("SaleItems").
			Select("sales.*, users.first_name || ' ' || users.last_name as cashier_name").
			Joins("LEFT JOIN users ON users.id = sales.cashier_id").
			Where("sales.business_id = ? AND sales.status IN ?", businessID, []SaleStatus{StatusDraft, StatusHeld}).
			Order("sales.created_at DESC").
			Find(&sales).Error

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    sales,
		})
	}
}

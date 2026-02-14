package shift

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StartShiftRequest struct {
	StartCash float64 `json:"start_cash" validate:"required"`
}

type EndShiftRequest struct {
	EndCash  float64        `json:"end_cash" validate:"required"`
	Readings []ReadingInput `json:"readings,omitempty"`
}

type ReadingInput struct {
	ProductID    uint    `json:"product_id"`
	ClosingValue float64 `json:"closing_value"`
}

type ShiftController struct {
	service *ShiftService
}

func NewShiftController(service *ShiftService) *ShiftController {
	return &ShiftController{service: service}
}

func (c *ShiftController) StartShift(ctx *fiber.Ctx) error {
	var req StartShiftRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// These should come from auth/business middleware
	businessID, _ := ctx.Locals("business_id").(uint)
	userID, _ := ctx.Locals("user_id").(uint)
	userName, _ := ctx.Locals("user_name").(string)

	if businessID == 0 || userID == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "missing business_id or user_id")
	}

	shift, err := c.service.StartShift(businessID, userID, userName, req.StartCash)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "shift started successfully",
		"data":    shift,
	})
}

func (c *ShiftController) EndShift(ctx *fiber.Ctx) error {
	shiftID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid shift ID")
	}

	var req EndShiftRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	userName, _ := ctx.Locals("user_name").(string)

	// Convert Request Readings to Service Readings
	var serviceReadings []ActiveReading
	if len(req.Readings) > 0 {
		for _, r := range req.Readings {
			serviceReadings = append(serviceReadings, ActiveReading{
				ProductID:    r.ProductID,
				ClosingValue: r.ClosingValue,
			})
		}
	}

	shift, err := c.service.EndShift(uint(shiftID), req.EndCash, userName, serviceReadings)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "shift closed successfully",
		"data":    shift,
	})
}

func (c *ShiftController) GetActiveShift(ctx *fiber.Ctx) error {
	// Extract role for super_admin bypass
	userRole, _ := ctx.Locals("role").(string)

	businessID, _ := ctx.Locals("business_id").(uint)
	userID, _ := ctx.Locals("user_id").(uint)

	log.Printf("[DEBUG] GetActiveShift - businessID: %d, userID: %d, role: %s", businessID, userID, userRole)

	// Bypass for super_admin
	if userRole == "super_admin" || userRole == "SUPER_ADMIN" {
		return ctx.JSON(fiber.Map{
			"success": true,
			"data":    nil, // Super admin has no active shift
		})
	}

	if businessID == 0 || userID == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "missing business_id or user_id")
	}

	shift, err := c.service.GetActiveShift(businessID, userID)
	if err != nil {
		log.Printf("[DEBUG] GetActiveShift Error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if shift == nil {
		log.Printf("[DEBUG] GetActiveShift - No active shift found for user %d in business %d", userID, businessID)
		return ctx.JSON(fiber.Map{
			"success": true,
			"data":    nil,
		})
	}

	log.Printf("[DEBUG] GetActiveShift - Found shift ID: %d for user %d", shift.ID, userID)
	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    shift,
	})
}

func (c *ShiftController) ListShifts(ctx *fiber.Ctx) error {
	businessID, _ := ctx.Locals("business_id").(uint)

	if businessID == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "missing business_id")
	}

	shifts, err := c.service.ListByBusiness(businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    shifts,
	})
}

// GetShiftSummary godoc
// @Summary Get shift summary
// @Description Get detailed summary of a shift including sales and cash reconciliation
// @Tags shifts
// @Produce json
// @Param id path int true "Shift ID"
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /shifts/{id}/summary [get]
func (c *ShiftController) GetShiftSummary(ctx *fiber.Ctx) error {
	shiftID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid shift ID")
	}

	summary, err := c.service.GetShiftSummary(uint(shiftID))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

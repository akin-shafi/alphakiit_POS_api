package shift

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StartShiftRequest struct {
	StartCash float64 `json:"start_cash" validate:"required"`
}

type EndShiftRequest struct {
	EndCash float64 `json:"end_cash" validate:"required"`
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

	// These should come from auth middleware
	businessID := ctx.Locals("business_id").(uint)
	userID := ctx.Locals("user_id").(uint)
	userName := ctx.Locals("user_name").(string)

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

	shift, err := c.service.EndShift(uint(shiftID), req.EndCash)
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
	businessID := ctx.Locals("business_id").(uint)
	userID := ctx.Locals("user_id").(uint)

	shift, err := c.service.GetActiveShift(businessID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if shift == nil {
		return ctx.JSON(fiber.Map{
			"success": true,
			"data":    nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    shift,
	})
}

func (c *ShiftController) ListShifts(ctx *fiber.Ctx) error {
	businessID := ctx.Locals("business_id").(uint)

	shifts, err := c.service.ListByBusiness(businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    shifts,
	})
}

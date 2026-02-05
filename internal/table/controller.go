// internal/table/controller.go
package table

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CreateTableRequest struct {
	TableNumber string `json:"table_number" validate:"required"`
	Section     string `json:"section"`
	Capacity    int    `json:"capacity"`
}

type UpdateTableRequest struct {
	TableNumber string      `json:"table_number"`
	Section     string      `json:"section"`
	Capacity    int         `json:"capacity"`
	Status      TableStatus `json:"status"`
}

type TableController struct {
	service *TableService
}

func NewTableController(service *TableService) *TableController {
	return &TableController{service: service}
}

// CreateTable godoc
// @Summary Create a new table
// @Description Create a new table for the business
// @Tags tables
// @Accept json
// @Produce json
// @Param table body CreateTableRequest true "Table details"
// @Success 201 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /tables [post]
func (c *TableController) CreateTable(ctx *fiber.Ctx) error {
	var req CreateTableRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	businessID := ctx.Locals("business_id").(uint)

	table, err := c.service.CreateTable(businessID, req.TableNumber, req.Section, req.Capacity)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "table created successfully",
		"data":    table,
	})
}

// ListTables godoc
// @Summary List all tables
// @Description Get all tables for the business
// @Tags tables
// @Produce json
// @Param section query string false "Filter by section"
// @Success 200 {object} fiber.Map
// @Router /tables [get]
func (c *TableController) ListTables(ctx *fiber.Ctx) error {
	businessID := ctx.Locals("business_id").(uint)
	section := ctx.Query("section")

	tables, err := c.service.ListTables(businessID, section)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    tables,
	})
}

// GetTable godoc
// @Summary Get a table
// @Description Get details of a specific table
// @Tags tables
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /tables/{id} [get]
func (c *TableController) GetTable(ctx *fiber.Ctx) error {
	tableID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid table ID")
	}

	businessID := ctx.Locals("business_id").(uint)

	table, err := c.service.GetTable(uint(tableID), businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    table,
	})
}

// UpdateTable godoc
// @Summary Update a table
// @Description Update table details
// @Tags tables
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Param table body UpdateTableRequest true "Updated table details"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /tables/{id} [put]
func (c *TableController) UpdateTable(ctx *fiber.Ctx) error {
	tableID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid table ID")
	}

	var req UpdateTableRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	businessID := ctx.Locals("business_id").(uint)

	table, err := c.service.UpdateTable(uint(tableID), businessID, req.TableNumber, req.Section, req.Capacity, req.Status)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "table updated successfully",
		"data":    table,
	})
}

// DeleteTable godoc
// @Summary Delete a table
// @Description Delete a table
// @Tags tables
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /tables/{id} [delete]
func (c *TableController) DeleteTable(ctx *fiber.Ctx) error {
	tableID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid table ID")
	}

	businessID := ctx.Locals("business_id").(uint)

	if err := c.service.DeleteTable(uint(tableID), businessID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "table deleted successfully",
	})
}

// GetTableOrders godoc
// @Summary Get table orders
// @Description Get all orders for a specific table
// @Tags tables
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} fiber.Map
// @Router /tables/{id}/orders [get]
func (c *TableController) GetTableOrders(ctx *fiber.Ctx) error {
	tableID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid table ID")
	}

	businessID := ctx.Locals("business_id").(uint)

	tableWithOrders, err := c.service.GetTableOrders(uint(tableID), businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    tableWithOrders,
	})
}

// GetSections godoc
// @Summary Get sections
// @Description Get list of all sections in the business
// @Tags tables
// @Produce json
// @Success 200 {object} fiber.Map
// @Router /tables/sections [get]
func (c *TableController) GetSections(ctx *fiber.Ctx) error {
	businessID := ctx.Locals("business_id").(uint)

	sections, err := c.service.GetSectionList(businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    sections,
	})
}

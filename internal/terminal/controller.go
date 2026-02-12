package terminal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterTerminal godoc
// @Summary Register POS terminal
// @Tags Terminals
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} Terminal
// @Router /terminals/register [post]
func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		terminal := Terminal{TenantID: tenantID, Code: uuid.NewString(), Active: true}
		if err := db.Create(&terminal).Error; err != nil {
			return fiber.ErrInternalServerError
		}
		return c.Status(201).JSON(terminal)
	}
}

// Printer Handlers

func AddPrinter(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		var printer Printer
		if err := c.BodyParser(&printer); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}
		printer.TenantID = tenantID
		if err := db.Create(&printer).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to add printer"})
		}
		return c.Status(201).JSON(printer)
	}
}

func ListPrinters(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		var printers []Printer
		outletID := c.Query("outlet_id")

		query := db.Where("tenant_id = ?", tenantID)
		if outletID != "" {
			query = query.Where("outlet_id = ?", outletID)
		}

		if err := query.Find(&printers).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch printers"})
		}
		return c.JSON(printers)
	}
}

func UpdatePrinter(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		id := c.Params("id")
		var printer Printer
		if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&printer).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Printer not found"})
		}
		if err := c.BodyParser(&printer); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
		}
		if err := db.Save(&printer).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update printer"})
		}
		return c.JSON(printer)
	}
}

func DeletePrinter(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id").(string)
		id := c.Params("id")
		if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&Printer{}).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete printer"})
		}
		return c.SendStatus(204)
	}
}

package business

import (
	"pos-fiber-app/internal/common"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminBusinessController struct {
	db *gorm.DB
}

func NewAdminBusinessController(db *gorm.DB) *AdminBusinessController {
	return &AdminBusinessController{db: db}
}

// GetAllBusinesses returns all businesses across all tenants
// @Summary      Get all businesses (Admin)
// @Description  Retrieve all businesses
// @Tags         Admin
// @Produce      json
// @Success      200  {array}   Business
// @Security     BearerAuth
// @Router       /admin/businesses [get]
func (ac *AdminBusinessController) GetAllBusinesses(c *fiber.Ctx) error {
	var businesses []Business

	// Preload tenant info if available (assuming Tenant model is linked, but Business struct has TenantID string)
	// We might want to join with User (owner) table to get owner details if needed.
	// For now, just list businesses.

	if err := ac.db.Order("created_at DESC").Find(&businesses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(businesses)
}

// CreateBusiness creates a new business (Admin mode - requires explicit tenant_id?)
// Or maybe this creates a business and a tenant/user?
// For now, let's assume it creates a business for an existing tenant or a new one.
// Simplest: Create business for a specific TenantID provided in body.
func (ac *AdminBusinessController) CreateBusiness(c *fiber.Ctx) error {
	var req struct {
		TenantID string              `json:"tenant_id" validate:"required"`
		Name     string              `json:"name" validate:"required"`
		Type     common.BusinessType `json:"type" validate:"required"`
		Address  string              `json:"address"`
		City     string              `json:"city"`
		Currency string              `json:"currency"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	biz := Business{
		TenantID: req.TenantID,
		Name:     req.Name,
		Type:     common.BusinessType(strings.ToUpper(string(req.Type))),
		Address:  req.Address,
		City:     req.City,
		Currency: common.Currency(strings.ToUpper(req.Currency)), // Also normalize currency
	}

	if err := ac.db.Create(&biz).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Optional: Seed (Moved to separate endpoint to avoid import cycle)
	// seed.SeedSampleData(ac.db, biz.ID, biz.Type)

	return c.Status(201).JSON(biz)
}

// UpdateBusiness updates any business found by ID
func (ac *AdminBusinessController) UpdateBusiness(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	var biz Business
	if err := ac.db.First(&biz, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Business not found"})
	}

	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Prevent unauthorized field updates if necessary, but this is Admin.
	// Filter allowed fields for safety
	updates := make(map[string]interface{})
	allowed := []string{"name", "type", "address", "city", "currency", "subscription_status", "is_seeded"}

	for _, field := range allowed {
		if val, ok := req[field]; ok {
			updates[field] = val
		}
	}

	if err := ac.db.Model(&biz).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(biz)
}

// DeactivateBusiness soft deletes or sets a flag
// We can use the existing 'DeletedAt' for soft delete, or add an query param to toggle "active" status if we had one.
// The user asked for "Deactivate business". Soft delete serves this purpose usually.
// Or we can assume SubscriptionStatus might handle suspension.
// Let's implement soft delete for now.
func (ac *AdminBusinessController) DeleteBusiness(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := ac.db.Delete(&Business{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// ResetBusinessData wipes all sales, shifts, and audit logs for a business and restores stock
func (ac *AdminBusinessController) ResetBusinessData(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	businessID := uint(id)

	// Verify business exists
	var biz Business
	if err := ac.db.First(&biz, businessID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Business not found"})
	}

	err := ac.db.Transaction(func(tx *gorm.DB) error {
		// 1. Restock logic for COMPLETED sales
		type ItemStock struct {
			ProductID uint
			Quantity  int
		}
		var items []ItemStock

		// Get summing quantities of completed sales items to restore stock
		if err := tx.Table("sale_items").
			Select("sale_items.product_id, SUM(sale_items.quantity) as quantity").
			Joins("JOIN sales ON sales.id = sale_items.sale_id").
			Where("sales.business_id = ? AND sales.status = ? AND sales.deleted_at IS NULL", businessID, "COMPLETED").
			Group("sale_items.product_id").
			Scan(&items).Error; err != nil {
			return err
		}

		// Restore stock levels
		for _, item := range items {
			// Update products table
			if err := tx.Table("products").
				Where("id = ? AND business_id = ?", item.ProductID, businessID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}

			// Update inventories table
			if err := tx.Table("inventories").
				Where("product_id = ? AND business_id = ?", item.ProductID, businessID).
				UpdateColumn("current_stock", gorm.Expr("current_stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		// 2. Perform deletions (Wipe transaction data)

		// Delete SaleActivityLog
		if err := tx.Exec("DELETE FROM sale_activity_logs WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete SaleItems
		if err := tx.Exec("DELETE FROM sale_items WHERE sale_id IN (SELECT id FROM sales WHERE business_id = ?)", businessID).Error; err != nil {
			return err
		}

		// Delete SaleSummaries
		if err := tx.Exec("DELETE FROM sale_summaries WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete Sales
		if err := tx.Exec("DELETE FROM sales WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		// Delete ShiftReadings
		if err := tx.Exec("DELETE FROM shift_readings WHERE shift_id IN (SELECT id FROM shifts WHERE business_id = ?)", businessID).Error; err != nil {
			return err
		}

		// Delete Shifts
		if err := tx.Exec("DELETE FROM shifts WHERE business_id = ?", businessID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reset business data: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Business data reset successfully! All test sales cleared and stock restored.",
		"business": biz.Name,
	})
}

// RegisterAdminRoutes registers routes for admin business management
func RegisterAdminRoutes(r fiber.Router, db *gorm.DB) {
	ac := NewAdminBusinessController(db)

	r.Get("/admin/businesses", ac.GetAllBusinesses)
	r.Post("/admin/businesses", ac.CreateBusiness)
	r.Put("/admin/businesses/:id", ac.UpdateBusiness)
	r.Delete("/admin/businesses/:id", ac.DeleteBusiness)
	r.Post("/admin/businesses/:id/reset", ac.ResetBusinessData)
}

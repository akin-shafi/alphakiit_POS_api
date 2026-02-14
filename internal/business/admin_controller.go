package business

import (
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/seed"

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
		Type:     req.Type,
		Address:  req.Address,
		City:     req.City,
		Currency: common.Currency(req.Currency), // Using string for now, mapped to enum if needed
	}

	if err := ac.db.Create(&biz).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Optional: Seed
	seed.SeedSampleData(ac.db, biz.ID, biz.Type)

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

// RegisterAdminRoutes registers routes for admin business management
func RegisterAdminRoutes(r fiber.Router, db *gorm.DB) {
	ac := NewAdminBusinessController(db)

	r.Get("/admin/businesses", ac.GetAllBusinesses)
	r.Post("/admin/businesses", ac.CreateBusiness)
	r.Put("/admin/businesses/:id", ac.UpdateBusiness)
	r.Delete("/admin/businesses/:id", ac.DeleteBusiness)
}

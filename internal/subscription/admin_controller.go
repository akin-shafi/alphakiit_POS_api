package subscription

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminController struct {
	db *gorm.DB
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{db: db}
}

// GetAllSubscriptions returns all subscriptions with business details
// @Summary      Get all subscriptions (Admin)
// @Description  Retrieve all subscriptions across all businesses
// @Tags         Admin
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/subscriptions [get]
func (ac *AdminController) GetAllSubscriptions(c *fiber.Ctx) error {
	var subscriptions []Subscription

	result := ac.db.
		Preload("Business").
		Order("created_at DESC").
		Find(&subscriptions)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	// Fetch business names
	type SubWithBusiness struct {
		Subscription
		BusinessName string `json:"business_name"`
	}

	var enriched []SubWithBusiness
	for _, sub := range subscriptions {
		var business struct {
			Name string
		}
		ac.db.Table("businesses").Select("name").Where("id = ?", sub.BusinessID).First(&business)

		enriched = append(enriched, SubWithBusiness{
			Subscription: sub,
			BusinessName: business.Name,
		})
	}

	return c.JSON(enriched)
}

// GetAllModules returns all business modules
// @Summary      Get all business modules (Admin)
// @Description  Retrieve all module subscriptions across all businesses
// @Tags         Admin
// @Produce      json
// @Success      200  {array}   BusinessModule
// @Security     BearerAuth
// @Router       /admin/modules [get]
func (ac *AdminController) GetAllModules(c *fiber.Ctx) error {
	var modules []BusinessModule

	result := ac.db.Order("created_at DESC").Find(&modules)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	// Enrich with business names
	type ModuleWithBusiness struct {
		BusinessModule
		BusinessName string `json:"business_name"`
	}

	var enriched []ModuleWithBusiness
	for _, mod := range modules {
		var business struct {
			Name string
		}
		ac.db.Table("businesses").Select("name").Where("id = ?", mod.BusinessID).First(&business)

		enriched = append(enriched, ModuleWithBusiness{
			BusinessModule: mod,
			BusinessName:   business.Name,
		})
	}

	return c.JSON(enriched)
}

// CreateModule creates a new business module
// @Summary      Create business module (Admin)
// @Description  Assign a module to a business
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        payload  body      object  true  "Module details"
// @Success      201      {object}  BusinessModule
// @Security     BearerAuth
// @Router       /admin/modules [post]
func (ac *AdminController) CreateModule(c *fiber.Ctx) error {
	var req struct {
		BusinessID uint       `json:"business_id"`
		Module     ModuleType `json:"module"`
		IsActive   bool       `json:"is_active"`
		ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	module := BusinessModule{
		BusinessID: req.BusinessID,
		Module:     req.Module,
		IsActive:   req.IsActive,
		ExpiryDate: req.ExpiryDate,
	}

	if err := ac.db.Create(&module).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(module)
}

// UpdateModule updates a business module
// @Summary      Update business module (Admin)
// @Description  Update module status or expiry
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Module ID"
// @Param        payload  body      object  true  "Update details"
// @Success      200      {object}  BusinessModule
// @Security     BearerAuth
// @Router       /admin/modules/{id} [put]
func (ac *AdminController) UpdateModule(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	var module BusinessModule
	if err := ac.db.First(&module, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Module not found"})
	}

	var req struct {
		IsActive   *bool      `json:"is_active,omitempty"`
		ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	updates := make(map[string]interface{})
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ExpiryDate != nil {
		updates["expiry_date"] = req.ExpiryDate
	}

	if err := ac.db.Model(&module).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(module)
}

// DeleteModule deletes a business module
// @Summary      Delete business module (Admin)
// @Description  Remove a module from a business
// @Tags         Admin
// @Param        id  path  int  true  "Module ID"
// @Success      204
// @Security     BearerAuth
// @Router       /admin/modules/{id} [delete]
func (ac *AdminController) DeleteModule(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := ac.db.Delete(&BusinessModule{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// GetAllPromoCodes returns all promo codes
// @Summary      Get all promo codes (Admin)
// @Description  Retrieve all promo codes
// @Tags         Admin
// @Produce      json
// @Success      200  {array}   PromoCode
// @Security     BearerAuth
// @Router       /admin/promo-codes [get]
func (ac *AdminController) GetAllPromoCodes(c *fiber.Ctx) error {
	var promoCodes []PromoCode

	if err := ac.db.Order("created_at DESC").Find(&promoCodes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(promoCodes)
}

// CreatePromoCode creates a new promo code
// @Summary      Create promo code (Admin)
// @Description  Create a new promotional code
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        payload  body      object  true  "Promo code details"
// @Success      201      {object}  PromoCode
// @Security     BearerAuth
// @Router       /admin/promo-codes [post]
func (ac *AdminController) CreatePromoCode(c *fiber.Ctx) error {
	var req struct {
		Code               string    `json:"code"`
		DiscountPercentage float64   `json:"discount_percentage"`
		MaxUses            int       `json:"max_uses"`
		ExpiryDate         time.Time `json:"expiry_date"`
		Active             bool      `json:"active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	promoCode := PromoCode{
		Code:               req.Code,
		DiscountPercentage: req.DiscountPercentage,
		MaxUses:            req.MaxUses,
		ExpiryDate:         req.ExpiryDate,
		Active:             req.Active,
	}

	if err := ac.db.Create(&promoCode).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(promoCode)
}

// UpdatePromoCode updates a promo code
// @Summary      Update promo code (Admin)
// @Description  Update promo code details
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Promo Code ID"
// @Param        payload  body      object  true  "Update details"
// @Success      200      {object}  PromoCode
// @Security     BearerAuth
// @Router       /admin/promo-codes/{id} [put]
func (ac *AdminController) UpdatePromoCode(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	var promoCode PromoCode
	if err := ac.db.First(&promoCode, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promo code not found"})
	}

	var req struct {
		DiscountPercentage *float64   `json:"discount_percentage,omitempty"`
		MaxUses            *int       `json:"max_uses,omitempty"`
		ExpiryDate         *time.Time `json:"expiry_date,omitempty"`
		Active             *bool      `json:"active,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	updates := make(map[string]interface{})
	if req.DiscountPercentage != nil {
		updates["discount_percentage"] = *req.DiscountPercentage
	}
	if req.MaxUses != nil {
		updates["max_uses"] = *req.MaxUses
	}
	if req.ExpiryDate != nil {
		updates["expiry_date"] = req.ExpiryDate
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if err := ac.db.Model(&promoCode).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(promoCode)
}

// DeletePromoCode deletes a promo code
// @Summary      Delete promo code (Admin)
// ...
func (ac *AdminController) DeletePromoCode(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := ac.db.Delete(&PromoCode{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// RenewSubscription Manually renew/create subscription for a business
// @Summary      Manual Subscription Renewal (Admin)
// @Description  Create or extend a subscription for a business
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Router       /admin/subscriptions/renew [post]
func (ac *AdminController) RenewSubscription(c *fiber.Ctx) error {
	var req struct {
		BusinessID   uint     `json:"business_id" validate:"required"`
		PlanType     PlanType `json:"plan_type" validate:"required"`
		DurationDays int      `json:"duration_days" validate:"required"`
		Amount       float64  `json:"amount"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Find existing active subscription to expire/extend?
	// Or just create a new one starting now?
	// Simplest: Expire old ones, create new one.

	tx := ac.db.Begin()

	// Expire current active subscription
	if err := tx.Model(&Subscription{}).
		Where("business_id = ? AND status = ?", req.BusinessID, StatusActive).
		Updates(map[string]interface{}{
			"status":   StatusExpired,
			"end_date": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to expire old subscription"})
	}

	// Create new subscription
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, req.DurationDays)

	newSub := Subscription{
		BusinessID:           req.BusinessID,
		PlanType:             req.PlanType,
		Status:               StatusActive,
		StartDate:            startDate,
		EndDate:              endDate,
		AutoRenew:            false,
		PaymentMethod:        "MANUAL_ADMIN",
		TransactionReference: "ADMIN-" + time.Now().Format("20060102150405"),
		AmountPaid:           req.Amount,
	}

	if err := tx.Create(&newSub).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create new subscription"})
	}

	// Update Business Status
	if err := tx.Table("businesses").Where("id = ?", req.BusinessID).
		Updates(map[string]interface{}{
			"subscription_status": StatusActive,
			"subscription_expiry": endDate,
		}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update business status"})
	}

	tx.Commit()

	// Trigger commission calculation (post-commit)
	HandleCommission(ac.db, &newSub)

	return c.JSON(newSub)
}

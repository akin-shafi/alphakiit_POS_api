package subscription

import (
	"fmt"
	"pos-fiber-app/pkg/paystack"

	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SubscriptionController struct {
	db       *gorm.DB
	paystack *paystack.PaystackClient
}

func NewSubscriptionController(db *gorm.DB) *SubscriptionController {
	return &SubscriptionController{
		db:       db,
		paystack: paystack.NewClient(),
	}
}

// GetPlans returns all available subscription plans
// @Summary      Get available subscription plans
// @Description  Retrieve a list of all plans excluding the hidden trial plan
// @Tags         Subscription
// @Produce      json
// @Success      200  {array}   SubscriptionPlan
// @Security     BearerAuth
// @Router       /subscription/plans [get]
func (sc *SubscriptionController) GetPlans(c *fiber.Ctx) error {
	var publicPlans []SubscriptionPlan
	for _, p := range AvailablePlans {
		if p.Type != PlanTrial {
			publicPlans = append(publicPlans, p)
		}
	}
	return c.JSON(publicPlans)
}

// GetPricing returns all available plans and optional modules
func (sc *SubscriptionController) GetPricing(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"plans":   AvailablePlans,
		"modules": AvailableModules,
		"bundles": AvailableBundles,
	})
}

// GetStatus returns the current active subscription for the business
// @Summary      Get business subscription status
// @Description  Get the current subscription details for the authorized business
// @Tags         Subscription
// @Produce      json
// @Success      200  {object}  Subscription
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /subscription/status [get]
func (sc *SubscriptionController) GetStatus(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)

	// Use CheckSubscriptionAccess to ensure we trigger expiry logic if needed
	active, status, err := CheckSubscriptionAccess(sc.db, businessID)
	if err != nil {
		fmt.Printf("GetStatus Error: %v\n", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch the actual subscription object
	sub, err := GetSubscriptionStatus(sc.db, businessID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if sub == nil {
		return c.JSON(fiber.Map{"status": "NONE"})
	}

	// Override status with the one from CheckSubscriptionAccess (which might have just updated it to EXPIRED)
	sub.Status = status

	fmt.Printf("[DEBUG] GetStatus - Business: %d, Active: %v, Status: %s\n", businessID, active, status)

	var modules []BusinessModule
	// Find active modules. If subscription is expired, CheckSubscriptionAccess doesn't auto-expire modules,
	// but HasModule checks expiration dates.
	// Ideally, we should filter here too or let the frontend see them but know sub is expired.
	// For now, return what's in DB.
	sc.db.Where("business_id = ? AND is_active = ?", businessID, true).Find(&modules)

	return c.JSON(fiber.Map{
		"subscription": sub,
		"modules":      modules,
	})
}

// Subscribe processes a new subscription payment via Paystack
// @Summary      Subscribe to a plan
// @Description  Verify a Paystack transaction and activate a subscription plan
// @Tags         Subscription
// @Accept       json
// @Produce      json
// @Param        payload  body      object  true  "Plan details and Paystack reference"
// @Success      200      {object}  Subscription
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /subscription/subscribe [post]
func (sc *SubscriptionController) Subscribe(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)

	var req struct {
		PlanType   PlanType     `json:"plan_type"`
		Modules    []ModuleType `json:"modules"`     // Optional individual modules
		BundleCode string       `json:"bundle_code"` // Optional bundle code
		Reference  string       `json:"reference"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Verify payment with Paystack
	verification, err := sc.paystack.VerifyTransaction(req.Reference)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Payment verification failed",
			"details": err.Error(),
		})
	}

	// 2. Calculate Total Amount
	var selectedPlan *SubscriptionPlan
	for _, p := range AvailablePlans {
		if p.Type == req.PlanType {
			selectedPlan = &p
			break
		}
	}

	if selectedPlan == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid plan type"})
	}

	// Base plan price
	totalPrice := selectedPlan.Price

	// Add module prices scaled to plan duration
	// Duration multiplier based on 30-day month
	monthMultiplier := float64(selectedPlan.DurationDays) / 30.0

	// Track modules that are part of a bundle to avoid double charging
	bundleModules := make(map[ModuleType]bool)
	var activeBundle *ModuleBundle

	if req.BundleCode != "" {
		for _, b := range AvailableBundles {
			if b.Code == req.BundleCode {
				activeBundle = &b
				totalPrice += b.Price * monthMultiplier
				for _, m := range b.Modules {
					bundleModules[m] = true
				}
				break
			}
		}
	}

	// Add individual modules (only if not already in bundle)
	for _, modType := range req.Modules {
		if bundleModules[modType] {
			continue
		}
		for _, m := range AvailableModules {
			if m.Type == modType {
				totalPrice += m.Price * monthMultiplier
				break
			}
		}
	}

	// Calculate expected price with optional discount
	promoCode := c.Query("promo_code")
	if promoCode != "" {
		var promo PromoCode
		if err := sc.db.Where("code = ? AND active = true", promoCode).First(&promo).Error; err == nil {
			// Validate promo
			if time.Now().Before(promo.ExpiryDate) && (promo.MaxUses == 0 || promo.UsedCount < promo.MaxUses) {
				discountAmount := (promo.DiscountPercentage / 100) * totalPrice
				totalPrice -= discountAmount
			}
		}
	}

	// Paystack amount is in kobo, our price is in Naira
	expectedKobo := totalPrice * 100
	if verification.Data.Amount < expectedKobo-1 {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Insufficient payment amount. Expected %v, got %v", expectedKobo, verification.Data.Amount),
		})
	}

	// 3. Create/Update subscription
	sub, err := CreateSubscription(
		sc.db,
		businessID,
		req.PlanType,
		"PAYSTACK",
		req.Reference,
		verification.Data.Amount/100,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 4. Activate Modules
	// Combine individual modules and bundle modules
	finalModules := req.Modules
	if activeBundle != nil {
		finalModules = append(finalModules, activeBundle.Modules...)
	}

	// Use a map for deduplication
	processed := make(map[ModuleType]bool)
	for _, modType := range finalModules {
		if processed[modType] {
			continue
		}
		processed[modType] = true

		var busMod BusinessModule
		err := sc.db.Where("business_id = ? AND module = ?", businessID, modType).First(&busMod).Error

		expiry := sub.EndDate
		if err == gorm.ErrRecordNotFound {
			sc.db.Create(&BusinessModule{
				BusinessID: businessID,
				Module:     modType,
				IsActive:   true,
				ExpiryDate: &expiry,
			})
		} else {
			sc.db.Model(&busMod).Updates(map[string]interface{}{
				"is_active":   true,
				"expiry_date": expiry,
			})
		}
	}

	// Increment promo use count if applicable
	if promoCode != "" {
		sc.db.Model(&PromoCode{}).Where("code = ?", promoCode).UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))
	}

	// 5. Update Business Subscription status
	if err := sc.db.Table("businesses").Where("id = ?", businessID).Updates(map[string]interface{}{
		"subscription_status": string(StatusActive),
		"subscription_expiry": sub.EndDate,
	}).Error; err != nil {
		fmt.Printf("Error updating business subscription status: %v\n", err)
	}

	return c.JSON(sub)
}

// ValidatePromoCode checks if a promo code is valid and returns its discount percentage
// @Summary      Validate promo code
// @Description  Check if a promo code is active, not expired, and has usage remaining
// @Tags         Subscription
// @Param        code  query     string  true  "Promo code to validate"
// @Produce      json
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Security     BearerAuth
// @Router       /subscription/promo/validate [get]
func (sc *SubscriptionController) ValidatePromoCode(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Promo code is required"})
	}

	var promo PromoCode
	if err := sc.db.Where("code = ? AND active = true", code).First(&promo).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid promo code"})
	}

	if time.Now().After(promo.ExpiryDate) {
		return c.Status(400).JSON(fiber.Map{"error": "Promo code has expired"})
	}

	if promo.MaxUses > 0 && promo.UsedCount >= promo.MaxUses {
		return c.Status(400).JSON(fiber.Map{"error": "Promo code has reached maximum uses"})
	}

	return c.JSON(fiber.Map{
		"success":             true,
		"discount_percentage": promo.DiscountPercentage,
	})
}

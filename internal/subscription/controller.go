package subscription

import (
	"fmt"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/pkg/paystack"

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
	sub, err := GetSubscriptionStatus(sc.db, businessID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if sub == nil {
		return c.JSON(fiber.Map{"status": "NONE"})
	}

	return c.JSON(sub)
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
		PlanType  PlanType `json:"plan_type"`
		Reference string   `json:"reference"`
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

	// 2. Validate amount matches plan (convert kobo/pesos if necessary)
	// Paystack returns amount in kobo (multiply by 100)
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

	// Paystack amount is in kobo, our price is in Naira
	expectedKobo := selectedPlan.Price * 100
	if verification.Data.Amount < expectedKobo {
		return c.Status(400).JSON(fiber.Map{
			"error": "Insufficient payment amount",
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

	// 4. Update Business Subscription status
	if err := sc.db.Model(&business.Business{}).Where("id = ?", businessID).Updates(map[string]interface{}{
		"subscription_status": string(StatusActive),
		"subscription_expiry": sub.EndDate,
	}).Error; err != nil {
		// Log error but subscription is created
		fmt.Printf("Error updating business subscription status: %v\n", err)
	}

	return c.JSON(sub)
}

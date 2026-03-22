package reconciliation

import (
	// "fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ReconciliationController struct {
	service *ReconciliationService
}

func NewReconciliationController(db *gorm.DB) *ReconciliationController {
	svc := NewReconciliationService(db)
	// Register default providers (Need keys from config in production)
	svc.RegisterProvider(&OPayProvider{SecretKey: "OPAY_PROD_SECRET"}) 
	svc.RegisterProvider(&MonnifyProvider{SecretKey: "MONNIFY_SECRET"})
	svc.RegisterProvider(&PalmPayProvider{SecretKey: "PALMPAY_SECRET"})
	svc.RegisterProvider(&TransferProvider{AuthToken: "INTERNAL_TRANSFER_TOKEN"})
	
	return &ReconciliationController{service: svc}
}

func (ctrl *ReconciliationController) HandleWebhook(c *fiber.Ctx) error {
	provider := c.Params("provider")
	payload := c.Body()
	
	headers := make(map[string]string)
	c.Context().Request.Header.VisitAll(func(key, value []byte) {
		headers[strings.ToLower(string(key))] = string(value)
	})

	err := ctrl.service.HandleWebhook(provider, payload, headers)
	if err != nil {
		log.Printf("[Webhook] Error handling %s: %v", provider, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "received",
	})
}

func (ctrl *ReconciliationController) GetStatus(c *fiber.Ctx) error {
	ref := c.Params("reference")
	status, err := ctrl.service.GetPaymentStatus(ref)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}

	return c.JSON(fiber.Map{
		"reference": ref,
		"status":    status,
	})
}

func (ctrl *ReconciliationController) GetSummary(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)
	summary, err := ctrl.service.GetSummary(businessID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(summary)
}

func (ctrl *ReconciliationController) ListPayments(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)
	payments, err := ctrl.service.GetPayments(businessID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(payments)
}

func (ctrl *ReconciliationController) ListLogs(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)
	logs, err := ctrl.service.GetLogs(businessID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(logs)
}

func (ctrl *ReconciliationController) GetSettlement(c *fiber.Ctx) error {
	businessID := c.Locals("business_id").(uint)
	dateStr := c.Query("date")
	
	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	settlement, err := ctrl.service.GenerateDailySettlement(businessID, date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(settlement)
}

func (ctrl *ReconciliationController) ManuallyVerify(c *fiber.Ctx) error {
	var body struct {
		PaymentID uint   `json:"payment_id"`
		Reason    string `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := ctrl.service.ManuallyVerify(body.PaymentID, body.Reason); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Payment verified manually"})
}

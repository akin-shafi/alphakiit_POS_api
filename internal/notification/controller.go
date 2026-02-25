package notification

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type NotificationController struct {
	service *NotificationService
}

func NewNotificationController(service *NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

func RegisterNotificationRoutes(router fiber.Router, db *gorm.DB) {
	service := GetDefaultService(db)
	controller := NewNotificationController(service)

	group := router.Group("/notifications")
	group.Post("/test-email", controller.TestEmail)
	group.Post("/tokens", controller.RegisterToken)
}

func (c *NotificationController) RegisterToken(ctx *fiber.Ctx) error {
	var req struct {
		Token      string `json:"token" validate:"required"`
		DeviceType string `json:"device_type"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID := ctx.Locals("user_id").(uint)
	businessID := ctx.Locals("current_business_id").(uint)

	var deviceToken DeviceToken
	err := c.service.db.Where("token = ?", req.Token).First(&deviceToken).Error
	if err == nil {
		// Update existing
		deviceToken.UserID = userID
		deviceToken.BusinessID = businessID
		deviceToken.DeviceType = req.DeviceType
		deviceToken.LastUsed = time.Now()
		c.service.db.Save(&deviceToken)
	} else {
		// Create new
		deviceToken = DeviceToken{
			UserID:     userID,
			BusinessID: businessID,
			Token:      req.Token,
			DeviceType: req.DeviceType,
			LastUsed:   time.Now(),
		}
		if err := c.service.db.Create(&deviceToken).Error; err != nil {
			return ctx.Status(500).JSON(fiber.Map{"error": "failed to register token"})
		}
	}

	return ctx.JSON(fiber.Map{"success": true, "message": "token registered successfully"})
}

func (c *NotificationController) TestEmail(ctx *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" validate:"required"`
		Type  string `json:"type" validate:"required"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	title := "Test Email Alert"
	message := "This is a test notification from your POS system."

	switch req.Type {
	case "security_alert":
		title = "Security Alert Test"
		message = "This is a test of the security alert system. Unauthorized access was NOT detected."
	case "shift_variance":
		title = "Shift Variance Test"
		message = "Shift #TEST closed by TEST USER has a shortage of NGN 500.00.\nExpected: NGN 1,000.00 | Actual: NGN 500.00"
	case "daily_report":
		title = "Daily Business Report Test"
		message = "Your daily business report summary for today.\nTotal Sales: NGN 50,000.00\nStock Alerts: 3 items low"
	}

	// For testing, we use SendSecurityEmail directly on the emailSender to any address passed
	err := c.service.emailSender.SendSecurityEmail(req.Email, "User", title, message)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("failed to send email: %v", err)})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("test email of type %s sent to %s", req.Type, req.Email),
	})
}

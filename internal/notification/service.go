package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"pos-fiber-app/internal/email"
	"strings"

	"gorm.io/gorm"
)

type NotificationService struct {
	db          *gorm.DB
	emailSender *email.Sender
}

func NewNotificationService(db *gorm.DB, emailSender *email.Sender) *NotificationService {
	return &NotificationService{
		db:          db,
		emailSender: emailSender,
	}
}

// GetDefaultService returns a new service with production/env config
func GetDefaultService(db *gorm.DB) *NotificationService {
	emailCfg := email.LoadConfig()
	emailSender := email.NewSender(emailCfg)
	return NewNotificationService(db, emailSender)
}

// SendVoidAlert sends an alert when a sale is voided
func (n *NotificationService) SendVoidAlert(businessID uint, saleID uint, cashierName, reason string, amount float64, currency string) {
	title := "Sale Voided"
	message := fmt.Sprintf(
		"Sale #%d worth %s%s was voided by %s.\nReason: %s",
		saleID, currency, formatCurrency(amount), cashierName, reason,
	)
	n.SendSecurityAlert(businessID, title, message)
}

// SendShiftVarianceAlert sends an alert if a shift has a high variance
func (n *NotificationService) SendShiftVarianceAlert(businessID uint, shiftID uint, cashierName string, expected, actual, variance float64, currency string) {
	title := "Shift Cash Discrepancy"
	status := "shortage"
	if variance > 0 {
		status = "surplus"
	} else if variance < 0 {
		variance = -variance // make positive for display
	}
	message := fmt.Sprintf(
		"Shift #%d closed by %s has a %s of %s%s.\nExpected: %s%s | Actual: %s%s",
		shiftID, cashierName, status, currency, formatCurrency(variance),
		currency, formatCurrency(expected), currency, formatCurrency(actual),
	)
	n.SendSecurityAlert(businessID, title, message)
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// SendShiftClosedReport sends a standard report when a shift is closed normally
func (n *NotificationService) SendShiftClosedReport(businessID uint, shiftID uint, cashierName string, totalSales float64, count int, currency string) {
	title := "Shift Closed Successfully"
	message := fmt.Sprintf(
		"Shift #%d has been closed by %s.\nTotal Sales: %s%s\nTransactions: %d",
		shiftID, cashierName, currency, formatCurrency(totalSales), count,
	)
	n.SendSecurityAlert(businessID, title, message)
}

// SendSecurityAlert sends an alert to the business owner about sensitive actions
func (n *NotificationService) SendSecurityAlert(businessID uint, title, message string) {
	// 1. Find the owner of this business
	var businessOwner struct {
		Email     string
		FirstName string
		Phone     string
	}
	var biz struct {
		TenantID string
	}

	// Get business tenant
	if err := n.db.Table("businesses").Select("tenant_id").Where("id = ?", businessID).Scan(&biz).Error; err != nil {
		fmt.Printf("Notification Error: Failed to find business %d: %v\n", businessID, err)
		return
	}

	// Get owner of that tenant
	if err := n.db.Table("users").Where("tenant_id = ? AND role = ?", biz.TenantID, "OWNER").First(&businessOwner).Error; err != nil {
		fmt.Printf("Notification Error: Failed to find owner for business %d: %v\n", businessID, err)
		return
	}

	// 2. Send via Email
	go func() {
		if err := n.emailSender.SendSecurityEmail(businessOwner.Email, businessOwner.FirstName, title, message); err != nil {
			fmt.Printf("Email Alert Error: %v\n", err)
		}
	}()

	// 3. Send via WhatsApp (Only if subscribed to WHATSAPP_ALERTS module and enabled)
	go func() {
		// a. Check if module is active
		var moduleCount int64
		if err := n.db.Table("business_modules").
			Where("business_id = ? AND module = ? AND is_active = ?", businessID, "WHATSAPP_ALERTS", true).
			Count(&moduleCount).Error; err != nil || moduleCount == 0 {
			// Module not active, skip WhatsApp
			fmt.Printf("WhatsApp Alert Skipped: Module WHATSAPP_ALERTS not active for business %d\n", businessID)
			return
		}

		// b. Check if business has enabled it and has a number
		var bizSettings struct {
			WhatsAppEnabled bool
			WhatsAppNumber  string
		}
		if err := n.db.Table("businesses").Select("whats_app_enabled, whats_app_number").Where("id = ?", businessID).Scan(&bizSettings).Error; err != nil {
			return
		}

		if !bizSettings.WhatsAppEnabled {
			fmt.Printf("WhatsApp Alert Skipped: Feature disabled by business %d\n", businessID)
			return
		}

		targetNumber := bizSettings.WhatsAppNumber
		if targetNumber == "" {
			targetNumber = businessOwner.Phone // Fallback to owner phone
		}

		if targetNumber == "" {
			fmt.Printf("WhatsApp Alert Skipped: No phone number available for business %d\n", businessID)
			return
		}

		// c. Send via Twilio API
		accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
		authToken := os.Getenv("TWILIO_AUTH_TOKEN")
		fromNumber := os.Getenv("TWILIO_FROM_NUMBER") // or WHATSAPP_SANDBOX_NUMBER if testing

		if accountSid == "" || authToken == "" {
			fmt.Println("WhatsApp Alert Error: Twilio credentials missing in .env")
			return
		}

		// Using Twilio API directly to avoid external dependency issues if library not installed
		apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)

		// Ensure number has whatsapp: prefix for Twilio WhatsApp
		to := targetNumber
		if !strings.HasPrefix(to, "whatsapp:") {
			to = "whatsapp:" + to
		}
		from := fromNumber
		if !strings.HasPrefix(from, "whatsapp:") {
			// If env var is just a number, prepend whatsapp:
			// If env var is WHATSAPP_SANDBOX_NUMBER, it usually already has it.
			// Let's assume user configured it correctly or we prepend if missing
			if os.Getenv("WHATSAPP_SANDBOX_NUMBER") != "" {
				from = os.Getenv("WHATSAPP_SANDBOX_NUMBER")
			} else {
				from = "whatsapp:" + from
			}
		}

		v := url.Values{}
		v.Set("To", to)
		v.Set("From", from)
		v.Set("Body", fmt.Sprintf("*%s*\n\n%s", title, message))

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(v.Encode()))
		if err != nil {
			fmt.Printf("WhatsApp Alert Error creating request: %v\n", err)
			return
		}

		req.SetBasicAuth(accountSid, authToken)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("WhatsApp Alert Network Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("WhatsApp Alert Sent to %s: %s\n", to, title)
		} else {
			fmt.Printf("WhatsApp Alert Failed (Status %d)\n", resp.StatusCode)
		}
	}()
}

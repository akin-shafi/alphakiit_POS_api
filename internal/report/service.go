package report

import (
	"fmt"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/notification"
	"pos-fiber-app/internal/sale"
	"pos-fiber-app/internal/shift"
	"time"

	"gorm.io/gorm"
)

type ReportService struct {
	db       *gorm.DB
	notifier *notification.NotificationService
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{
		db:       db,
		notifier: notification.GetDefaultService(db),
	}
}

func (s *ReportService) GenerateAndSendDailyReport(businessID uint) error {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfToday := startOfToday.Add(24 * time.Hour)

	// 1. Check Activity
	var shiftCount int64
	s.db.Model(&shift.Shift{}).Where("business_id = ? AND (start_time >= ? OR end_time >= ?)", businessID, startOfToday, startOfToday).Count(&shiftCount)

	if shiftCount == 0 {
		var saleCount int64
		s.db.Model(&sale.Sale{}).Where("business_id = ? AND sale_date >= ? AND sale_date < ? AND status = ?", businessID, startOfToday, endOfToday, sale.StatusCompleted).Count(&saleCount)
		if saleCount == 0 {
			return nil // No activity
		}
	}

	// 2. Aggregate Data
	dailySales, _ := sale.GenerateDailyReport(s.db, businessID, now.Format("2006-01-02"))
	lowStock, _ := inventory.ListLowStockItems(s.db, businessID, 0)

	// Security Alerts
	var securityAlerts []string
	var voids []sale.SaleActivityLog
	s.db.Where("business_id = ? AND action_type = ? AND created_at >= ?", businessID, sale.ActionVoided, startOfToday).Find(&voids)
	for _, v := range voids {
		securityAlerts = append(securityAlerts, fmt.Sprintf("Sale #%d was voided.", v.SaleID))
	}

	var varianceShifts []shift.Shift
	s.db.Where("business_id = ? AND status = ? AND end_time >= ? AND cash_variance != 0", businessID, "closed", startOfToday).Find(&varianceShifts)
	for _, sh := range varianceShifts {
		securityAlerts = append(securityAlerts, fmt.Sprintf("Shift #%d closed with variance: %.2f", sh.ID, sh.CashVariance))
	}

	// 3. Prepare Email
	var biz struct {
		Name     string
		Currency string
		TenantID string
	}
	s.db.Table("businesses").Where("id = ?", businessID).Select("name, currency, tenant_id").Scan(&biz)

	var owner struct {
		Email     string
		FirstName string
	}
	s.db.Table("users").Where("tenant_id = ? AND role = ?", biz.TenantID, "OWNER").Select("email, first_name").Scan(&owner)

	if owner.Email == "" {
		return fmt.Errorf("owner email not found")
	}

	emailLowStock := make([]email.EmailInventoryItem, len(lowStock))
	for i, item := range lowStock {
		// Need product name
		var pName string
		s.db.Table("products").Where("id = ?", item.ProductID).Select("name").Scan(&pName)
		emailLowStock[i] = email.EmailInventoryItem{Name: pName, Stock: item.CurrentStock}
	}

	data := email.EmailData{
		Name:             owner.FirstName,
		BusinessName:     biz.Name,
		Currency:         biz.Currency,
		Date:             now.Format("January 02, 2006"),
		TotalSales:       fmt.Sprintf("%.2f", dailySales.TotalSales),
		TransactionCount: dailySales.TotalTransactions,
		CashSales:        fmt.Sprintf("%.2f", dailySales.CashSales),
		CardSales:        fmt.Sprintf("%.2f", dailySales.CardSales),
		TransferSales:    fmt.Sprintf("%.2f", dailySales.TransferSales),
		LowStockItems:    emailLowStock,
		SecurityAlerts:   securityAlerts,
	}

	emailCfg := email.LoadConfig()
	sender := email.NewSender(emailCfg)
	return sender.SendDailyReport(owner.Email, data)
}

func (s *ReportService) TriggerWeeklyAuditReminders() error {
	// 1. Fetch businesses that need audit reminders
	fmt.Println("Triggering weekly audit reminders...")
	var businesses []struct {
		ID       uint
		Name     string
		TenantID string
	}
	s.db.Table("businesses").Where("reporting_enabled = ?", true).Select("id, name, tenant_id").Find(&businesses)

	for _, b := range businesses {
		// Fetch owner
		var owner struct {
			Email     string
			FirstName string
		}
		s.db.Table("users").Where("tenant_id = ? AND role = ?", b.TenantID, "OWNER").Select("email, first_name").Scan(&owner)

		if owner.Email == "" {
			continue
		}

		// Prepare data for audit reminder
		s.notifier.SendSecurityAlert(b.ID, "Physical Stock Audit Reminder",
			fmt.Sprintf("Hello %s, it's time for the weekly headcount of your stock at %s. Please perform a physical audit to ensure inventory accuracy.", owner.FirstName, b.Name))
	}
	return nil
}

func (s *ReportService) TriggerMonthlyFinancialReports() error {
	fmt.Println("Triggering monthly financial reports...")

	now := time.Now()
	// Get start and end of previous month
	lastMonth := now.AddDate(0, -1, 0)
	startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfLastMonth := startOfLastMonth.AddDate(0, 1, 0).Add(-time.Second)

	var businesses []struct {
		ID       uint
		Name     string
		TenantID string
		Currency string
	}
	s.db.Table("businesses").Where("reporting_enabled = ?", true).Select("id, name, tenant_id, currency").Find(&businesses)

	monthName := lastMonth.Format("January 2006")

	for _, b := range businesses {
		// 1. Fetch Month Stats
		monthlyStats, err := sale.GenerateSalesReport(s.db, b.ID, startOfLastMonth.Format("2006-01-02"), endOfLastMonth.Format("2006-01-02"), "")
		if err != nil {
			fmt.Printf("Error generating monthly stats for %s: %v\n", b.Name, err)
			continue
		}

		if monthlyStats.TotalTransactions == 0 {
			continue // Skip if no activity
		}

		// Fetch owner
		var owner struct {
			Email     string
			FirstName string
		}
		s.db.Table("users").Where("tenant_id = ? AND role = ?", b.TenantID, "OWNER").Select("email, first_name").Scan(&owner)

		if owner.Email == "" {
			continue
		}

		// 2. Prepare Data
		data := email.EmailData{
			Name:             owner.FirstName,
			BusinessName:     b.Name,
			Currency:         b.Currency,
			Month:            monthName,
			TotalSales:       fmt.Sprintf("%.2f", monthlyStats.TotalSales),
			TotalCost:        fmt.Sprintf("%.2f", monthlyStats.TotalCost),
			Expenses:         fmt.Sprintf("%.2f", monthlyStats.TotalExpenses),
			NetProfit:        fmt.Sprintf("%.2f", monthlyStats.NetProfit),
			Profit:           fmt.Sprintf("%.2f", monthlyStats.TotalProfit), // Gross Profit
			TransactionCount: monthlyStats.TotalTransactions,
			Message:          fmt.Sprintf("%.2f", monthlyStats.AverageSale), // Using message as placeholder for Avg Sale
		}

		// 3. Send Email
		emailCfg := email.LoadConfig()
		sender := email.NewSender(emailCfg)
		if err := sender.SendMonthlyReport(owner.Email, data); err != nil {
			fmt.Printf("Error sending monthly report to %s: %v\n", owner.Email, err)
		} else {
			fmt.Printf("Successfully sent monthly report for %s to %s\n", b.Name, owner.Email)
		}
	}
	return nil
}

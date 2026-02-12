package notification

import (
	"fmt"
	"pos-fiber-app/internal/email"
	"time"
)

func (n *NotificationService) GenerateAndSendDailyReport(businessID uint) error {
	// 1. Check for shift activity today
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfToday := startOfToday.Add(24 * time.Hour)

	var shiftCount int64
	n.db.Table("shifts").Where("business_id = ? AND (start_time >= ? OR end_time >= ?)", businessID, startOfToday, startOfToday).Count(&shiftCount)

	if shiftCount == 0 {
		var saleCount int64
		n.db.Table("sales").Where("business_id = ? AND sale_date >= ? AND sale_date < ? AND status = ?", businessID, startOfToday, endOfToday, "COMPLETED").Count(&saleCount)
		if saleCount == 0 {
			fmt.Printf("Daily Report Skipped: No activity for business %d today\n", businessID)
			return nil
		}
	}

	// 2. Gather Data
	// a. Sales Summary
	type salesSummary struct {
		TotalSales        float64
		TotalTransactions int
		CashSales         float64
		CardSales         float64
		TransferSales     float64
	}
	var report salesSummary
	n.db.Table("sales").
		Where("business_id = ? AND sale_date >= ? AND sale_date < ? AND status = ?", businessID, startOfToday, endOfToday, "COMPLETED").
		Select("SUM(total) as total_sales, COUNT(*) as total_transactions, " +
			"SUM(CASE WHEN payment_method = 'CASH' THEN total ELSE 0 END) as cash_sales, " +
			"SUM(CASE WHEN payment_method = 'CARD' THEN total ELSE 0 END) as card_sales, " +
			"SUM(CASE WHEN payment_method = 'TRANSFER' THEN total ELSE 0 END) as transfer_sales").
		Scan(&report)

	// b. Low Stock Items (using raw query to avoid inventory package import)
	var lowStock []struct {
		Name  string
		Stock int
	}
	n.db.Table("inventories").
		Select("products.name, inventories.current_stock as stock").
		Joins("JOIN products ON products.id = inventories.product_id").
		Where("inventories.business_id = ? AND inventories.current_stock <= inventories.low_stock_alert", businessID).
		Scan(&lowStock)

	// c. Security Alerts (Voids)
	type activityLog struct {
		SaleID uint
	}
	var voids []activityLog
	n.db.Table("sale_activity_logs").Where("business_id = ? AND action_type = ? AND created_at >= ?", businessID, "voided", startOfToday).Find(&voids)

	// d. Security Alerts (High Variance Shifts)
	type shiftInfo struct {
		ID           uint
		CashVariance float64
	}
	var varianceShifts []shiftInfo
	n.db.Table("shifts").Where("business_id = ? AND status = ? AND end_time >= ? AND cash_variance != 0", businessID, "closed", startOfToday).Find(&varianceShifts)

	// 3. Format Alerts
	var alerts []string
	for _, v := range voids {
		alerts = append(alerts, fmt.Sprintf("Sale #%d was voided.", v.SaleID))
	}
	for _, s := range varianceShifts {
		alerts = append(alerts, fmt.Sprintf("Shift #%d had a variance of %.2f", s.ID, s.CashVariance))
	}

	// 4. Send Email
	var biz struct {
		Name     string
		Currency string
		TenantID string
	}
	n.db.Table("businesses").Where("id = ?", businessID).Select("name, currency, tenant_id").Scan(&biz)

	var owner struct {
		Email     string
		FirstName string
	}
	n.db.Table("users").Where("tenant_id = ? AND role = ?", biz.TenantID, "OWNER").First(&owner)

	if owner.Email == "" {
		return fmt.Errorf("owner email not found for business %d", businessID)
	}

	// Convert raw inventory to email format
	emailLowStock := make([]email.EmailInventoryItem, len(lowStock))
	for i, item := range lowStock {
		emailLowStock[i] = email.EmailInventoryItem{Name: item.Name, Stock: item.Stock}
	}

	emailData := email.EmailData{
		Name:             owner.FirstName,
		BusinessName:     biz.Name,
		Currency:         biz.Currency,
		Date:             now.Format("January 02, 2006"),
		TotalSales:       fmt.Sprintf("%.2f", report.TotalSales),
		TransactionCount: report.TotalTransactions,
		CashSales:        fmt.Sprintf("%.2f", report.CashSales),
		CardSales:        fmt.Sprintf("%.2f", report.CardSales),
		TransferSales:    fmt.Sprintf("%.2f", report.TransferSales),
		LowStockItems:    emailLowStock,
		SecurityAlerts:   alerts,
	}

	return n.emailSender.SendDailyReport(owner.Email, emailData)
}

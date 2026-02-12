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

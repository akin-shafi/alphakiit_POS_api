package sale

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TaxReportItem represents a row in the tax report CSV
type TaxReportItem struct {
	Date          string
	InvoiceNumber string
	Customer      string
	TaxID         string
	Subtotal      float64
	VAT           float64
	Total         float64
	PaymentMethod string
	Status        string
}

// GenerateTaxReport generates a CSV report for VAT/FIRS filings
func GenerateTaxReport(db *gorm.DB, businessID uint, startDate, endDate string) ([]byte, error) {
	// 1. Fetch sales within range
	var sales []Sale
	query := db.Where("business_id = ? AND status = ?", businessID, StatusCompleted)

	if startDate != "" && endDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		// Set end to end of day
		end = end.Add(24 * time.Hour).Add(-1 * time.Second)
		query = query.Where("created_at BETWEEN ? AND ?", start, end)
	}

	if err := query.Find(&sales).Error; err != nil {
		return nil, err
	}

	// 2. Prepare CSV Data
	var reportData []TaxReportItem
	for _, s := range sales {
		// Calculate VAT (Assuming inclusive or exclusive logic, here we simplify)
		// For this example, let's assume specific tax logic based on business settings or fixed rate
		// In a real app, you'd fetch tax settings. Here we'll show raw totals and calculate example 7.5% VAT from Total if needed,
		// or just list the tax amount if stored.
		// For now, let's use the stored Total and assume a standard extraction or 0 if not recorded.

		// If tax amount isn't explicitly stored, we might calculate it.
		// Let's assume 7.5% VAT is included in Total for reporting purposes if not separated.
		// VAT = Total - (Total / 1.075)
		vatAmount := s.Total - (s.Total / 1.075)
		subtotal := s.Total - vatAmount

		reportData = append(reportData, TaxReportItem{
			Date:          s.CreatedAt.Format("2006-01-02 15:04:05"),
			InvoiceNumber: fmt.Sprintf("INV-%d", s.ID),
			Customer:      s.CustomerName, // Assuming this field exists or needs to be joined
			TaxID:         "N/A",          // Placeholder for Customer Tax ID if we track it
			Subtotal:      subtotal,
			VAT:           vatAmount,
			Total:         s.Total,
			PaymentMethod: s.PaymentMethod,
			Status:        string(s.Status),
		})
	}

	// 3. Write to CSV
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	// Header
	if err := w.Write([]string{"Date", "Invoice No", "Customer", "Tax ID", "Subtotal", "VAT (7.5%)", "Total", "Payment Method", "Status"}); err != nil {
		return nil, err
	}

	// Rows
	for _, r := range reportData {
		record := []string{
			r.Date,
			r.InvoiceNumber,
			r.Customer,
			r.TaxID,
			fmt.Sprintf("%.2f", r.Subtotal),
			fmt.Sprintf("%.2f", r.VAT),
			fmt.Sprintf("%.2f", r.Total),
			r.PaymentMethod,
			r.Status,
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return b.Bytes(), nil
}

// GetAuditTrail fetches rich activity logs with optional filtering
func GetAuditTrail(db *gorm.DB, businessID uint, date string, actionType string) ([]SaleActivityLogWithUser, error) {
	query := db.Table("sale_activity_logs").
		Select("sale_activity_logs.*, users.first_name || ' ' || users.last_name as user_name").
		Joins("LEFT JOIN users ON users.id = sale_activity_logs.performed_by").
		Where("sale_activity_logs.business_id = ?", businessID)

	if date != "" {
		start, _ := time.Parse("2006-01-02", date)
		end := start.Add(24 * time.Hour).Add(-1 * time.Second)
		query = query.Where("sale_activity_logs.created_at BETWEEN ? AND ?", start, end)
	}

	if actionType != "" {
		query = query.Where("sale_activity_logs.action_type = ?", actionType)
	}

	// Order by latest first for "Replay" (user can reverse in UI or we order asc)
	// For "Replay" usually chronological is better (ASC), but Audit Log usually DESC.
	// Let's return DESC and UI can reverse or we support sort param.
	// Defaulting to DESC as per standard log views.
	query = query.Order("sale_activity_logs.created_at DESC")

	var logs []SaleActivityLogWithUser
	if err := query.Scan(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetAuditStats returns summary for the day
func GetAuditStats(db *gorm.DB, businessID uint, date string) (map[string]int64, error) {
	var stats = make(map[string]int64)

	start, _ := time.Parse("2006-01-02", date)
	end := start.Add(24 * time.Hour).Add(-1 * time.Second)

	// Count Voids
	var voidCount int64
	db.Model(&SaleActivityLog{}).Where("business_id = ? AND action_type = ? AND created_at BETWEEN ? AND ?", businessID, ActionVoided, start, end).Count(&voidCount)
	stats["voids"] = voidCount

	// Count Logins (we need to track logins in activity log or separate table. Assuming basic sale actions for now, we might not have logins here yet)
	// If logins are not in sale_activity_logs, we might skip or check another table.
	// For now, let's just return what we have.

	return stats, nil
}

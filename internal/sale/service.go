// internal/sale/service.go
package sale

import (
	"errors"
	"fmt"
	"time"

	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/product"

	"gorm.io/gorm"
)

type AddItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type CompleteSaleRequest struct {
	PaymentMethod string  `json:"payment_method" validate:"required"`
	AmountPaid    float64 `json:"amount_paid" validate:"required,gte=0"`
	Discount      float64 `json:"discount" validate:"gte=0"`
}

type VoidSaleRequest struct {
	Reason string `json:"reason" validate:"required,min=5"`
}

type SaleFilters struct {
	Status SaleStatus
	From   string
	To     string
}

type SaleResult struct {
	Sale  *Sale
	Items []SaleItem
}

type SaleReceipt struct {
	Sale        *Sale      `json:"sale"`
	Items       []SaleItem `json:"items"`
	Change      float64    `json:"change"`
	ReceiptNo   string     `json:"receipt_no"`
	GeneratedAt time.Time  `json:"generated_at"`
}

type DailyReport struct {
	Date              string  `json:"date"`
	TotalSales        float64 `json:"total_sales"`
	TotalTransactions int     `json:"total_transactions"`
	CashSales         float64 `json:"cash_sales"`
	CardSales         float64 `json:"card_sales"`
	TransferSales     float64 `json:"transfer_sales"`
	AverageSale       float64 `json:"average_sale"`
}

// CreateDraft starts a new sale
func CreateDraft(db *gorm.DB, businessID, cashierID uint) (*Sale, error) {
	sale := &Sale{
		BusinessID: businessID,
		Status:     StatusDraft,
		CashierID:  cashierID,
		SaleDate:   time.Now(),
	}

	return sale, db.Create(sale).Error
}

// AddItemToSale adds or updates quantity of a product in a sale
func AddItemToSale(db *gorm.DB, saleID, businessID uint, productID uint, qty int) (*SaleResult, error) {
	var sale Sale
	if err := db.First(&sale, "id = ? AND business_id = ? AND status IN ?", saleID, businessID, []SaleStatus{StatusDraft, StatusHeld}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sale not found or not editable")
		}
		return nil, err
	}

	var prod product.Product
	if err := db.First(&prod, "id = ? AND business_id = ?", productID, businessID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Check stock
	var inv inventory.Inventory
	if err := db.First(&inv, "product_id = ? AND business_id = ?", productID, businessID).Error; err != nil || inv.CurrentStock < qty {
		return nil, errors.New("insufficient stock")
	}

	// Upsert sale item
	var item SaleItem
	db.FirstOrCreate(&item, "sale_id = ? AND product_id = ?", saleID, productID)
	item.Quantity += qty
	item.UnitPrice = prod.Price
	item.TotalPrice = float64(item.Quantity) * prod.Price
	item.ProductName = prod.Name

	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}

	// Recalculate sale totals
	if err := recalculateSaleTotals(db, &sale); err != nil {
		return nil, err
	}

	items := []SaleItem{}
	db.Where("sale_id = ?", saleID).Find(&items)

	return &SaleResult{Sale: &sale, Items: items}, nil
}

// CompleteSale finalizes sale and deducts inventory
func CompleteSale(db *gorm.DB, saleID, businessID uint, req CompleteSaleRequest) (*SaleReceipt, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var sale Sale
	if err := tx.Preload("SaleItems").First(&sale, "id = ? AND business_id = ? AND status = ?", saleID, businessID, StatusDraft).Error; err != nil {
		return nil, errors.New("sale not found or already completed")
	}

	if sale.Total-req.Discount > req.AmountPaid {
		return nil, errors.New("insufficient payment")
	}

	// Deduct inventory
	for _, item := range sale.SaleItems {
		if err := inventory.AdjustStock(tx, item.ProductID, businessID, -item.Quantity); err != nil {
			return nil, errors.New("failed to update inventory: " + err.Error())
		}
	}

	sale.Status = StatusCompleted
	sale.PaymentMethod = req.PaymentMethod
	sale.Discount = req.Discount
	sale.SyncedAt = &time.Time{} // mark as synced

	if err := tx.Save(&sale).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SaleReceipt{
		Sale:        &sale,
		Items:       sale.SaleItems,
		Change:      req.AmountPaid - (sale.Total - req.Discount),
		ReceiptNo:   generateReceiptNo(sale.ID),
		GeneratedAt: time.Now(),
	}, nil
}

func HoldSale(db *gorm.DB, saleID, businessID uint) (*Sale, error) {
	return updateSaleStatus(db, saleID, businessID, StatusHeld)
}

func VoidSale(db *gorm.DB, saleID, businessID uint, reason string) (*Sale, error) {
	tx := db.Begin()

	var sale Sale
	if err := tx.Preload("SaleItems").First(&sale, "id = ? AND business_id = ? AND status = ?", saleID, businessID, StatusCompleted).Error; err != nil {
		return nil, errors.New("sale not found or cannot be voided")
	}

	// Restock
	for _, item := range sale.SaleItems {
		inventory.AdjustStock(tx, item.ProductID, businessID, item.Quantity)
	}

	sale.Status = StatusVoided
	// Add reason field if you extend model

	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &sale, tx.Commit().Error
}

func updateSaleStatus(db *gorm.DB, saleID, businessID uint, status SaleStatus) (*Sale, error) {
	var sale Sale
	if err := db.First(&sale, "id = ? AND business_id = ?", saleID, businessID).Error; err != nil {
		return nil, err
	}
	sale.Status = status
	return &sale, db.Save(&sale).Error
}

func recalculateSaleTotals(db *gorm.DB, sale *Sale) error {
	var items []SaleItem
	if err := db.Where("sale_id = ?", sale.ID).Find(&items).Error; err != nil {
		return err
	}

	subtotal := 0.0
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	sale.Subtotal = subtotal
	sale.Total = subtotal // discount applied later

	return db.Save(sale).Error
}

func generateReceiptNo(saleID uint) string {
	return time.Now().Format("20060102") + "-" + fmt.Sprintf("%06d", saleID)
}

// RemoveItemFromSale removes a specific sale item and recalculates totals
func RemoveItemFromSale(db *gorm.DB, saleID, itemID, businessID uint) (*SaleResult, error) {
	var sale Sale
	if err := db.First(&sale, "id = ? AND business_id = ? AND status IN ?", saleID, businessID, []SaleStatus{StatusDraft, StatusHeld}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sale not found or not editable")
		}
		return nil, err
	}

	// Delete the item
	result := db.Where("id = ? AND sale_id = ?", itemID, saleID).Delete(&SaleItem{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("item not found")
	}

	// Recalculate sale totals
	if err := recalculateSaleTotals(db, &sale); err != nil {
		return nil, err
	}

	// Reload items
	var items []SaleItem
	if err := db.Where("sale_id = ?", saleID).Find(&items).Error; err != nil {
		return nil, err
	}

	return &SaleResult{Sale: &sale, Items: items}, nil
}

// ListHeldSales returns all held sales for the business (optionally filtered by cashier)
func ListHeldSales(db *gorm.DB, businessID, cashierID uint) ([]Sale, error) {
	var heldSales []Sale

	query := db.Where("business_id = ? AND status = ?", businessID, StatusHeld)

	// Optional: limit to current cashier's held sales (recommended for multi-cashier terminals)
	query = query.Where("cashier_id = ?", cashierID)

	if err := query.Preload("SaleItems").Order("updated_at DESC").Find(&heldSales).Error; err != nil {
		return nil, err
	}

	return heldSales, nil
}

// ListSales, GetSaleDetails, GenerateDailyReport implementations follow similar patterns...
func ListSales(db *gorm.DB, businessID uint, filters SaleFilters) ([]Sale, error) {
	var sales []Sale
	query := db.Where("business_id = ?", businessID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.From != "" {
		query = query.Where("sale_date >= ?", filters.From)
	}
	if filters.To != "" {
		query = query.Where("sale_date <= ?", filters.To)
	}
	if err := query.Find(&sales).Error; err != nil {
		return nil, err
	}
	return sales, nil
}

func GetSaleDetails(db *gorm.DB, saleID, businessID uint) (*SaleResult, error) {
	var sale Sale
	if err := db.First(&sale, "id = ? AND business_id = ?", saleID, businessID).Error; err != nil {
		return nil, errors.New("sale not found")
	}
	var items []SaleItem
	if err := db.Where("sale_id = ?", saleID).Find(&items).Error; err != nil {
		return nil, err
	}
	return &SaleResult{Sale: &sale, Items: items}, nil
}

// internal/sale/service.go (add this function)

func GenerateDailyReport(db *gorm.DB, businessID uint, dateStr string) (*DailyReport, error) {
	// Parse date or default to today
	var targetDate time.Time
	if dateStr == "" {
		targetDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date format, expected YYYY-MM-DD")
		}
		targetDate = parsed
	}

	// Define start and end of the day (midnight to midnight)
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Initialize report
	report := &DailyReport{
		Date: targetDate.Format("2006-01-02"),
	}

	// Query to get totals grouped by payment method
	type result struct {
		PaymentMethod     string
		TotalSales        float64
		TotalTransactions int
	}

	var results []result

	err := db.Model(&Sale{}).
		Where("business_id = ? AND sale_date >= ? AND sale_date < ? AND status = ?",
			businessID, startOfDay, endOfDay, StatusCompleted).
		Select("payment_method, SUM(total) as total_sales, COUNT(*) as total_transactions").
		Group("payment_method").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Aggregate results
	var grandTotalSales float64
	var grandTotalTransactions int

	for _, r := range results {
		grandTotalSales += r.TotalSales
		grandTotalTransactions += r.TotalTransactions

		// Map payment methods
		switch r.PaymentMethod {
		case "CASH":
			report.CashSales = r.TotalSales
		case "CARD":
			report.CardSales = r.TotalSales
		case "TRANSFER":
			report.TransferSales = r.TotalSales
		// Add more methods as needed (e.g., MOBILE_MONEY, etc.)
		default:
			// If unknown, add to cash or create Other category
			report.CashSales += r.TotalSales // fallback
		}
	}

	// Fill final fields
	report.TotalSales = grandTotalSales
	report.TotalTransactions = grandTotalTransactions

	if grandTotalTransactions > 0 {
		report.AverageSale = grandTotalSales / float64(grandTotalTransactions)
	} else {
		report.AverageSale = 0
	}

	return report, nil
}

// GenerateSalesReport generates a sales report for a given date range
// If startDate or endDate is empty, defaults to today
// startDate and endDate expected in YYYY-MM-DD format
// GenerateSalesReport generates a sales report for a date range with optional payment method filter
func GenerateSalesReport(db *gorm.DB, businessID uint, startDateStr, endDateStr, paymentMethod string) (*SalesReport, error) {
	// Parse dates (same as before)
	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		startDate = time.Now()
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, errors.New("invalid start_date format")
		}
	}

	if endDateStr == "" {
		endDate = startDate
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, errors.New("invalid end_date format")
		}
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end_date cannot be before start_date")
	}

	startOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfPeriod := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// Base query
	query := db.Model(&Sale{}).
		Where("business_id = ? AND sale_date >= ? AND sale_date <= ? AND status = ?",
			businessID, startOfPeriod, endOfPeriod, StatusCompleted)

	// Optional payment method filter
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	// Aggregate by payment method
	type result struct {
		PaymentMethod     string
		TotalSales        float64
		TotalTransactions int
	}

	var results []result

	err = query.
		Select("payment_method, SUM(total - discount) as total_sales, COUNT(*) as total_transactions").
		Group("payment_method").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Initialize report
	report := &SalesReport{
		FromDate: startOfPeriod.Format("2006-01-02"),
		ToDate:   endOfPeriod.Format("2006-01-02"),
	}

	var grandTotalSales float64
	var grandTotalTransactions int

	for _, r := range results {
		grandTotalSales += r.TotalSales
		grandTotalTransactions += r.TotalTransactions

		switch r.PaymentMethod {
		case "CASH":
			report.CashSales = r.TotalSales
			report.CashTransactions = r.TotalTransactions
		case "CARD":
			report.CardSales = r.TotalSales
			report.CardTransactions = r.TotalTransactions
		case "TRANSFER":
			report.TransferSales = r.TotalSales
			report.TransferTransactions = r.TotalTransactions
		case "MOBILE_MONEY":
			report.MobileMoneySales = r.TotalSales
			report.MobileMoneyTransactions = r.TotalTransactions
		default:
			report.OtherSales += r.TotalSales
			report.OtherTransactions += r.TotalTransactions
		}
	}

	report.TotalSales = grandTotalSales
	report.TotalTransactions = grandTotalTransactions

	if grandTotalTransactions > 0 {
		report.AverageSale = grandTotalSales / float64(grandTotalTransactions)
	}

	return report, nil
}

// internal/sale/service.go
package sale

import (
	"errors"
	"fmt"
	"log"
	"time"

	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/notification"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/recipe"
	"pos-fiber-app/internal/shift"

	"gorm.io/gorm"
)

type AddItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type CompleteSaleRequest struct {
	PaymentMethod    string  `json:"payment_method" validate:"required"`
	AmountPaid       float64 `json:"amount_paid" validate:"required,gte=0"`
	Discount         float64 `json:"discount" validate:"gte=0"`
	Tax              float64 `json:"tax"`
	TerminalProvider string  `json:"terminal_provider,omitempty"`
	ShiftID          *uint   `json:"shift_id,omitempty"`
}

type CreateSaleRequest struct {
	Items            []SaleItemRequest `json:"items" validate:"required,min=1"`
	PaymentMethod    string            `json:"payment_method" validate:"required"`
	AmountPaid       float64           `json:"amount_paid" validate:"required,gte=0"`
	Discount         float64           `json:"discount" validate:"gte=0"`
	Tax              float64           `json:"tax"`
	CustomerName     string            `json:"customer_name,omitempty"`
	CustomerPhone    string            `json:"customer_phone,omitempty"`
	TerminalProvider string            `json:"terminal_provider,omitempty"`
	ShiftID          *uint             `json:"shift_id,omitempty"`
}

type SaleItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

type VoidSaleRequest struct {
	Reason string `json:"reason" validate:"required,min=5"`
}

type CreateDraftRequest struct {
	Items         []SaleItemRequest `json:"items"`
	TableID       *uint             `json:"table_id"`
	TableNumber   string            `json:"table_number"`
	CustomerName  string            `json:"customer_name"`
	CustomerPhone string            `json:"customer_phone"`
	OrderType     string            `json:"order_type"`
}

type SaleFilters struct {
	Status        SaleStatus
	From          string
	To            string
	PaymentMethod string
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
	Date                  string  `json:"date"`
	TotalSales            float64 `json:"total_sales"`
	TotalTransactions     int     `json:"total_transactions"`
	CashSales             float64 `json:"cash_sales"`
	CardSales             float64 `json:"card_sales"`
	TransferSales         float64 `json:"transfer_sales"`
	ExternalTerminalSales float64 `json:"external_terminal_sales"`
	CreditSales           float64 `json:"credit_sales"`
	AverageSale           float64 `json:"average_sale"`
}

// CreateDraft starts a new sale with optional items and table info
func CreateDraft(db *gorm.DB, businessID uint, tenantID string, outletID uint, cashierID uint, req CreateDraftRequest) (*Sale, error) {
	tx := db.Begin()
	defer tx.Rollback()

	sale := &Sale{
		BusinessID:    businessID,
		TenantID:      tenantID,
		OutletID:      outletID,
		Status:        StatusDraft,
		CashierID:     cashierID,
		TableID:       req.TableID,
		TableNumber:   req.TableNumber,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		OrderType:     req.OrderType,
		SaleDate:      time.Now(),
	}

	if sale.OrderType == "" {
		sale.OrderType = "dine-in"
	}

	if err := tx.Create(sale).Error; err != nil {
		return nil, err
	}

	var subtotal float64
	for _, itemReq := range req.Items {
		var prod product.Product
		if err := tx.First(&prod, "id = ? AND business_id = ?", itemReq.ProductID, businessID).Error; err != nil {
			continue // Skip if product not found
		}

		itemTotal := float64(itemReq.Quantity) * prod.Price
		item := SaleItem{
			SaleID:      sale.ID,
			ProductID:   prod.ID,
			ProductName: prod.Name,
			Quantity:    itemReq.Quantity,
			UnitPrice:   prod.Price,
			TotalPrice:  itemTotal,
		}

		if err := tx.Create(&item).Error; err != nil {
			return nil, err
		}
		subtotal += itemTotal
	}

	sale.Subtotal = subtotal
	sale.Total = subtotal
	if err := tx.Save(sale).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return sale, nil
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

	// Recalculate total with tax and discount
	sale.Total = sale.Subtotal - req.Discount + req.Tax

	if sale.Total > req.AmountPaid {
		return nil, errors.New("insufficient payment")
	}

	// Deduct inventory
	recipeSvc := recipe.NewRecipeService(db)
	for _, item := range sale.SaleItems {
		if err := recipeSvc.AdjustStockWithRecipe(tx, item.ProductID, businessID, item.Quantity); err != nil {
			return nil, errors.New("failed to update inventory: " + err.Error())
		}
	}

	now := time.Now()
	sale.Status = StatusCompleted
	sale.PaymentMethod = req.PaymentMethod
	sale.TerminalProvider = req.TerminalProvider
	sale.Discount = req.Discount
	sale.Tax = req.Tax
	sale.SyncedAt = &now // mark as synced
	if req.ShiftID != nil {
		sale.ShiftID = req.ShiftID
	}

	// ... inside CompleteSale before saving
	// Assign daily sequence number
	seq, err := getNextDailySequence(tx, businessID)
	if err != nil {
		return nil, err
	}
	sale.DailySequence = seq

	if err := tx.Save(&sale).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Update Shift Metrics
	if sale.ShiftID != nil {
		shiftSvc := shift.NewShiftService(db)
		_ = shiftSvc.UpdateShiftMetrics(*sale.ShiftID, sale.Total, sale.PaymentMethod)
	}

	// Log Activity
	_ = LogActivity(db, sale.ID, businessID, sale.CashierID, ActionCompleted, ActivityDetails{
		AmountPaid:    req.AmountPaid,
		PaymentMethod: req.PaymentMethod,
	})

	return &SaleReceipt{
		Sale:        &sale,
		Items:       sale.SaleItems,
		Change:      req.AmountPaid - sale.Total,
		ReceiptNo:   generateReceiptNo(sale.DailySequence),
		GeneratedAt: time.Now(),
	}, nil
}

// CreateSale creates and completes a sale in one atomic operation (One-Shot)
func CreateSale(db *gorm.DB, businessID uint, tenantID string, outletID uint, cashierID uint, req CreateSaleRequest) (*SaleReceipt, error) {
	tx := db.Begin()
	defer tx.Rollback()

	now := time.Now()

	// 1. Get Next Sequence
	seq, err := getNextDailySequence(tx, businessID)
	if err != nil {
		return nil, err
	}

	// 2. Create Sale Header
	sale := &Sale{
		BusinessID:       businessID,
		TenantID:         tenantID,
		CashierID:        cashierID,
		OutletID:         outletID,
		Status:           StatusCompleted, // Direct to completed
		PaymentMethod:    req.PaymentMethod,
		Discount:         req.Discount,
		Tax:              req.Tax,
		CustomerName:     req.CustomerName,
		CustomerPhone:    req.CustomerPhone,
		TerminalProvider: req.TerminalProvider,
		SaleDate:         now,
		DailySequence:    seq,
		Subtotal:         0.0,
		Total:            0.0,
		SyncedAt:         &now,
		ShiftID:          req.ShiftID,
	}

	if err := tx.Create(sale).Error; err != nil {
		return nil, err
	}

	var subtotal float64
	var saleItems []SaleItem

	// 3. Process Items
	for _, itemReq := range req.Items {
		var prod product.Product
		if err := tx.First(&prod, "id = ? AND business_id = ?", itemReq.ProductID, businessID).Error; err != nil {
			return nil, fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Inventory check & deduction
		recipeSvc := recipe.NewRecipeService(db)
		if err := recipeSvc.AdjustStockWithRecipe(tx, prod.ID, businessID, itemReq.Quantity); err != nil {
			return nil, fmt.Errorf("insufficient stock for %s: %w", prod.Name, err)
		}

		itemPrice := prod.Price
		itemTotal := float64(itemReq.Quantity) * itemPrice

		saleItem := SaleItem{
			SaleID:      sale.ID,
			ProductID:   prod.ID,
			ProductName: prod.Name,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemPrice,
			TotalPrice:  itemTotal,
		}

		if err := tx.Create(&saleItem).Error; err != nil {
			return nil, err
		}

		subtotal += itemTotal
		saleItems = append(saleItems, saleItem)
	}

	// 4. Finalize Totals
	sale.Subtotal = subtotal
	sale.Total = subtotal - req.Discount + req.Tax

	if sale.Total > req.AmountPaid {
		return nil, errors.New("insufficient payment")
	}

	if err := tx.Save(sale).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 5. Update Shift Metrics
	if sale.ShiftID != nil {
		shiftSvc := shift.NewShiftService(db)
		_ = shiftSvc.UpdateShiftMetrics(*sale.ShiftID, sale.Total, sale.PaymentMethod)
	}

	// 6. Log Activity
	_ = LogActivity(db, sale.ID, businessID, cashierID, ActionCompleted, ActivityDetails{
		AmountPaid:    req.AmountPaid,
		PaymentMethod: req.PaymentMethod,
	})

	return &SaleReceipt{
		Sale:        sale,
		Items:       saleItems,
		Change:      req.AmountPaid - sale.Total,
		ReceiptNo:   generateReceiptNo(sale.DailySequence),
		GeneratedAt: time.Now(),
	}, nil
}

// ... existing code ...

func generateReceiptNo(sequence int) string {
	// Format: YYYYMMDD-SEQUENCE (e.g. 20260201-001)
	return time.Now().Format("20060102") + "-" + fmt.Sprintf("%03d", sequence)
}

func getNextDailySequence(tx *gorm.DB, businessID uint) (int, error) {
	var sales []Sale
	startOfDay := time.Now().Truncate(24 * time.Hour) // 00:00:00
	endOfDay := startOfDay.Add(24 * time.Hour)        // Next day 00:00:00

	// Get max daily_sequence for today including COMPLETED and VOIDED sales (to prevent reuse)
	// Using Limit(1).Find to avoid Error: Record Not Found in logger
	err := tx.Model(&Sale{}).
		Where("business_id = ? AND created_at >= ? AND created_at < ? AND status IN ?",
			businessID, startOfDay, endOfDay, []SaleStatus{StatusCompleted, StatusVoided}).
		Order("daily_sequence DESC").
		Limit(1).
		Find(&sales).Error

	if err != nil {
		return 0, err
	}

	if len(sales) == 0 {
		return 1, nil // First sale of the day
	}

	return sales[0].DailySequence + 1, nil
}

func HoldSale(db *gorm.DB, saleID, businessID uint) (*Sale, error) {
	return updateSaleStatus(db, saleID, businessID, StatusHeld)
}

func VoidSale(db *gorm.DB, saleID, businessID, userID uint, reason string) (*Sale, error) {
	tx := db.Begin()

	var sale Sale
	if err := tx.Preload("SaleItems").First(&sale, "id = ? AND business_id = ? AND status = ?", saleID, businessID, StatusCompleted).Error; err != nil {
		return nil, errors.New("sale not found or cannot be voided")
	}

	// Restock
	recipeSvc := recipe.NewRecipeService(db)
	for _, item := range sale.SaleItems {
		// Pass negative quantity to AdjustStockWithRecipe to restock (since it negates the input)
		_ = recipeSvc.AdjustStockWithRecipe(tx, item.ProductID, businessID, -item.Quantity)
	}

	sale.Status = StatusVoided
	// Add reason field if you extend model

	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Log Activity
	_ = LogActivity(db, sale.ID, businessID, userID, ActionVoided, ActivityDetails{
		Reason: reason,
	})

	// Real-time Alert for Owner
	go func() {
		notifier := notification.GetDefaultService(db)
		// Get business for currency
		var businessObj struct {
			Currency string
		}
		db.Table("businesses").Select("currency").Where("id = ?", businessID).Scan(&businessObj)

		// Get cashier name
		var cashier struct {
			FirstName string
			LastName  string
		}
		db.Table("users").Select("first_name, last_name").Where("id = ?", userID).Scan(&cashier)
		cashierName := cashier.FirstName + " " + cashier.LastName

		notifier.SendVoidAlert(businessID, sale.ID, cashierName, reason, sale.Total, businessObj.Currency)
	}()

	return &sale, nil
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

	if filters.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filters.PaymentMethod)
	}

	if filters.From != "" {
		// Try to parse date to ensure we compare correctly (timestamp vs date string)
		if t, err := time.Parse("2006-01-02", filters.From); err == nil {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			query = query.Where("sale_date >= ?", startOfDay)
		} else {
			query = query.Where("sale_date >= ?", filters.From)
		}
	} else if filters.To == "" {
		// Default to today if no date range is provided
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		query = query.Where("sale_date >= ?", startOfDay)
	}

	if filters.To != "" {
		// Include the entire end day
		if t, err := time.Parse("2006-01-02", filters.To); err == nil {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			query = query.Where("sale_date <= ?", endOfDay)
		} else {
			query = query.Where("sale_date <= ?", filters.To)
		}
	}

	// Order by latest first
	query = query.Order("sale_date DESC")

	// Select sales.* and alias the concatenated name
	query = query.Select("sales.*, users.first_name || ' ' || users.last_name as cashier_name").
		Joins("LEFT JOIN users ON users.id = sales.cashier_id").
		Preload("SaleItems") // Load sale items with product names

	if err := query.Find(&sales).Error; err != nil {
		return nil, err
	}
	return sales, nil
}

func GetSaleDetails(db *gorm.DB, saleID, businessID uint) (*SaleResult, error) {
	var sale Sale
	if err := db.Select("sales.*, users.first_name || ' ' || users.last_name as cashier_name").
		Joins("LEFT JOIN users ON users.id = sales.cashier_id").
		First(&sale, "sales.id = ? AND sales.business_id = ?", saleID, businessID).Error; err != nil {
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

	if len(results) == 0 {
		// No raw data found, check if we have a summary for this day
		var archivedSummary SaleSummary
		err := db.Where("business_id = ? AND date = ?", businessID, startOfDay).First(&archivedSummary).Error
		if err == nil {
			report.TotalSales = archivedSummary.TotalSales
			report.TotalTransactions = archivedSummary.TotalTransactions
			report.CashSales = archivedSummary.CashSales
			report.CardSales = archivedSummary.CardSales
			report.TransferSales = archivedSummary.TransferSales
			report.ExternalTerminalSales = archivedSummary.ExternalTerminalSales
			report.CreditSales = archivedSummary.CreditSales
			report.AverageSale = 0
			if report.TotalTransactions > 0 {
				report.AverageSale = report.TotalSales / float64(report.TotalTransactions)
			}
			return report, nil
		}
	}

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
		case "EXTERNAL_TERMINAL":
			report.ExternalTerminalSales = r.TotalSales
		case "CREDIT":
			report.CreditSales = r.TotalSales
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

	// 1. Process Raw Data (Status COMPLETED)
	for _, r := range results {
		grandTotalSales += r.TotalSales
		grandTotalTransactions += r.TotalTransactions

		switch r.PaymentMethod {
		case "CASH":
			report.CashSales += r.TotalSales
			report.CashTransactions += r.TotalTransactions
		case "CARD":
			report.CardSales += r.TotalSales
			report.CardTransactions += r.TotalTransactions
		case "TRANSFER":
			report.TransferSales += r.TotalSales
			report.TransferTransactions += r.TotalTransactions
		case "MOBILE_MONEY":
			report.MobileMoneySales += r.TotalSales
			report.MobileMoneyTransactions += r.TotalTransactions
		case "EXTERNAL_TERMINAL":
			report.ExternalTerminalSales += r.TotalSales
			report.ExternalTerminalTransactions += r.TotalTransactions
		case "CREDIT":
			report.CreditSales += r.TotalSales
			report.CreditTransactions += r.TotalTransactions
		default:
			report.OtherSales += r.TotalSales
			report.OtherTransactions += r.TotalTransactions
		}
	}

	// 2. Process Archived Data (SaleSummary)
	var archivedSummaries []SaleSummary
	err = db.Where("business_id = ? AND date >= ? AND date <= ?", businessID, startOfPeriod, endOfPeriod).Find(&archivedSummaries).Error
	if err == nil {
		for _, s := range archivedSummaries {
			grandTotalSales += s.TotalSales
			grandTotalTransactions += s.TotalTransactions

			report.CashSales += s.CashSales
			report.CardSales += s.CardSales
			report.TransferSales += s.TransferSales
			report.ExternalTerminalSales += s.ExternalTerminalSales
			report.CreditSales += s.CreditSales
			// Note: MobileMoney and Other are aggregated into totals if not stored individually in Summary
		}
	}

	report.TotalSales = grandTotalSales
	report.TotalTransactions = grandTotalTransactions

	if grandTotalTransactions > 0 {
		report.AverageSale = grandTotalSales / float64(grandTotalTransactions)
	}

	return report, nil
}

// PerformCleanup of old records based on retention policy
func PerformCleanup(db *gorm.DB, businessID uint, retentionMonths int, businessName string) {
	if retentionMonths <= 0 {
		return
	}

	retentionDate := time.Now().AddDate(0, -retentionMonths, 0)

	// 1. Get all dates that have raw data older than retention date
	var dates []time.Time
	db.Model(&Sale{}).
		Where("business_id = ? AND sale_date < ? AND status = ?", businessID, retentionDate, StatusCompleted).
		Select("DISTINCT date_trunc('day', sale_date)").
		Find(&dates)

	for _, d := range dates {
		// 2. Generate summary for this day
		var summary SaleSummary
		err := db.Model(&Sale{}).
			Where("business_id = ? AND sale_date >= ? AND sale_date < ? AND status = ?",
				businessID, d, d.AddDate(0, 0, 1), StatusCompleted).
			Select(`
				COUNT(*) as total_transactions,
				SUM(total) as total_sales,
				SUM(tax) as tax,
				SUM(discount) as discount,
				SUM(CASE WHEN payment_method = 'CASH' THEN total ELSE 0 END) as cash_sales,
				SUM(CASE WHEN payment_method = 'CARD' THEN total ELSE 0 END) as card_sales,
				SUM(CASE WHEN payment_method = 'TRANSFER' THEN total ELSE 0 END) as transfer_sales,
				SUM(CASE WHEN payment_method = 'EXTERNAL_TERMINAL' THEN total ELSE 0 END) as external_terminal_sales,
				SUM(CASE WHEN payment_method = 'CREDIT' THEN total ELSE 0 END) as credit_sales
			`).Scan(&summary).Error

		if err == nil && summary.TotalTransactions > 0 {
			summary.BusinessID = businessID
			summary.Date = d

			// 3. Upsert into SaleSummary
			db.Where("business_id = ? AND date = ?", businessID, d).
				FirstOrCreate(&SaleSummary{}).
				Updates(summary)
		}
	}

	// 4. Finally perform the removal of raw data
	var count int64
	db.Model(&Sale{}).Where("business_id = ? AND sale_date < ? AND status IN ?",
		businessID, retentionDate, []SaleStatus{StatusCompleted, StatusVoided}).Count(&count)

	if count > 0 {
		log.Printf("[CLEANUP] Removing %d old transactions for %s (Older than %s)\n", count, businessName, retentionDate.Format("2006-01-02"))

		// 1. Delete Sale Items first
		db.Exec("DELETE FROM sale_items WHERE sale_id IN (SELECT id FROM sales WHERE business_id = ? AND sale_date < ?)", businessID, retentionDate)

		// 2. Delete Sales
		result := db.Where("business_id = ? AND sale_date < ? AND status IN ?",
			businessID, retentionDate, []SaleStatus{StatusCompleted, StatusVoided}).Delete(&Sale{})

		if result.Error != nil {
			log.Printf("[CLEANUP ERROR] Failed for %s: %v\n", businessName, result.Error)
		} else {
			log.Printf("[CLEANUP SUCCESS] Purged old data for %s after summarizing\n", businessName)
		}
	}
}

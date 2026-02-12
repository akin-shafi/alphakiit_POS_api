// internal/sale/service_enhanced.go
package sale

import (
	"errors"
	"time"

	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/product"

	"gorm.io/gorm"
)

// CreateDraftWithReservation creates a new draft sale with table assignment
func CreateDraftWithReservation(db *gorm.DB, businessID uint, tenantID string, cashierID uint, shiftID *uint, req CreateDraftRequest) (*Sale, error) {
	sale := &Sale{
		BusinessID:    businessID,
		TenantID:      tenantID,
		Status:        StatusDraft,
		CashierID:     cashierID,
		ShiftID:       shiftID,
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

	if err := db.Create(sale).Error; err != nil {
		return nil, err
	}

	// Log activity
	details := ActivityDetails{}
	if req.TableNumber != "" {
		details.NewValue = req.TableNumber
	}
	LogActivity(db, sale.ID, businessID, cashierID, ActionCreated, details)

	return sale, nil
}

// AddItemToSaleWithReservation adds item to sale and creates stock reservation
func AddItemToSaleWithReservation(db *gorm.DB, saleID, businessID, cashierID uint, productID uint, qty int) (*SaleResult, error) {
	tx := db.Begin()
	defer tx.Rollback()

	// Get sale
	var sale Sale
	if err := tx.First(&sale, "id = ? AND business_id = ? AND status IN ?", saleID, businessID, []SaleStatus{StatusDraft, StatusHeld}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sale not found or not editable")
		}
		return nil, err
	}

	// Get product
	var prod product.Product
	if err := tx.Table("products").First(&prod, "id = ? AND business_id = ?", productID, businessID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Initialize reservation service
	reservationService := inventory.NewReservationService(tx)

	// Check if we already have a reservation for this product in this sale
	var existingItem SaleItem
	existingErr := tx.First(&existingItem, "sale_id = ? AND product_id = ?", saleID, productID).Error

	var currentReservedQty int
	if existingErr == nil {
		// Item already exists, we have a reservation
		currentReservedQty = existingItem.Quantity
	}

	// Calculate new total quantity
	newTotalQty := currentReservedQty + qty

	// Check available stock (accounting for current reservation)
	availableStock, err := reservationService.GetAvailableStock(productID, businessID)
	if err != nil {
		return nil, err
	}

	// Add back current reservation to available for this check
	effectiveAvailable := availableStock + currentReservedQty

	if effectiveAvailable < newTotalQty {
		return nil, errors.New("insufficient stock available for reservation")
	}

	// Upsert sale item
	var item SaleItem
	if existingErr == nil {
		// Update existing item
		item = existingItem
		item.Quantity = newTotalQty
		item.TotalPrice = float64(item.Quantity) * prod.Price
	} else {
		// Create new item
		item = SaleItem{
			SaleID:      saleID,
			ProductID:   productID,
			ProductName: prod.Name,
			Quantity:    qty,
			UnitPrice:   prod.Price,
			TotalPrice:  float64(qty) * prod.Price,
		}
	}

	if err := tx.Save(&item).Error; err != nil {
		return nil, err
	}

	// Update or create stock reservation
	if currentReservedQty > 0 {
		// Update existing reservation
		if err := reservationService.UpdateReservationQuantity(saleID, productID, newTotalQty); err != nil {
			return nil, err
		}
	} else {
		// Create new reservation
		if err := reservationService.ReserveStock(saleID, productID, businessID, cashierID, qty); err != nil {
			return nil, err
		}
	}

	// Recalculate sale totals
	if err := recalculateSaleTotals(tx, &sale); err != nil {
		return nil, err
	}

	// Log activity
	details := ActivityDetails{
		ProductID:   productID,
		ProductName: prod.Name,
		Quantity:    qty,
	}
	LogActivity(tx, saleID, businessID, cashierID, ActionItemAdded, details)

	// Get all items
	var items []SaleItem
	tx.Where("sale_id = ?", saleID).Find(&items)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SaleResult{Sale: &sale, Items: items}, nil
}

// CompleteSaleWithReservation completes a sale, deducts inventory, and releases reservations
func CompleteSaleWithReservation(db *gorm.DB, saleID, businessID, cashierID uint, req CompleteSaleRequest) (*SaleReceipt, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var sale Sale
	if err := tx.Preload("SaleItems").First(&sale, "id = ? AND business_id = ? AND status = ?", saleID, businessID, StatusDraft).Error; err != nil {
		return nil, errors.New("sale not found or already completed")
	}

	if sale.Total-req.Discount > req.AmountPaid {
		return nil, errors.New("insufficient payment")
	}

	// Initialize reservation service
	reservationService := inventory.NewReservationService(tx)

	// Deduct inventory and release reservations
	for _, item := range sale.SaleItems {
		// Deduct actual inventory
		if err := inventory.AdjustStock(tx, item.ProductID, businessID, -item.Quantity); err != nil {
			return nil, errors.New("failed to update inventory: " + err.Error())
		}

		// Release reservation
		if err := reservationService.ReleaseReservation(saleID, item.ProductID); err != nil {
			// Log but don't fail - reservation might have expired
			// TODO: Add proper logging
		}
	}

	now := time.Now()
	sale.Status = StatusCompleted
	sale.PaymentMethod = req.PaymentMethod
	sale.Discount = req.Discount
	sale.SyncedAt = &now

	// Assign daily sequence number
	seq, err := getNextDailySequence(tx, businessID)
	if err != nil {
		return nil, err
	}
	sale.DailySequence = seq

	if err := tx.Save(&sale).Error; err != nil {
		return nil, err
	}

	// Update shift metrics if sale is linked to a shift
	if sale.ShiftID != nil {
		tx.Exec("UPDATE shifts SET total_sales = total_sales + ?, transaction_count = transaction_count + 1 WHERE id = ?",
			sale.Total-sale.Discount, *sale.ShiftID)
	}

	// Log activity
	details := ActivityDetails{
		PaymentMethod: req.PaymentMethod,
		AmountPaid:    req.AmountPaid,
	}
	LogActivity(tx, saleID, businessID, cashierID, ActionCompleted, details)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SaleReceipt{
		Sale:        &sale,
		Items:       sale.SaleItems,
		Change:      req.AmountPaid - (sale.Total - req.Discount),
		ReceiptNo:   generateReceiptNo(sale.DailySequence),
		GeneratedAt: time.Now(),
	}, nil
}

// VoidSaleWithReservation voids a completed sale and releases any reservations
func VoidSaleWithReservation(db *gorm.DB, saleID, businessID, cashierID uint, reason string) (*Sale, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var sale Sale
	if err := tx.Preload("SaleItems").First(&sale, "id = ? AND business_id = ?", saleID, businessID).Error; err != nil {
		return nil, errors.New("sale not found")
	}

	// Initialize reservation service
	reservationService := inventory.NewReservationService(tx)

	if sale.Status == StatusCompleted {
		// Restock inventory for completed sales
		for _, item := range sale.SaleItems {
			inventory.AdjustStock(tx, item.ProductID, businessID, item.Quantity)
		}

		// Update shift metrics if applicable
		if sale.ShiftID != nil {
			tx.Exec("UPDATE shifts SET total_sales = total_sales - ?, transaction_count = transaction_count - 1 WHERE id = ?",
				sale.Total-sale.Discount, *sale.ShiftID)
		}
	} else if sale.Status == StatusDraft || sale.Status == StatusHeld {
		// Release reservations for draft/held sales
		reservationService.ReleaseAllReservations(saleID)
	}

	sale.Status = StatusVoided

	if err := tx.Save(&sale).Error; err != nil {
		return nil, err
	}

	// Log activity
	details := ActivityDetails{
		Reason: reason,
	}
	LogActivity(tx, saleID, businessID, cashierID, ActionVoided, details)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &sale, nil
}

// DeleteDraft deletes a draft sale and releases all reservations
func DeleteDraft(db *gorm.DB, saleID, businessID uint) error {
	tx := db.Begin()
	defer tx.Rollback()

	var sale Sale
	if err := tx.First(&sale, "id = ? AND business_id = ?", saleID, businessID).Error; err != nil {
		return errors.New("sale not found")
	}

	if sale.Status != StatusDraft && sale.Status != StatusHeld {
		return errors.New("can only delete draft or held sales")
	}

	// Release all stock reservations
	reservationService := inventory.NewReservationService(tx)
	if err := reservationService.ReleaseAllReservations(saleID); err != nil {
		// Log but continue - reservations might have expired
	}

	// Delete sale items (cascade should handle this, but being explicit)
	tx.Where("sale_id = ?", saleID).Delete(&SaleItem{})

	// Delete the sale
	if err := tx.Delete(&sale).Error; err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// ResumeDraft extends the reservation expiry for a draft order
func ResumeDraft(db *gorm.DB, saleID, businessID, cashierID uint) (*SaleResult, error) {
	var sale Sale
	if err := db.Preload("SaleItems").First(&sale, "id = ? AND business_id = ? AND status IN ?",
		saleID, businessID, []SaleStatus{StatusDraft, StatusHeld}).Error; err != nil {
		return nil, errors.New("draft sale not found")
	}

	// Extend reservation expiry by updating each reservation
	reservationService := inventory.NewReservationService(db)
	reservations, err := reservationService.GetReservationsBySale(saleID)
	if err != nil {
		return nil, err
	}

	// Update expiry time for each reservation (reset to 4 hours from now)
	for _, res := range reservations {
		db.Model(&inventory.StockReservation{}).
			Where("id = ?", res.ID).
			Update("expire_at", time.Now().Add(4*time.Hour))
	}

	// Log activity
	details := ActivityDetails{}
	LogActivity(db, saleID, businessID, cashierID, ActionResumed, details)

	return &SaleResult{Sale: &sale, Items: sale.SaleItems}, nil
}

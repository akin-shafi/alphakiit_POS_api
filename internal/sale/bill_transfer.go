// internal/sale/bill_transfer.go
package sale

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// TransferBillRequest contains the data needed to transfer a bill
type TransferBillRequest struct {
	ToTableID     *uint  `json:"to_table_id"`
	ToTableNumber string `json:"to_table_number" validate:"required"`
}

// MergeBillsRequest contains the data needed to merge multiple bills
type MergeBillsRequest struct {
	SecondarySaleIDs  []uint `json:"secondary_sale_ids" validate:"required,min=1"`
	TargetTableID     *uint  `json:"target_table_id"`
	TargetTableNumber string `json:"target_table_number"`
}

// TransferBill moves a sale from one table to another
func TransferBill(db *gorm.DB, saleID, businessID, userID uint, req TransferBillRequest) (*Sale, error) {
	tx := db.Begin()
	defer tx.Rollback()

	// Get the sale
	var sale Sale
	if err := tx.First(&sale, "id = ? AND business_id = ?", saleID, businessID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("sale not found")
		}
		return nil, err
	}

	// Validate sale status
	if sale.Status != StatusDraft && sale.Status != StatusHeld {
		return nil, errors.New("can only transfer draft or held sales")
	}

	// Store old table for logging
	oldTableNumber := sale.TableNumber

	// Update sale with new table
	sale.TableID = req.ToTableID
	sale.TableNumber = req.ToTableNumber

	if err := tx.Save(&sale).Error; err != nil {
		return nil, err
	}

	// Log the transfer activity
	details := ActivityDetails{
		FromTable: oldTableNumber,
		ToTable:   req.ToTableNumber,
	}
	if err := LogActivity(tx, saleID, businessID, userID, ActionTransferred, details); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &sale, nil
}

// MergeBills combines multiple sales into one primary sale
func MergeBills(db *gorm.DB, primarySaleID, businessID, userID uint, req MergeBillsRequest) (*Sale, error) {
	tx := db.Begin()
	defer tx.Rollback()

	// 1. Get primary sale
	var primarySale Sale
	if err := tx.Preload("SaleItems").First(&primarySale, "id = ? AND business_id = ?", primarySaleID, businessID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("primary sale not found")
		}
		return nil, err
	}

	// Validate primary sale status
	if primarySale.Status != StatusDraft && primarySale.Status != StatusHeld {
		return nil, errors.New("can only merge draft or held sales")
	}

	// 2. Process each secondary sale
	for _, secondarySaleID := range req.SecondarySaleIDs {
		if secondarySaleID == primarySaleID {
			return nil, errors.New("cannot merge a sale with itself")
		}

		var secondarySale Sale
		if err := tx.Preload("SaleItems").First(&secondarySale, "id = ? AND business_id = ?", secondarySaleID, businessID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("secondary sale %d not found", secondarySaleID)
			}
			return nil, err
		}

		// Validate secondary sale status
		if secondarySale.Status != StatusDraft && secondarySale.Status != StatusHeld {
			return nil, fmt.Errorf("secondary sale %d must be draft or held", secondarySaleID)
		}

		// 3. Move sale items from secondary to primary
		result := tx.Model(&SaleItem{}).
			Where("sale_id = ?", secondarySaleID).
			Update("sale_id", primarySaleID)

		if result.Error != nil {
			return nil, fmt.Errorf("failed to move items from sale %d: %w", secondarySaleID, result.Error)
		}

		// 4. Transfer stock reservations (if reservation system is being used)
		// Update reservations to point to primary sale
		tx.Exec("UPDATE stock_reservations SET sale_id = ? WHERE sale_id = ?", primarySaleID, secondarySaleID)

		// 5. Delete the secondary sale
		if err := tx.Delete(&Sale{}, secondarySaleID).Error; err != nil {
			return nil, fmt.Errorf("failed to delete secondary sale %d: %w", secondarySaleID, err)
		}
	}

	// 6. Recalculate primary sale totals
	if err := recalculateSaleTotals(tx, &primarySale); err != nil {
		return nil, err
	}

	// 7. Update table assignment if provided
	if req.TargetTableID != nil || req.TargetTableNumber != "" {
		primarySale.TableID = req.TargetTableID
		if req.TargetTableNumber != "" {
			primarySale.TableNumber = req.TargetTableNumber
		}
		if err := tx.Save(&primarySale).Error; err != nil {
			return nil, err
		}
	}

	// 8. Log the merge activity
	details := ActivityDetails{
		MergedFrom: req.SecondarySaleIDs,
	}
	if err := LogActivity(tx, primarySaleID, businessID, userID, ActionMerged, details); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload the sale with updated items
	if err := db.Preload("SaleItems").First(&primarySale, primarySaleID).Error; err != nil {
		return nil, err
	}

	return &primarySale, nil
}

// SplitBill splits a sale into multiple sales (future enhancement)
// This is a placeholder for future implementation
func SplitBill(db *gorm.DB, saleID, businessID, userID uint, splitConfig interface{}) error {
	// TODO: Implement bill splitting logic
	// This would involve:
	// 1. Creating new sales for each split
	// 2. Distributing items across the new sales
	// 3. Handling stock reservations
	// 4. Logging the split activity
	return errors.New("bill splitting not yet implemented")
}

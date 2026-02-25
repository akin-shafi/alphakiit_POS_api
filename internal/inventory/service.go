// internal/inventory/service.go
package inventory

import (
	"errors"
	"fmt"
	"pos-fiber-app/internal/notification"
	"time"

	"gorm.io/gorm"
)

func AdjustStock(tx *gorm.DB, productID, businessID uint, quantity int) error {
	// 1. Get product tracking info and current stock in one go
	var prodInfo struct {
		TrackByRound bool
		Stock        int
	}
	if err := tx.Table("products").Select("track_by_round, stock").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prodInfo).Error; err != nil {
		return fmt.Errorf("failed to fetch product info: %w", err)
	}

	// 2. If tracked by round, delegate and return
	if prodInfo.TrackByRound {
		return AdjustStockFromRound(tx, productID, businessID, float64(quantity))
	}

	// 3. Normal inventory adjustment
	var inv Inventory
	err := tx.Where("product_id = ? AND business_id = ?", productID, businessID).First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create missing inventory record from products.stock
			inv = Inventory{
				ProductID:    productID,
				BusinessID:   businessID,
				CurrentStock: prodInfo.Stock,
			}
			if err := tx.Create(&inv).Error; err != nil {
				return fmt.Errorf("failed to initialize inventory record: %w", err)
			}
		} else {
			return fmt.Errorf("inventory lookup failed: %w", err)
		}
	}

	// 4. Stock Sync (Best effort for out-of-sync records)
	if inv.CurrentStock == 0 && inv.CreatedAt.Equal(inv.UpdatedAt) && prodInfo.Stock > 0 {
		inv.CurrentStock = prodInfo.Stock
	}

	// 5. Final validation and updates
	newStock := inv.CurrentStock + quantity
	if newStock < 0 {
		return fmt.Errorf("available: %d, needed: %d", inv.CurrentStock, -quantity)
	}

	inv.CurrentStock = newStock
	inv.LastRestocked = time.Now()

	if err := tx.Save(&inv).Error; err != nil {
		return fmt.Errorf("failed to save inventory update: %w", err)
	}

	if err := tx.Table("products").Where("id = ? AND business_id = ?", productID, businessID).Update("stock", newStock).Error; err != nil {
		return fmt.Errorf("failed to sync product stock: %w", err)
	}

	// 6. Check for Low Stock Alert
	if newStock <= inv.LowStockAlert && quantity < 0 { // Alert only on deduction
		go func() {
			notifier := notification.GetDefaultService(tx)
			var prodName string
			tx.Table("products").Select("name").Where("id = ?", productID).Scan(&prodName)
			notifier.SendLowStockAlert(businessID, prodName, newStock, inv.LowStockAlert)
		}()
	}

	return nil
}

func AdjustStockFromRound(db *gorm.DB, productID, businessID uint, quantity float64) error {
	var round InventoryRound
	// Find the current OPEN round for this product
	err := db.Where("product_id = ? AND business_id = ? AND status = 'OPEN'", productID, businessID).
		Order("start_date DESC").First(&round).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && quantity < 0 {
			return errors.New("no open stock round found for this product")
		}
		// If it's a restock (voiding) or other error, return it unless we decide to ignore it for restocks
		return err
	}

	newRemaining := round.RemainingVolume + quantity
	if newRemaining < 0 && quantity < 0 {
		return fmt.Errorf("insufficient round stock: available %.3f, requested %.3f", round.RemainingVolume, -quantity)
	}

	round.RemainingVolume = newRemaining

	if err := db.Save(&round).Error; err != nil {
		return err
	}

	// Check for Low Stock in Bulk Round (15% threshold as default)
	if newRemaining <= (0.15*round.TotalVolume) && quantity < 0 {
		go func() {
			notifier := notification.GetDefaultService(db)
			var prodName string
			db.Table("products").Select("name").Where("id = ?", productID).Scan(&prodName)
			notifier.SendLowStockAlert(businessID, prodName+" (Bulk)", int(newRemaining), int(0.15*round.TotalVolume))
		}()
	}

	return nil
}

func StartNewRound(db *gorm.DB, businessID, productID uint, totalVolume float64) (*InventoryRound, error) {
	// First, check if there's already an OPEN round
	var existing int64
	db.Model(&InventoryRound{}).Where("product_id = ? AND business_id = ? AND status = 'OPEN'", productID, businessID).Count(&existing)
	if existing > 0 {
		return nil, errors.New("there is already an open round for this product")
	}

	round := &InventoryRound{
		BusinessID:      businessID,
		ProductID:       productID,
		TotalVolume:     totalVolume,
		RemainingVolume: totalVolume,
		Status:          "OPEN",
		StartDate:       time.Now(),
	}

	if err := db.Create(round).Error; err != nil {
		return nil, err
	}

	return round, nil
}

func CloseRound(db *gorm.DB, businessID, roundID uint) error {
	var round InventoryRound
	if err := db.First(&round, "id = ? AND business_id = ?", roundID, businessID).Error; err != nil {
		return err
	}

	now := time.Now()
	round.Status = "CLOSED"
	round.ClosedAt = &now

	return db.Save(&round).Error
}

func GetActiveRound(db *gorm.DB, businessID, productID uint) (*InventoryRound, error) {
	var round InventoryRound
	err := db.Where("product_id = ? AND business_id = ? AND status = 'OPEN'", productID, businessID).First(&round).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

func GetAllActiveRounds(db *gorm.DB, businessID uint) ([]InventoryRound, error) {
	var rounds []InventoryRound
	err := db.Where("business_id = ? AND status = 'OPEN'", businessID).Order("start_date DESC").Find(&rounds).Error
	if err != nil {
		return nil, err
	}
	return rounds, nil
}

func GetStock(db *gorm.DB, productID, businessID uint) (*Inventory, error) {
	var inv Inventory
	err := db.Where("product_id = ? AND business_id = ?", productID, businessID).First(&inv).Error
	return &inv, err
}

func ListLowStockItems(db *gorm.DB, businessID uint, threshold int) ([]Inventory, error) {
	var items []Inventory
	err := db.Where("business_id = ? AND current_stock <= low_stock_alert", businessID).Find(&items).Error
	return items, err
}

func ListInventoryByBusiness(db *gorm.DB, businessID uint) ([]Inventory, error) {
	var items []Inventory
	err := db.Where("business_id = ?", businessID).Find(&items).Error
	return items, err
}

// GetEffectiveStock returns current stock, syncing from Product table if necessary
func GetEffectiveStock(db *gorm.DB, productID, businessID uint) (int, error) {
	var inv Inventory
	err := db.Where("product_id = ? AND business_id = ?", productID, businessID).First(&inv).Error
	if err == nil {
		// If inventory record exists but has 0 stock and hasn't been updated yet,
		// it might have been created by a previous failed attempt/bug.
		if inv.CurrentStock == 0 && inv.CreatedAt.Equal(inv.UpdatedAt) {
			var prod struct {
				Stock int
			}
			if err := db.Table("products").Select("stock").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prod).Error; err == nil {
				if prod.Stock > 0 {
					return prod.Stock, nil
				}
			}
		}
		return inv.CurrentStock, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var prod struct {
			Stock int
		}
		if err := db.Table("products").Select("stock").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prod).Error; err == nil {
			return prod.Stock, nil
		}
		return 0, errors.New("product not found")
	}

	return 0, err
}

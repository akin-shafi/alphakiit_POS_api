// internal/inventory/service.go
package inventory

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func AdjustStock(db *gorm.DB, productID, businessID uint, quantity int) error {
	var inv Inventory
	// Check if inventory record exists
	err := db.Where("product_id = ? AND business_id = ?", productID, businessID).First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If inventory record doesn't exist, try to get initial stock from Product table
			var prod struct {
				Stock int
			}
			// Use a raw query or direct model to avoid circular imports if product package is used
			if err := db.Table("products").Select("stock").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prod).Error; err == nil {
				inv = Inventory{
					ProductID:    productID,
					BusinessID:   businessID,
					CurrentStock: prod.Stock,
				}
				// Create the inventory record with the product's initial stock
				if err := db.Create(&inv).Error; err != nil {
					return err
				}
			} else {
				// Fallback to creating with 0 if product not found (unlikely)
				inv = Inventory{ProductID: productID, BusinessID: businessID, CurrentStock: 0}
				if err := db.Create(&inv).Error; err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	// If inventory record exists but has 0 stock and hasn't been updated yet,
	// it might have been created by a previous failed attempt/bug.
	// We sync from Product table in this case too.
	if inv.CurrentStock == 0 && inv.CreatedAt.Equal(inv.UpdatedAt) {
		var prod struct {
			Stock int
		}
		if err := db.Table("products").Select("stock").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prod).Error; err == nil {
			if prod.Stock > 0 {
				inv.CurrentStock = prod.Stock
				db.Save(&inv)
			}
		}
	}

	// Check if deduction would result in negative stock
	newStock := inv.CurrentStock + quantity
	if newStock < 0 {
		return fmt.Errorf("insufficient stock: available %d, requested %d", inv.CurrentStock, -quantity)
	}

	// Check if product is tracked by rounds (Bulk Stock Rounds)
	var prod struct {
		TrackByRound bool
	}
	if err := db.Table("products").Select("track_by_round").Where("id = ? AND business_id = ?", productID, businessID).Scan(&prod).Error; err == nil && prod.TrackByRound {
		// quantity is negative for deduction, positive for restock/void
		return AdjustStockFromRound(db, productID, businessID, float64(quantity))
	}

	inv.CurrentStock = newStock
	inv.LastRestocked = time.Now()
	return db.Save(&inv).Error
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
		// If it's a restock (voiding) maybe we don't strictly need a round, but for now we follow the rule
		if err != nil {
			return err
		}
	}

	newRemaining := round.RemainingVolume + quantity
	if newRemaining < 0 && quantity < 0 {
		return fmt.Errorf("insufficient round stock: available %.3f, requested %.3f", round.RemainingVolume, -quantity)
	}

	round.RemainingVolume = newRemaining
	// Auto-close if it reaches exactly 0? Maybe better to let the user close it manually as requested.
	return db.Save(&round).Error
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

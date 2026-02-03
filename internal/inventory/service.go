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

	inv.CurrentStock = newStock
	inv.LastRestocked = time.Now()
	return db.Save(&inv).Error
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

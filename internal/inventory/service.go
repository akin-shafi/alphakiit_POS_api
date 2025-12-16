// internal/inventory/service.go
package inventory

import (
	"time"

	"gorm.io/gorm"
)

func AdjustStock(db *gorm.DB, productID, businessID uint, quantity int) error {
	var inv Inventory
	err := db.FirstOrCreate(&inv, Inventory{ProductID: productID, BusinessID: businessID}).Error
	if err != nil {
		return err
	}
	inv.CurrentStock += quantity
	if inv.CurrentStock < 0 {
		inv.CurrentStock = 0
	}
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
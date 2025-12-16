// internal/inventory/model.go
package inventory

import "time"

type Inventory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductID     uint      `gorm:"uniqueIndex:idx_product_business" json:"product_id"`
	BusinessID    uint      `gorm:"uniqueIndex:idx_product_business" json:"business_id"`
	CurrentStock  int       `json:"current_stock"`
	LowStockAlert int       `json:"low_stock_alert" default:"10"`
	LastRestocked time.Time `json:"last_restocked,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

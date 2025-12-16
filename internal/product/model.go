// internal/product/model.go
package product

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	BusinessID  uint           `gorm:"index" json:"business_id"`
	CategoryID  uint           `gorm:"index" json:"category_id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	SKU         string         `gorm:"size:50;uniqueIndex" json:"sku,omitempty"`
	Description string         `json:"description,omitempty"`
	Price       float64        `gorm:"type:decimal(10,2)" json:"price"`
	Cost        float64        `gorm:"type:decimal(10,2)" json:"cost,omitempty"`
	ImageURL    string         `json:"image_url,omitempty"`
	Active      bool           `json:"active" default:"true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
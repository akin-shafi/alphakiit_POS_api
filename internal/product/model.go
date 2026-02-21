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
	Stock       int            `json:"stock"`
	MinStock    int            `json:"min_stock"`
	Barcode     string         `gorm:"size:100" json:"barcode,omitempty"`
	TrackByRound bool          `gorm:"default:false" json:"track_by_round"`
	UnitOfMeasure string        `gorm:"size:20" json:"unit_of_measure,omitempty"` // e.g., Liters, Tons
	Active      bool           `json:"active" default:"true"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
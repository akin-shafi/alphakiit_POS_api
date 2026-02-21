package inventory

import (
	"time"

	"gorm.io/gorm"
)

type InventoryRound struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	BusinessID      uint           `gorm:"index" json:"business_id"`
	ProductID       uint           `gorm:"index" json:"product_id"`
	TotalVolume     float64        `gorm:"type:decimal(12,3)" json:"total_volume"`
	RemainingVolume float64        `gorm:"type:decimal(12,3)" json:"remaining_volume"`
	Status          string         `gorm:"type:varchar(20);default:'OPEN'" json:"status"` // OPEN, CLOSED
	StartDate       time.Time      `json:"start_date"`
	ClosedAt        *time.Time     `json:"closed_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InventoryRound) TableName() string {
	return "inventory_rounds"
}

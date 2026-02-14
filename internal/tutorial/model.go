package tutorial

import (
	"pos-fiber-app/internal/common"
	"time"

	"gorm.io/gorm"
)

type Tutorial struct {
	ID           uint                `gorm:"primaryKey" json:"id"`
	BusinessType common.BusinessType `gorm:"size:50;index" json:"business_type"` // LPG_STATION, RESTAURANT, etc. or "ALL"
	Topic        string              `gorm:"size:100" json:"topic"`              // Sales, Stock, Flow
	Title        string              `gorm:"size:200" json:"title"`
	Content      string              `gorm:"type:text" json:"content"` // Markdown or plain text
	DisplayOrder int                 `json:"display_order"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Tutorial{})
}

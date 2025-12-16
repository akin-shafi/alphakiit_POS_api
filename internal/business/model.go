// internal/business/model.go
package business

import (
	"time"

	"pos-fiber-app/internal/common" // Import shared types

	"gorm.io/gorm"
)

// Business model
type Business struct {
	ID        uint                `gorm:"primaryKey" json:"id"`
	TenantID  string              `gorm:"index;size:8" json:"tenant_id"`
	Name      string              `gorm:"size:255;not null" json:"name"`
	Type      common.BusinessType `gorm:"type:varchar(50);not null" json:"type"`
	Address   string              `json:"address,omitempty"`
	City      string              `json:"city,omitempty"`
	Currency  common.Currency     `gorm:"type:varchar(3);default:'NGN';not null" json:"currency"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `gorm:"index" json:"-"`
}

// Tenant model (keep if still used elsewhere)
type Tenant struct {
	ID        string `gorm:"primaryKey;size:8"`
	Name      string `gorm:"size:255"`
	OwnerID   uint   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

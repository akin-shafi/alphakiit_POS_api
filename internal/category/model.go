// internal/category/model.go
package category

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	BusinessID  uint           `gorm:"index" json:"business_id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

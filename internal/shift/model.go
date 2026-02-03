package shift

import (
	"time"

	"gorm.io/gorm"
)

type Shift struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	BusinessID uint       `gorm:"index" json:"business_id"`
	UserID     uint       `gorm:"index" json:"user_id"`
	UserName   string     `json:"user_name"` // Snapshot for history
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	StartCash  float64    `json:"start_cash"`
	EndCash    *float64   `json:"end_cash"`
	Status     string     `gorm:"type:varchar(20);default:'open'" json:"status"` // open, closed
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Shift{})
}

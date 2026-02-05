package shift

import (
	"time"

	"gorm.io/gorm"
)

type Shift struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	BusinessID       uint       `gorm:"index;index:idx_business_created" json:"business_id"`
	UserID           uint       `gorm:"index" json:"user_id"`
	UserName         string     `json:"user_name"` // Snapshot for history
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	StartCash        float64    `json:"start_cash"`
	EndCash          *float64   `json:"end_cash"`
	Status           string     `gorm:"type:varchar(20);default:'open'" json:"status"` // open, closed
	TerminalID       *uint      `json:"terminal_id,omitempty"`                         // Which device/terminal
	TotalSales       float64    `gorm:"type:decimal(12,2);default:0" json:"total_sales"`
	TransactionCount int        `gorm:"default:0" json:"transaction_count"`
	Notes            string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time  `gorm:"index:idx_business_created" json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Shift{})
}

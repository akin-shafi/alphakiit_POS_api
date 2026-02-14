package shift

import (
	"time"

	"gorm.io/gorm"
)

type Shift struct {
	ID                         uint           `gorm:"primaryKey" json:"id"`
	BusinessID                 uint           `gorm:"index;index:idx_business_created" json:"business_id"`
	UserID                     uint           `gorm:"index" json:"user_id"`
	UserName                   string         `json:"user_name"`      // Person who started shift
	ClosedByName               string         `json:"closed_by_name"` // Person who ended shift
	StartTime                  time.Time      `json:"start_time"`
	EndTime                    *time.Time     `json:"end_time"`
	StartCash                  float64        `json:"start_cash"`
	EndCash                    *float64       `json:"end_cash"`
	Status                     string         `gorm:"type:varchar(20);default:'open'" json:"status"` // open, closed
	TerminalID                 *uint          `json:"terminal_id,omitempty"`                         // Which device/terminal
	TotalSales                 float64        `gorm:"type:decimal(12,2);default:0" json:"total_sales"`
	TotalCashSales             float64        `gorm:"type:decimal(12,2);default:0" json:"total_cash_sales"`
	TotalCardSales             float64        `gorm:"type:decimal(12,2);default:0" json:"total_card_sales"`
	TotalTransferSales         float64        `gorm:"type:decimal(12,2);default:0" json:"total_transfer_sales"`
	TotalExternalTerminalSales float64        `gorm:"type:decimal(12,2);default:0" json:"total_external_terminal_sales"`
	TotalCreditSales           float64        `gorm:"type:decimal(12,2);default:0" json:"total_credit_sales"`
	TransactionCount           int            `gorm:"default:0" json:"transaction_count"`
	ExpectedCash               float64        `gorm:"type:decimal(12,2);default:0" json:"expected_cash"`
	CashVariance               float64        `gorm:"type:decimal(12,2);default:0" json:"cash_variance"`
	Notes                      string         `gorm:"type:text" json:"notes,omitempty"`
	Readings                   []ShiftReading `json:"readings,omitempty" gorm:"foreignKey:ShiftID"`
	CreatedAt                  time.Time      `gorm:"index:idx_business_created" json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
}

// ShiftReading tracks non-monetary metrics like Fuel/Gas pump readings
type ShiftReading struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ShiftID      uint    `gorm:"index" json:"shift_id"`
	ProductID    uint    `json:"product_id"`    // Typically the "Pump" or "Tank" product
	OpeningValue float64 `json:"opening_value"` // e.g. Meter reading at start
	ClosingValue float64 `json:"closing_value"` // e.g. Meter reading at end
	Difference   float64 `json:"difference"`    // Calculated consumption
	CreatedAt    time.Time
}

// ActiveReading represents simplified reading inputs used in service/controller
type ActiveReading struct {
	ProductID    uint
	ClosingValue float64
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Shift{}, &ShiftReading{})
}

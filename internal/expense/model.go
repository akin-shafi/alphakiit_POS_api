package expense

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	BusinessID  uint           `gorm:"index" json:"business_id"`
	Amount      float64        `gorm:"type:decimal(20,2)" json:"amount"`
	Category    string         `json:"category"` // e.g., Rent, Utilities, Salary, Other
	Description string         `json:"description"`
	Date        time.Time      `gorm:"index" json:"date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateExpenseRequest struct {
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Category    string    `json:"category" validate:"required"`
	Description string    `json:"description"`
	Date        time.Time `json:"date" validate:"required"`
}

type UpdateExpenseRequest struct {
	Amount      *float64   `json:"amount"`
	Category    *string    `json:"category"`
	Description *string    `json:"description"`
	Date        *time.Time `json:"date"`
}

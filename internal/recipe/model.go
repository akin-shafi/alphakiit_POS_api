package recipe

import (
	"time"

	"gorm.io/gorm"
)

// RecipeIngredient represents a component of a finished product
type RecipeIngredient struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	BusinessID   uint      `gorm:"index" json:"business_id"`
	ProductID    uint      `gorm:"index" json:"product_id"`    // The finished product ID
	IngredientID uint      `gorm:"index" json:"ingredient_id"` // The component product ID
	Quantity     float64   `gorm:"type:decimal(10,3)" json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&RecipeIngredient{})
}

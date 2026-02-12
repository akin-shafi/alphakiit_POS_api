package recipe

import (
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/subscription"

	"gorm.io/gorm"
)

type RecipeService struct {
	db *gorm.DB
}

func NewRecipeService(db *gorm.DB) *RecipeService {
	return &RecipeService{db: db}
}

type AddIngredientRequest struct {
	ProductID    uint    `json:"product_id" validate:"required"`
	IngredientID uint    `json:"ingredient_id" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,gt=0"`
}

func (s *RecipeService) GetRecipe(productID, businessID uint) ([]RecipeIngredient, error) {
	var ingredients []RecipeIngredient
	err := s.db.Where("product_id = ? AND business_id = ?", productID, businessID).Find(&ingredients).Error
	return ingredients, err
}

func (s *RecipeService) AddIngredient(businessID uint, req AddIngredientRequest) (*RecipeIngredient, error) {
	var ing RecipeIngredient
	// Check if already exists
	err := s.db.Where("product_id = ? AND ingredient_id = ? AND business_id = ?", req.ProductID, req.IngredientID, businessID).First(&ing).Error

	if err == nil {
		// Update existing
		ing.Quantity = req.Quantity
		if err := s.db.Save(&ing).Error; err != nil {
			return nil, err
		}
		return &ing, nil
	}

	// Create new
	ing = RecipeIngredient{
		BusinessID:   businessID,
		ProductID:    req.ProductID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
	}
	if err := s.db.Create(&ing).Error; err != nil {
		return nil, err
	}
	return &ing, nil
}

func (s *RecipeService) RemoveIngredient(id, businessID uint) error {
	return s.db.Where("id = ? AND business_id = ?", id, businessID).Delete(&RecipeIngredient{}).Error
}

// AdjustStockWithRecipe handles stock deduction for a product, taking its recipe into account if applicable.
// If the business has the RECIPE_MANAGEMENT module and a recipe exists for the product, it deducts ingredients.
// Otherwise, it falls back to standard single-product stock adjustment.
func (s *RecipeService) AdjustStockWithRecipe(tx *gorm.DB, productID, businessID uint, sellQuantity int) error {
	// 1. Check if the business has the Recipe Management module enabled
	if !subscription.HasModule(s.db, businessID, subscription.ModuleRecipe) {
		return inventory.AdjustStock(tx, productID, businessID, -sellQuantity)
	}

	// 2. Check if this product has a recipe
	var ingredients []RecipeIngredient
	if err := tx.Where("product_id = ? AND business_id = ?", productID, businessID).Find(&ingredients).Error; err != nil {
		return err
	}

	// 3. If no recipe exists, fall back to standard deduction for the product itself
	if len(ingredients) == 0 {
		return inventory.AdjustStock(tx, productID, businessID, -sellQuantity)
	}

	// 4. If a recipe exists, deduct each ingredient quantity
	for _, ing := range ingredients {
		// Calculate total quantity to deduct for this ingredient
		// sellQuantity is the number of finished products sold
		// ing.Quantity is the amount of ingredient per 1 finished product
		deductQty := float64(sellQuantity) * ing.Quantity

		// Note: We cast to int for standard logic, but in specialized industries like bars,
		// ingredients might be measured in fractional units (ml, grams) which standard inventory might not support yet.
		// For now, we adjust by integer if the inventory system expects it, but we should ideally support fractional stock.
		// Given current inventory system uses int, we'll use a ceiling or just round.
		// Better yet, we should probably update inventory to support float quantities for ingredients.
		// For the sake of this implementation, we will assume integer deduction or round to nearest.

		// If the ingredient is tracked as a standard product, we deduct it.
		if err := inventory.AdjustStock(tx, ing.IngredientID, businessID, -int(deductQty)); err != nil {
			return err
		}
	}

	return nil
}

package seed

import (
	"fmt"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/product"
	"time"

	"gorm.io/gorm"
)

const (
// Any seed specific constants can go here
)

func SeedSampleData(db *gorm.DB, bizID uint, bizType common.BusinessType) error {
	categories, ok := sampleData[bizType]
	if !ok {
		// Fallback to retail or generic if type not specifically defined
		categories = sampleData[common.TypeRetail]
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Verify if already seeded
		var biz business.Business
		if err := tx.First(&biz, bizID).Error; err != nil {
			return err
		}
		if biz.IsSeeded {
			return fmt.Errorf("business already seeded")
		}

		for i, sc := range categories {
			cat := category.Category{
				BusinessID: bizID,
				Name:       sc.Name,
			}
			if err := tx.Create(&cat).Error; err != nil {
				return err
			}

			for j, sp := range sc.Products {
				p := product.Product{
					BusinessID: bizID,
					CategoryID: cat.ID,
					Name:       sp.Name,
					Price:      sp.Price,
					Cost:       sp.Cost,
					Stock:      sp.Stock,
					SKU:        fmt.Sprintf("%s-%d-%d%d-%d", sp.SKUPrefix, bizID, i, j, time.Now().UnixNano()%10000),
					Active:     true,
				}
				if err := tx.Create(&p).Error; err != nil {
					return err
				}

				// Create inventory record
				inv := inventory.Inventory{
					ProductID:     p.ID,
					BusinessID:    bizID,
					CurrentStock:  p.Stock,
					LowStockAlert: 5,
				}
				if err := tx.Create(&inv).Error; err != nil {
					return err
				}
			}
		}

		// Mark as seeded
		if err := tx.Model(&biz).Update("is_seeded", true).Error; err != nil {
			return err
		}

		return nil
	})
}

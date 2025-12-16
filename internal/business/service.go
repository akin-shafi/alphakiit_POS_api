package business

import (
	"errors"
	"log"

	// "time"

	"gorm.io/gorm"
)

// CreateBusiness creates a new business under the user's tenant
func CreateBusiness(db *gorm.DB, tenantID, name string, bizType BusinessType) (*Business, error) {
	biz := &Business{
		TenantID: tenantID,
		Name:     name,
		Type:     bizType,
	}

	return biz, db.Create(biz).Error
}

// ListBusinesses returns all businesses for a tenant
func ListBusinesses(db *gorm.DB, tenantID string) ([]Business, error) {
	var businesses []Business
	err := db.Where("tenant_id = ?", tenantID).Find(&businesses).Error
	return businesses, err
}

// GetBusiness by ID (with tenant check)
func GetBusiness(db *gorm.DB, id uint, tenantID string) (*Business, error) {
	var biz Business
	err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&biz).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("business not found")
		}
		return nil, err
	}
	return &biz, nil
}

// UpdateBusiness name/type/address
func UpdateBusiness(db *gorm.DB, id uint, tenantID string, updates map[string]interface{}) (*Business, error) {
	var biz Business
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&biz).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("business not found")
		}
		return nil, err
	}

	if err := db.Model(&biz).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &biz, nil
}

func DeleteBusiness(db *gorm.DB, id uint, tenantID string) error {
	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&Business{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("business not found")
	}
	return nil
}

func SeedNewBusiness(db *gorm.DB, biz *Business) {
	template, ok := SeedTemplates[biz.Type]
	if !ok {
		return // No template
	}

	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Seeding failed (tx begin): %v", tx.Error)
		return
	}

	// Create categories
	catMap := make(map[int]uint)
	for i, cat := range template.Categories {
		cat.BusinessID = biz.ID
		if err := tx.Create(&cat).Error; err != nil {
			tx.Rollback()
			log.Printf("Seeding categories failed: %v", err)
			return
		}
		catMap[i] = cat.ID
	}

	// Create products + inventory
	numCats := len(template.Categories)
	for i, prod := range template.Products {
		prod.BusinessID = biz.ID
		prod.CategoryID = catMap[i%numCats] // distribute
		prod.Active = true

		if err := tx.Create(&prod).Error; err != nil {
			tx.Rollback()
			log.Printf("Seeding products failed: %v", err)
			return
		}

		if i < len(template.Inventory) {
			inv := template.Inventory[i]
			inv.ProductID = prod.ID
			inv.BusinessID = biz.ID
			if err := tx.Create(&inv).Error; err != nil {
				tx.Rollback()
				log.Printf("Seeding inventory failed: %v", err)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Seeding commit failed: %v", err)
	}
}

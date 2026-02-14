package business

import (
	"errors"

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

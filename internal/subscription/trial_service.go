package subscription

import (
	"time"

	"gorm.io/gorm"
)

// EvaluateTrialActivation checks the progress of a business trial and updates its status
func EvaluateTrialActivation(db *gorm.DB, businessID uint) error {
	// Local struct to scan checklist data without importing business package
	var checklist struct {
		ID                    uint
		BusinessID            uint
		BusinessInfoCompleted bool
		ProductsAddedCount    int
		CashierCreated        bool
		FirstSaleRecorded     bool
	}

	err := db.Table("trial_checklists").Where("business_id = ?", businessID).First(&checklist).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		// Initialize checklist if missing
		if err := db.Table("trial_checklists").Create(map[string]interface{}{
			"business_id": businessID,
		}).Error; err != nil {
			return err
		}
		checklist.BusinessID = businessID
	}

	// 1. Check Business Info (Day 0)
	var biz struct {
		ID             uint
		Name           string
		Address        string
		City           string
		TenantID       string
		TrialActivated bool
	}
	if err := db.Table("businesses").Where("id = ?", businessID).First(&biz).Error; err != nil {
		return err
	}

	updatedChecklist := make(map[string]interface{})

	if biz.Name != "" && biz.Address != "" && biz.City != "" {
		updatedChecklist["business_info_completed"] = true
		checklist.BusinessInfoCompleted = true
	}

	// 2. Check Products (Day 1)
	var productCount int64
	db.Table("products").Where("business_id = ? AND deleted_at IS NULL", businessID).Count(&productCount)
	updatedChecklist["products_added_count"] = int(productCount)
	checklist.ProductsAddedCount = int(productCount)

	// 3. Check Staff (Day 2)
	var staffCount int64
	db.Table("users").Where("tenant_id = ? AND role IN ? AND active = ?", biz.TenantID, []string{"CASHIER", "MANAGER"}, true).Count(&staffCount)
	if staffCount > 0 {
		updatedChecklist["cashier_created"] = true
		checklist.CashierCreated = true
	}

	// 4. Check Sales (Day 3)
	var saleCount int64
	db.Table("sales").Where("business_id = ? AND status = ? AND deleted_at IS NULL", businessID, "COMPLETED").Count(&saleCount)
	if saleCount > 0 {
		updatedChecklist["first_sale_recorded"] = true
		checklist.FirstSaleRecorded = true
	}

	// Update checklist
	if len(updatedChecklist) > 0 {
		if err := db.Table("trial_checklists").Where("business_id = ?", businessID).Updates(updatedChecklist).Error; err != nil {
			return err
		}
	}

	// Determine if trial should be activated
	// Mandatory: Business Info, At least 5 products, 1 staff, 1 sale
	if checklist.BusinessInfoCompleted && checklist.ProductsAddedCount >= 5 && checklist.CashierCreated && checklist.FirstSaleRecorded {
		if !biz.TrialActivated {
			now := time.Now()
			if err := db.Table("businesses").Where("id = ?", businessID).Updates(map[string]interface{}{
				"trial_activated": true,
				"activated_at":    &now,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

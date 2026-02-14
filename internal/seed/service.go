package seed

import (
	"fmt"
	"pos-fiber-app/internal/category"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/inventory"
	"pos-fiber-app/internal/product"
	"pos-fiber-app/internal/subscription"
	"time"

	"gorm.io/gorm"
)

func SeedInstallerData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Default Commission Settings if none exist
		var count int64
		tx.Model(&subscription.CommissionSetting{}).Count(&count)
		if count == 0 {
			settings := subscription.CommissionSetting{
				OnboardingRate:          20.0,
				RenewalRate:             10.0,
				EnableRenewalCommission: true,
				MinRenewalDays:          0,
				CommissionDurationDays:  0,
			}
			tx.Create(&settings)
		}

		// 2. Sample Training Resources
		var trCount int64
		tx.Model(&subscription.TrainingResource{}).Count(&trCount)
		if trCount == 0 {
			resources := []subscription.TrainingResource{
				{
					Title:       "BETADAY POS: Full Installer Onboarding",
					Description: "Comprehensive guide on how to set up businesses, manage hardware, and secure your commissions.",
					URL:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ", // Placeholder
					Type:        "VIDEO",
					IsActive:    true,
				},
				{
					Title:       "Partner Commission Policy 2026",
					Description: "Official document outlining the payout structure, minimum balances, and renewal terms.",
					URL:         "https://example.com/policy.pdf",
					Type:        "PDF",
					IsActive:    true,
				},
				{
					Title:       "How to Use the Mobile App Offline",
					Description: "Learn how to guide retailers through offline transaction management and syncing.",
					URL:         "https://www.youtube.com/watch?v=example",
					Type:        "VIDEO",
					IsActive:    true,
				},
			}
			tx.Create(&resources)
		}

		// 3. Sample Installers and Mock Commissions (if any exist)
		var installers []struct {
			ID   uint
			Role string
		}
		tx.Table("users").Where("role = ?", "INSTALLER").Find(&installers)

		if len(installers) > 0 {
			installer := installers[0]
			// Create a mock referral code for this installer
			var refCode subscription.ReferralCode
			if err := tx.Where("installer_id = ?", installer.ID).First(&refCode).Error; err != nil {
				refCode = subscription.ReferralCode{
					Code:        "BETA" + fmt.Sprint(installer.ID) + "PRO",
					InstallerID: installer.ID,
					IsActive:    true,
				}
				tx.Create(&refCode)
			}

			// Create some mock commissions if none exist
			var commCount int64
			tx.Model(&subscription.CommissionRecord{}).Where("installer_id = ?", installer.ID).Count(&commCount)
			if commCount == 0 {
				commissions := []subscription.CommissionRecord{
					{
						InstallerID: installer.ID,
						BusinessID:  1, // Mock business ID
						Amount:      5000.0,
						Type:        "ONBOARDING",
						Status:      subscription.CommissionPending,
					},
					{
						InstallerID: installer.ID,
						BusinessID:  2,
						Amount:      2500.0,
						Type:        "RENEWAL",
						Status:      subscription.CommissionPending,
					},
					{
						InstallerID: installer.ID,
						BusinessID:  1,
						Amount:      5000.0,
						Type:        "ONBOARDING",
						Status:      subscription.CommissionPaid,
						PaidAt:      &[]time.Time{time.Now()}[0],
					},
				}
				tx.Create(&commissions)
			}
		}

		return nil
	})
}

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
		// Verify if already seeded (use anonymous struct to break import cycle)
		var biz struct {
			ID       uint `gorm:"primaryKey"`
			IsSeeded bool
		}
		if err := tx.Table("businesses").First(&biz, bizID).Error; err != nil {
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
		if err := tx.Table("businesses").Where("id = ?", bizID).Update("is_seeded", true).Error; err != nil {
			return err
		}

		return nil
	})
}

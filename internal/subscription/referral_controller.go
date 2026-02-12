package subscription

import (
	"crypto/rand"
	"encoding/hex"
	"pos-fiber-app/internal/types"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateReferralCodeHandler allows an installer to generate a unique code
func CreateReferralCodeHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)

		// Ensure user is an INSTALLER or ADMIN
		if claims.Role != "INSTALLER" && claims.Role != "ADMIN" && claims.Role != "super_admin" && claims.Role != "SUPER_ADMIN" {
			return c.Status(403).JSON(fiber.Map{"error": "Only installers can generate referral codes"})
		}

		// Generate a random code if not provided
		var req struct {
			Code string `json:"code"`
		}
		c.BodyParser(&req)

		code := req.Code
		if code == "" {
			b := make([]byte, 4)
			rand.Read(b)
			code = "AB-" + hex.EncodeToString(b)
		}

		referral := ReferralCode{
			Code:        code,
			InstallerID: claims.UserID,
			IsActive:    true,
		}

		if err := db.Create(&referral).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create referral code", "details": err.Error()})
		}

		return c.Status(201).JSON(referral)
	}
}

// GetMyReferralCodesHandler returns all codes belonging to the current installer
func GetMyReferralCodesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)

		var codes []ReferralCode
		if err := db.Where("installer_id = ?", claims.UserID).Find(&codes).Error; err != nil {
			return err
		}

		return c.JSON(codes)
	}
}

// GetMyCommissionsHandler returns all commission records for the current installer
func GetMyCommissionsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*types.UserClaims)

		var commissions []CommissionRecord
		if err := db.Where("installer_id = ?", claims.UserID).Order("created_at DESC").Find(&commissions).Error; err != nil {
			return err
		}

		// Calculate total earned
		var totalEarned float64
		var pendingAmount float64
		for _, com := range commissions {
			if com.Status == CommissionPaid {
				totalEarned += com.Amount
			} else if com.Status == CommissionPending {
				pendingAmount += com.Amount
			}
		}

		return c.JSON(fiber.Map{
			"commissions":   commissions,
			"total_paid":    totalEarned,
			"total_pending": pendingAmount,
		})
	}
}

// AdminListAllCommissionsHandler allows admins to see all commissions
func AdminListAllCommissionsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var commissions []CommissionRecord
		if err := db.Order("created_at DESC").Find(&commissions).Error; err != nil {
			return err
		}
		return c.JSON(commissions)
	}
}

// AdminUpdateCommissionStatusHandler allows admins to mark commissions as PAID
func AdminUpdateCommissionStatusHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req struct {
			Status CommissionStatus `json:"status"`
		}
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		var com CommissionRecord
		if err := db.First(&com, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Commission record not found"})
		}

		com.Status = req.Status
		if req.Status == CommissionPaid {
			now := time.Now()
			com.PaidAt = &now
		}

		if err := db.Save(&com).Error; err != nil {
			return err
		}

		return c.JSON(com)
	}
}

// GetCommissionSettingsHandler returns the global commission settings
func GetCommissionSettingsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var settings CommissionSetting
		// Find the first record or create a default one
		if err := db.First(&settings).Error; err != nil {
			if gorm.ErrRecordNotFound == err {
				settings = CommissionSetting{
					OnboardingRate:          20.0,
					RenewalRate:             10.0,
					EnableRenewalCommission: true,
					MinRenewalDays:          0,
					CommissionDurationDays:  0,
				}
				db.Create(&settings)
			} else {
				return err
			}
		}
		return c.JSON(settings)
	}
}

// UpdateCommissionSettingsHandler updates the global commission settings
func UpdateCommissionSettingsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CommissionSetting
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		var settings CommissionSetting
		if err := db.First(&settings).Error; err != nil {
			// If not found, Create it
			if err := db.Create(&req).Error; err != nil {
				return err
			}
			return c.JSON(req)
		}

		// Update fields using map to ensure zero values (false/0) are updated
		updates := map[string]interface{}{
			"onboarding_rate":           req.OnboardingRate,
			"renewal_rate":              req.RenewalRate,
			"enable_renewal_commission": req.EnableRenewalCommission,
			"min_renewal_days":          req.MinRenewalDays,
			"commission_duration_days":  req.CommissionDurationDays,
		}

		if err := db.Model(&settings).Updates(updates).Error; err != nil {
			return err
		}

		// Reload to get latest
		db.First(&settings)

		return c.JSON(settings)
	}
}

// --- Training Resource Handlers ---

// AdminCreateTrainingResourceHandler allows admins to add new training material
func AdminCreateTrainingResourceHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var res TrainingResource
		if err := c.BodyParser(&res); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
		}

		if err := db.Create(&res).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create resource"})
		}

		return c.Status(201).JSON(res)
	}
}

// AdminListTrainingResourcesHandler returns all resources for admin management
func AdminListTrainingResourcesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resources []TrainingResource
		if err := db.Order("created_at DESC").Find(&resources).Error; err != nil {
			return err
		}
		return c.JSON(resources)
	}
}

// AdminUpdateTrainingResourceHandler allows updating an existing resource
func AdminUpdateTrainingResourceHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req TrainingResource
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		var res TrainingResource
		if err := db.First(&res, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Resource not found"})
		}

		if err := db.Model(&res).Updates(req).Error; err != nil {
			return err
		}

		return c.JSON(res)
	}
}

// AdminDeleteTrainingResourceHandler allows deleting a resource
func AdminDeleteTrainingResourceHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := db.Delete(&TrainingResource{}, id).Error; err != nil {
			return err
		}
		return c.Status(204).Send(nil)
	}
}

// GetTrainingResourcesHandler returns active resources for installers
func GetTrainingResourcesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resources []TrainingResource
		if err := db.Where("is_active = ?", true).Order("created_at DESC").Find(&resources).Error; err != nil {
			return err
		}
		return c.JSON(resources)
	}
}

// RequestPayoutHandler allows an installer to request a payout of their commission
func RequestPayoutHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*types.UserClaims)

		// Use local struct to avoid import cycle with user package
		type installerUser struct {
			ID            uint
			BankName      string
			AccountNumber string
			AccountName   string
		}

		var u installerUser
		// Query the 'users' table directly to fetch bank details
		if err := db.Table("users").First(&u, userCtx.UserID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		if u.AccountNumber == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Please set up your bank details in profile first"})
		}

		// Calculate total pending commission
		var total float64
		db.Model(&CommissionRecord{}).
			Where("installer_id = ? AND status = ?", userCtx.UserID, CommissionPending).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&total)

		if total < 5000 { // Min payout 5k
			return c.Status(400).JSON(fiber.Map{"error": "Minimum payout balance is ₦5,000"})
		}

		// Check if there is already a pending request
		var existingCount int64
		db.Model(&PayoutRequest{}).Where("installer_id = ? AND status = ?", userCtx.UserID, PayoutRequested).Count(&existingCount)
		if existingCount > 0 {
			return c.Status(400).JSON(fiber.Map{"error": "You already have a pending payout request"})
		}

		payout := PayoutRequest{
			InstallerID: userCtx.UserID,
			Amount:      total,
			BankName:    u.BankName,
			AccountNum:  u.AccountNumber,
			AccountName: u.AccountName,
			Status:      PayoutRequested,
		}

		if err := db.Create(&payout).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create payout request"})
		}

		return c.Status(201).JSON(payout)
	}
}

// GetPayoutRequestsHandler returns payout history for installer
func GetPayoutRequestsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userCtx := c.Locals("user").(*types.UserClaims)
		var payouts []PayoutRequest
		db.Where("installer_id = ?", userCtx.UserID).Order("created_at DESC").Find(&payouts)
		return c.JSON(payouts)
	}
}

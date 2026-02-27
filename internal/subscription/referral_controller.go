package subscription

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/types"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// InternalUser is a local definition to avoid import cycle with internal/user package
type InternalUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Password  string    `json:"-"`
	Active    bool      `json:"active"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name for InternalUser
func (InternalUser) TableName() string {
	return "users"
}

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

// AdminListAffiliatesHandler returns all users with role INSTALLER
func AdminListAffiliatesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var followers []InternalUser
		if err := db.Where("role = ?", "INSTALLER").Find(&followers).Error; err != nil {
			return err
		}

		// Map to a cleaner response with referral codes
		type AffiliateResp struct {
			InternalUser
			ReferralCodes []ReferralCode `json:"referral_codes"`
		}

		var resp []AffiliateResp
		for _, u := range followers {
			var codes []ReferralCode
			db.Where("installer_id = ?", u.ID).Find(&codes)
			resp = append(resp, AffiliateResp{
				InternalUser:  u,
				ReferralCodes: codes,
			})
		}

		return c.JSON(resp)
	}
}

// AdminCreateInfluencerHandler creates a new installer user, referral code, and matching promo code
func AdminCreateInfluencerHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			FirstName             string  `json:"first_name" validate:"required"`
			LastName              string  `json:"last_name" validate:"required"`
			Email                 string  `json:"email" validate:"required,email"`
			OnlineName            string  `json:"online_name" validate:"required"` // e.g. TAOOMA
			CommissionRate        float64 `json:"commission_rate"`                 // onboarding %
			RenewalCommissionRate float64 `json:"renewal_commission_rate"`         // renewal %
			DiscountPercentage    float64 `json:"discount_percentage"`             // fan discount %
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Defaults
		if req.CommissionRate <= 0 {
			req.CommissionRate = 20.0
		}
		if req.RenewalCommissionRate <= 0 {
			req.RenewalCommissionRate = 10.0
		}
		if req.DiscountPercentage <= 0 {
			req.DiscountPercentage = 20.0
		}

		// Generate a random password
		bytes := make([]byte, 4)
		rand.Read(bytes)
		tempPassword := hex.EncodeToString(bytes) // e.g. "a1b2c3d4"

		// Use bcrypt directly to avoid import cycle
		hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), 14)
		hashed := string(hashedBytes)

		tx := db.Begin()

		// 1. Create User
		u := InternalUser{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  hashed,
			Role:      "INSTALLER",
			Active:    true,
		}

		if err := tx.Create(&u).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"error": "Email already exists or invalid"})
		}

		// 2. Create Referral Code
		ref := ReferralCode{
			Code:                     req.OnlineName,
			InstallerID:              u.ID,
			OnboardingCommissionRate: req.CommissionRate,
			RenewalCommissionRate:    req.RenewalCommissionRate,
			IsActive:                 true,
		}
		if err := tx.Create(&ref).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"error": "Online name (Code) already taken"})
		}

		// 3. Create Matching Promo Code (Discount for fans)
		promo := PromoCode{
			Code:               req.OnlineName,
			DiscountPercentage: req.DiscountPercentage,
			MaxUses:            0,                            // Unlimited
			ExpiryDate:         time.Now().AddDate(10, 0, 0), // 10 years
			Active:             true,
		}
		if err := tx.Create(&promo).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"error": "Failed to create matching promo code"})
		}

		tx.Commit()

		// 4. Send Email (Fire and forget)
		go func() {
			sender := email.NewSender(email.LoadConfig())
			subject := "Your Influencer/Affiliate Account Credentials"
			body := fmt.Sprintf(`
				<h1>Welcome to AB-POS Platform</h1>
				<p>Hello %s, your influencer account has been created.</p>
				<p><b>Your Personal Referral & Promo Code:</b> %s</p>
				<p>Businesses using this code get %.0f%% discount, and you get %.0f%% commission.</p>
				<hr/>
				<p><b>Login Credentials:</b></p>
				<p>Email: %s</p>
				<p>Password: %s</p>
				<p>Login here: <a href="https://app.ab-pos.com/auth/login">Affiliate Portal</a></p>
			`, req.FirstName, req.OnlineName, req.DiscountPercentage, req.CommissionRate, req.Email, tempPassword)
			sender.SendCustomEmail(req.Email, subject, body)
		}()

		return c.Status(201).JSON(fiber.Map{
			"message":       "Influencer created successfully",
			"user_id":       u.ID,
			"code":          req.OnlineName,
			"temp_password": tempPassword,
		})
	}
}

// AdminUpdateAffiliateHandler updates influencer profile and their codes
func AdminUpdateAffiliateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req struct {
			FirstName             string  `json:"first_name"`
			LastName              string  `json:"last_name"`
			Active                *bool   `json:"active"`
			CommissionRate        float64 `json:"commission_rate"`
			RenewalCommissionRate float64 `json:"renewal_commission_rate"`
			DiscountPercentage    float64 `json:"discount_percentage"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		tx := db.Begin()

		// 1. Update User
		var u InternalUser
		if err := tx.First(&u, id).Error; err != nil {
			tx.Rollback()
			return c.Status(404).JSON(fiber.Map{"error": "Influencer not found"})
		}

		if req.FirstName != "" {
			u.FirstName = req.FirstName
		}
		if req.LastName != "" {
			u.LastName = req.LastName
		}
		if req.Active != nil {
			u.Active = *req.Active
		}

		tx.Save(&u)

		// 2. Update Referral Code & Promo Code
		var ref ReferralCode
		if err := tx.Where("installer_id = ?", u.ID).First(&ref).Error; err == nil {
			if req.CommissionRate > 0 {
				ref.OnboardingCommissionRate = req.CommissionRate
			}
			if req.RenewalCommissionRate > 0 {
				ref.RenewalCommissionRate = req.RenewalCommissionRate
			}
			if req.Active != nil {
				ref.IsActive = *req.Active
			}
			tx.Save(&ref)

			// Promo Code (matching the same code string)
			var promo PromoCode
			if err := tx.Where("code = ?", ref.Code).First(&promo).Error; err == nil {
				if req.DiscountPercentage > 0 {
					promo.DiscountPercentage = req.DiscountPercentage
				}
				if req.Active != nil {
					promo.Active = *req.Active
				}
				tx.Save(&promo)
			}
		}

		tx.Commit()
		return c.JSON(fiber.Map{"message": "Influencer updated successfully"})
	}
}

// AdminDeleteAffiliateHandler wipes an influencer and their codes (or deactivates)
func AdminDeleteAffiliateHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		tx := db.Begin()

		var u InternalUser
		if err := tx.First(&u, id).Error; err != nil {
			tx.Rollback()
			return c.Status(404).JSON(fiber.Map{"error": "Influencer not found"})
		}

		// Delete their marks
		tx.Where("installer_id = ?", u.ID).Delete(&ReferralCode{})

		// Note: We don't delete PromoCodes as businesses might already be using them
		// But we can deactivate them
		var refs []ReferralCode
		db.Where("installer_id = ?", u.ID).Find(&refs)
		for _, r := range refs {
			tx.Model(&PromoCode{}).Where("code = ?", r.Code).Update("active", false)
		}

		tx.Delete(&u)
		tx.Commit()

		return c.JSON(fiber.Map{"message": "Influencer deleted and codes deactivated"})
	}
}

// AdminGetAffiliateStatsHandler returns performance metrics for an influencer
func AdminGetAffiliateStatsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var stats struct {
			TotalFans      int64           `json:"total_fans"`
			TotalEarned    float64         `json:"total_earned"`
			PendingPayout  float64         `json:"pending_payout"`
			PayoutHistory  []PayoutRequest `json:"payout_history"`
			RecentOnboards []struct {
				ID        uint      `json:"id"`
				Name      string    `json:"name"`
				CreatedAt time.Time `json:"onboarded_at"`
			} `json:"recent_onboards"`
		}

		// 1. Total Fans (Businesses linked)
		db.Table("businesses").Where("installer_id = ?", id).Count(&stats.TotalFans)

		// 2. Earnings
		db.Model(&CommissionRecord{}).Where("installer_id = ? AND status = ?", id, CommissionPaid).Select("SUM(amount)").Scan(&stats.TotalEarned)
		db.Model(&CommissionRecord{}).Where("installer_id = ? AND status = ?", id, CommissionPending).Select("SUM(amount)").Scan(&stats.PendingPayout)

		// 3. Payouts
		db.Where("installer_id = ?", id).Order("created_at DESC").Find(&stats.PayoutHistory)

		// 4. Recent Businesses
		db.Table("businesses").Where("installer_id = ?", id).Order("created_at DESC").Limit(5).Scan(&stats.RecentOnboards)

		return c.JSON(stats)
	}
}

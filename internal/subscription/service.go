package subscription

import (
	"errors"
	"pos-fiber-app/internal/common"
	"time"

	"gorm.io/gorm"
)

const (
	OnboardingCommissionRate = 0.20 // 20%
	RenewalCommissionRate    = 0.10 // 10%
)

func HandleCommission(db *gorm.DB, sub *Subscription) error {
	// 0. Preliminary checks
	if sub == nil || sub.AmountPaid <= 0 {
		return nil
	}

	// Fetch dynamic settings
	var settings CommissionSetting
	if err := db.First(&settings).Error; err != nil {
		// Fallback to default if settings not found
		settings = CommissionSetting{
			OnboardingRate:         20.0,
			RenewalRate:            10.0,
			MinRenewalDays:         0,
			CommissionDurationDays: 0,
		}
	}

	// Prevent duplicate commissions for the same subscription
	var existing int64
	db.Model(&CommissionRecord{}).Where("subscription_id = ?", sub.ID).Count(&existing)
	if existing > 0 {
		return nil
	}

	// 1. Get the business to find the installer and check activation status
	var biz struct {
		ID               uint
		InstallerID      *uint
		TrialActivated   bool
		DefaultPromoCode string
	}
	if err := db.Table("businesses").Where("id = ?", sub.BusinessID).First(&biz).Error; err != nil {
		return err
	}

	if biz.InstallerID == nil {
		return nil // No installer associated
	}

	// Try to find specific commission rates from the ReferralCode
	if biz.DefaultPromoCode != "" {
		var refCode ReferralCode
		if err := db.Where("code = ? AND installer_id = ?", biz.DefaultPromoCode, *biz.InstallerID).First(&refCode).Error; err == nil {
			if refCode.OnboardingCommissionRate > 0 {
				settings.OnboardingRate = refCode.OnboardingCommissionRate
			}
			if refCode.RenewalCommissionRate > 0 {
				settings.RenewalRate = refCode.RenewalCommissionRate
			}
		}
	}

	// Installer commission is ONLY valid for Activated Trials.
	if !biz.TrialActivated {
		return nil // Merchant hasn't reached activation threshold yet
	}

	// 2. Determine if this is ONBOARDING or RENEWAL
	// Check how many successful paid subscriptions this business had *before* this one
	var count int64
	db.Model(&Subscription{}).
		Where("business_id = ? AND amount_paid > 0 AND id < ?", sub.BusinessID, sub.ID).
		Count(&count)

	commissionType := "RENEWAL"
	commissionRate := settings.RenewalRate / 100.0

	// Handle Onboarding
	if count == 0 {
		commissionType = "ONBOARDING"
		commissionRate = settings.OnboardingRate / 100.0
	} else {
		// Handle Renewal specific conditions
		if !settings.EnableRenewalCommission {
			return nil // Renewal commission is disabled globally
		}

		// Condition A: Minimum Plan Duration (e.g., must be 1 year)
		if settings.MinRenewalDays > 0 {
			// Calculate duration of current sub
			durationDays := int(sub.EndDate.Sub(sub.StartDate).Hours() / 24)
			if durationDays < settings.MinRenewalDays {
				return nil // Does not meet minimum duration requirement
			}
		}

		// Condition B: Commission Duration Limit (e.g., only for the first year of the business)
		if settings.CommissionDurationDays > 0 {
			// Find the first ever commission record for this business
			var firstCommission CommissionRecord
			err := db.Where("business_id = ? AND type = 'ONBOARDING'", sub.BusinessID).First(&firstCommission).Error
			if err == nil {
				limitDate := firstCommission.CreatedAt.AddDate(0, 0, settings.CommissionDurationDays)
				if time.Now().After(limitDate) {
					return nil // Commission period has expired
				}
			}
		}
	}

	amount := sub.AmountPaid * commissionRate
	if amount <= 0 {
		return nil
	}

	// 3. Create commission record
	record := CommissionRecord{
		InstallerID:    *biz.InstallerID,
		BusinessID:     biz.ID,
		SubscriptionID: sub.ID,
		Amount:         amount,
		Type:           commissionType,
		Status:         CommissionPending,
	}

	return db.Create(&record).Error
}

func GetSubscriptionStatus(db *gorm.DB, businessID uint) (*Subscription, error) {
	var sub Subscription
	// Get the latest active or grace period subscription
	err := db.Where("business_id = ? AND status IN ?", businessID, []SubscriptionStatus{StatusActive, StatusGracePeriod}).
		Order("end_date DESC").
		First(&sub).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, err
}

// GetProductLimit returns the maximum allowed products for a business
func GetProductLimit(db *gorm.DB, businessID uint) int {
	sub, err := GetSubscriptionStatus(db, businessID)
	if err != nil || sub == nil {
		return 10 // default very low limit for unsubscribed
	}

	for _, p := range AvailablePlans {
		if p.Type == sub.PlanType {
			return p.ProductLimit
		}
	}

	return 10 // fallback
}

// GetRemainingDays calculates the number of full days remaining until expiry
func GetRemainingDays(expiry time.Time) int {
	now := time.Now()
	if now.After(expiry) {
		return 0
	}
	days := int(expiry.Sub(now).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// CheckLaunchOfferEligibility returns whether a plan is eligible for the launch discount
// Logic: March 1 - March 14, 2026. Quarterly/Annual Plans.
func CheckLaunchOfferEligibility(planType PlanType) (bool, float64) {
	now := time.Now()
	// Promo Window: March 1, 2026 to March 14, 2026
	startDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC) // Includes March 14

	if now.Before(startDate) || now.After(endDate) {
		return false, 0
	}

	switch planType {
	case PlanQuarterly, PlanServiceQuarterly:
		return true, 20.0
	case PlanAnnual, PlanServiceAnnual:
		return true, 40.0
	}

	return false, 0
}

func CreateSubscription(db *gorm.DB, businessID uint, planType PlanType, paymentMethod, ref string, amount float64) (*Subscription, error) {
	var plan *SubscriptionPlan
	for _, p := range AvailablePlans {
		if p.Type == planType {
			plan = &p
			break
		}
	}

	if plan == nil {
		return nil, errors.New("invalid plan type")
	}

	// Eligibility Check
	if len(plan.AllowedBusinessTypes) > 0 {
		var bizType common.BusinessType
		db.Table("businesses").Select("type").Where("id = ?", businessID).Scan(&bizType)

		allowed := false
		for _, t := range plan.AllowedBusinessTypes {
			if t == bizType {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errors.New("business type not eligible for this plan")
		}
	}

	now := time.Now()
	var startDate = now
	var endDate = now.AddDate(0, 0, plan.DurationDays)

	// Check if there's an existing active subscription to extend
	currentSub, _ := GetSubscriptionStatus(db, businessID)
	if currentSub != nil && currentSub.EndDate.After(now) && currentSub.PlanType != PlanTrial {
		// If renewing the same plan type or upgrading, we extend from the current end date
		// Note: Trial is never extended, it's always replaced.
		endDate = currentSub.EndDate.AddDate(0, 0, plan.DurationDays)
	}

	sub := &Subscription{
		BusinessID:           businessID,
		PlanType:             planType,
		Status:               StatusActive,
		StartDate:            startDate,
		EndDate:              endDate,
		PaymentMethod:        paymentMethod,
		TransactionReference: ref,
		AmountPaid:           amount,
	}

	if err := db.Create(sub).Error; err != nil {
		return nil, err
	}

	// Trigger commission calculation
	HandleCommission(db, sub)

	return sub, nil
}

func CheckSubscriptionAccess(db *gorm.DB, businessID uint) (bool, SubscriptionStatus, error) {
	sub, err := GetSubscriptionStatus(db, businessID)
	if err != nil {
		return false, "", err
	}

	if sub == nil {
		return false, "NONE", nil
	}

	now := time.Now()
	if now.After(sub.EndDate) {
		// No grace period for trials - immediate lock
		if sub.PlanType == PlanTrial {
			if sub.Status != StatusExpired {
				sub.Status = StatusExpired
				db.Save(sub)
			}
			return false, StatusExpired, nil
		}

		// Calculate grace period (7 days) for paid plans
		graceExpiry := sub.EndDate.AddDate(0, 0, 7)
		if now.Before(graceExpiry) {
			if sub.Status != StatusGracePeriod {
				sub.Status = StatusGracePeriod
				db.Save(sub)
			}
			return true, StatusGracePeriod, nil // Still active but in grace period
		}

		// Truly expired
		if sub.Status != StatusExpired {
			sub.Status = StatusExpired
			db.Save(sub)
		}
		return false, StatusExpired, nil
	}

	return true, sub.Status, nil
}

func HasModule(db *gorm.DB, businessID uint, module ModuleType) bool {
	var busMod BusinessModule
	err := db.Where("business_id = ? AND module = ? AND is_active = ?", businessID, module, true).First(&busMod).Error
	if err != nil {
		// Only RecordNotFound is a "safe" error that doesn't necessarily mean the transaction is poisoned
		// though in some cases even this can poison.
		// For a more robust fix, we should check if we are in a transaction, but GORM doesn't make that easy.
		return false
	}

	// If there's an expiry date, check it
	if busMod.ExpiryDate != nil && time.Now().After(*busMod.ExpiryDate) {
		return false
	}

	return true
}

// HasModuleWithError is a transaction-safe version of HasModule
func HasModuleWithError(db *gorm.DB, businessID uint, module ModuleType) (bool, error) {
	var busMod BusinessModule
	err := db.Where("business_id = ? AND module = ? AND is_active = ?", businessID, module, true).First(&busMod).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if busMod.ExpiryDate != nil && time.Now().After(*busMod.ExpiryDate) {
		return false, nil
	}

	return true, nil
}

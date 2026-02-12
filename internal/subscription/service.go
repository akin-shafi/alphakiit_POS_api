package subscription

import (
	"errors"
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

	// 1. Get the business to find the installer
	var biz struct {
		ID          uint
		InstallerID *uint
	}
	if err := db.Table("businesses").Where("id = ?", sub.BusinessID).First(&biz).Error; err != nil {
		return err
	}

	if biz.InstallerID == nil {
		return nil // No installer associated
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
	err := db.Where("business_id = ? AND status IN ?", businessID, []SubscriptionStatus{StatusActive, StatusGracePeriod, StatusPendingPayment}).
		Order("end_date DESC").
		First(&sub).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, err
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

	now := time.Now()
	endDate := now.AddDate(0, 0, plan.DurationDays)

	sub := &Subscription{
		BusinessID:           businessID,
		PlanType:             planType,
		Status:               StatusActive,
		StartDate:            now,
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
		// Calculate grace period (7 days)
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
		return false
	}

	// If there's an expiry date, check it
	if busMod.ExpiryDate != nil && time.Now().After(*busMod.ExpiryDate) {
		return false
	}

	return true
}

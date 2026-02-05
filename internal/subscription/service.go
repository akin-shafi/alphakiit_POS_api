package subscription

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func GetSubscriptionStatus(db *gorm.DB, businessID uint) (*Subscription, error) {
	var sub Subscription
	err := db.Where("business_id = ? AND status IN ?", businessID, []SubscriptionStatus{StatusActive, StatusGracePeriod}).
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

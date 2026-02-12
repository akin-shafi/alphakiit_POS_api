package onboarding

import (
	"errors"
	"fmt"
	"log"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
	"pos-fiber-app/internal/seed"
	"pos-fiber-app/internal/subscription"
	"pos-fiber-app/internal/user"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func OnboardBusiness(
	db *gorm.DB,
	payload *OnboardingRequest,
) (*business.Business, *user.User, error) {

	var createdBusiness *business.Business
	var createdUser *user.User

	err := db.Transaction(func(tx *gorm.DB) error {
		tenantID := uuid.New().String()[:8]

		biz := &business.Business{
			TenantID: tenantID,
			Name:     strings.TrimSpace(payload.Business.Name),
			Type:     common.BusinessType(payload.Business.Type),
			Address:  payload.Business.Address,
			City:     payload.Business.City,
			Currency: common.Currency(payload.Business.Currency),
		}

		if err := tx.Create(biz).Error; err != nil {
			return err
		}

		// Handle Referral Token
		if payload.ReferralToken != "" {
			var refCode subscription.ReferralCode
			if err := tx.Where("code = ? AND is_active = ?", payload.ReferralToken, true).First(&refCode).Error; err == nil {
				biz.InstallerID = &refCode.InstallerID
				// Increment token usage
				tx.Model(&refCode).Update("uses_count", gorm.Expr("uses_count + 1"))
			}
		}

		hashed, err := user.HashPassword(payload.User.Password)
		if err != nil {
			return err
		}

		u := &user.User{
			FirstName: strings.TrimSpace(payload.User.FirstName),
			LastName:  strings.TrimSpace(payload.User.LastName),
			Email:     strings.ToLower(payload.User.Email),
			Password:  hashed,
			TenantID:  tenantID,
			Role:      "OWNER",
			Active:    true,
		}

		if err := tx.Create(u).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value") {
				return fmt.Errorf("email already exists")
			}
			return err
		}

		// if err := tx.Create(u).Error; err != nil {
		// 	return err
		// }

		// Issue 14-day trial
		planType := subscription.PlanTrial
		if payload.BasePlanType == string(subscription.PlanServiceMonthly) {
			planType = subscription.PlanServiceMonthly
		}

		sub, err := subscription.CreateSubscription(tx, biz.ID, planType, "SYSTEM", "INITIAL_TRIAL", 0)
		if err != nil {
			return err
		}

		// If user wants to skip trial and pay now, it remains in PENDING_PAYMENT or EXPIRED status immediately
		if payload.SkipTrial {
			sub.Status = subscription.StatusPendingPayment
			sub.EndDate = time.Now() // Expire immediately to force payment
			tx.Save(sub)
			biz.SubscriptionStatus = string(subscription.StatusPendingPayment)
		}

		// Handle Bundles
		if payload.BundleCode != "" {
			for _, b := range subscription.AvailableBundles {
				if b.Code == payload.BundleCode {
					for _, modType := range b.Modules {
						// Add bundle modules if not already selected
						found := false
						for _, m := range payload.Modules {
							if m == string(modType) {
								found = true
								break
							}
						}
						if !found {
							payload.Modules = append(payload.Modules, string(modType))
						}
					}
					break
				}
			}
		}

		// Handle Selected Modules
		moduleTrialExpiry := sub.EndDate
		for _, modName := range payload.Modules {
			// Validate if module exists
			isValid := false
			for _, am := range subscription.AvailableModules {
				if string(am.Type) == modName {
					isValid = true
					break
				}
			}

			if isValid {
				tx.Create(&subscription.BusinessModule{
					BusinessID: biz.ID,
					Module:     subscription.ModuleType(modName),
					IsActive:   true,
					ExpiryDate: &moduleTrialExpiry,
				})
			}
		}

		// Update business with expiry
		biz.SubscriptionExpiry = &sub.EndDate
		if err := tx.Save(biz).Error; err != nil {
			return err
		}

		createdBusiness = biz
		createdUser = u

		// Handle Sample Data Seeding
		if payload.UseSampleData {
			if err := seed.SeedSampleData(tx, biz.ID, biz.Type); err != nil {
				// We log the error but don't fail the whole onboarding?
				// Actually, if they asked for it and it fails, it might be better to know.
				// For now, let's log it.
				log.Printf("Failed to seed sample data for business %d: %v", biz.ID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	sender := email.NewSender(email.LoadConfig()) // or inject via dependency
	// fire-and-forget
	// fire-and-forget
	go func(emailStr, name string) {
		// Generate OTP
		code, err := otp.GenerateOTP()
		if err != nil {
			log.Printf("Failed to generate OTP: %v", err)
			return
		}

		// Store OTP
		otpEntry := otp.OTP{
			Email:     emailStr,
			Code:      code,
			Type:      otp.TypeVerification,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		if err := db.Create(&otpEntry).Error; err != nil {
			log.Printf("Failed to store OTP: %v", err)
			return
		}

		// Send Verification use code as URL param or just code?
		// User requested: "Automatically send an OTP" and "Verify ... using OTP".
		// So we send the code. logic in SendEmailVerification takes a "verificationURL".
		// We can repurpose it to just send the code, or construct a fake URL "Your OTP is: CODE"
		// Better: Update Sender to support sending OTP specifically for verification.
		// For now, let's use SendCustomEmail for flexibility or update Sender later.
		// Actually, I'll assume SendEmailVerification can handle just the code if I pass it,
		// but the template likely expects a link.
		// Let's use SendCustomEmail for now to be safe.

		body := fmt.Sprintf("<h1>Verify your email</h1><p>Your OTP Code is: <b>%s</b></p><p>It expires in 15 minutes.</p>", code)
		if err := sender.SendCustomEmail(emailStr, "Verify Your Email", body); err != nil {
			log.Printf("verification email failed: %v", err)
		}
	}(createdUser.Email, createdUser.FirstName)

	return createdBusiness, createdUser, nil
}

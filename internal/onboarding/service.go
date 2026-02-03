package onboarding

import (
	"errors"
	"fmt"
	"log"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/otp"
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
		sub, err := subscription.CreateSubscription(tx, biz.ID, subscription.PlanTrial, "SYSTEM", "INITIAL_TRIAL", 0)
		if err != nil {
			return err
		}

		// Update business with expiry
		biz.SubscriptionExpiry = &sub.EndDate
		if err := tx.Save(biz).Error; err != nil {
			return err
		}

		createdBusiness = biz
		createdUser = u
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

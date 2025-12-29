package onboarding

import (
	"errors"
	"fmt"
	"log"
	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/common"
	"pos-fiber-app/internal/email"
	"pos-fiber-app/internal/user"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// func OnboardBusiness(db *gorm.DB, payload *OnboardingRequest) (*business.Business, *user.User, error) {
// 	var createdBusiness *business.Business
// 	var createdUser *user.User

// 	err := db.Transaction(func(tx *gorm.DB) error {
// 		sender := email.NewSender(email.LoadConfig())
// 		tenantID := uuid.New().String()[:8]

// 		biz := &business.Business{
// 			TenantID: tenantID,
// 			Name:     payload.Business.Name,
// 			Type:     common.BusinessType(payload.Business.Type),
// 			Address:  payload.Business.Address,
// 			City:     payload.Business.City,
// 		}

// 		if err := tx.Create(biz).Error; err != nil {
// 			return err
// 		}

// 		hashed, err := user.HashPassword(payload.User.Password)
// 		if err != nil {
// 			return err
// 		}

// 		u := &user.User{
// 			FirstName: payload.User.FirstName,
// 			LastName:  payload.User.LastName,
// 			Email:     payload.User.Email,
// 			Password:  hashed,
// 			TenantID:  tenantID,
// 			Role:      "OWNER",
// 			Active:    true,
// 		}

// 		if err := tx.Create(u).Error; err != nil {
// 			return err
// 		}

// 		if err := sender.SendWelcomeEmail(payload.User.Email, payload.User.FirstName); err != nil {
// 			log.Printf("Failed to send welcome email: %v", err)
// 			// Still return success to avoid info disclosure
// 		}

// 		createdBusiness = biz
// 		createdUser = u
// 		return nil
// 	})

// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	return createdBusiness, createdUser, nil
// }

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
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("email already exists")
			}
			return err
		}

		// if err := tx.Create(u).Error; err != nil {
		// 	return err
		// }

		createdBusiness = biz
		createdUser = u
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	sender := email.NewSender(email.LoadConfig()) // or inject via dependency
	// fire-and-forget
	go func(email, name string) {
		if err := sender.SendWelcomeEmail(email, name); err != nil {
			log.Printf("welcome email failed: %v", err)
		}
	}(createdUser.Email, createdUser.FirstName)

	return createdBusiness, createdUser, nil
}

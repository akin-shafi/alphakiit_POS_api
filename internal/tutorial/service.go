package tutorial

import (
	"pos-fiber-app/internal/common"

	"gorm.io/gorm"
)

type TutorialService struct {
	db *gorm.DB
}

func NewTutorialService(db *gorm.DB) *TutorialService {
	return &TutorialService{db: db}
}

func (s *TutorialService) GetTutorialsByBusinessType(businessType common.BusinessType) ([]Tutorial, error) {
	var tutorials []Tutorial
	// Fetch tutorials for specific business type or "ALL" for general ones
	err := s.db.Where("business_type = ? OR business_type = ?", businessType, "ALL").
		Order("display_order asc").
		Find(&tutorials).Error
	return tutorials, err
}

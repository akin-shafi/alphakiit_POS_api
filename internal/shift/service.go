package shift

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type ShiftService struct {
	db *gorm.DB
}

func NewShiftService(db *gorm.DB) *ShiftService {
	return &ShiftService{db: db}
}

func (s *ShiftService) StartShift(businessID uint, userID uint, userName string, startCash float64) (*Shift, error) {
	// Check if user already has an open shift in this business
	var existing Shift
	err := s.db.Where("business_id = ? AND user_id = ? AND status = ?", businessID, userID, "open").First(&existing).Error
	if err == nil {
		return nil, errors.New("you already have an open shift")
	}

	shift := &Shift{
		BusinessID: businessID,
		UserID:     userID,
		UserName:   userName,
		StartTime:  time.Now(),
		StartCash:  startCash,
		Status:     "open",
	}

	if err := s.db.Create(shift).Error; err != nil {
		return nil, err
	}

	return shift, nil
}

func (s *ShiftService) EndShift(shiftID uint, endCash float64) (*Shift, error) {
	var shift Shift
	if err := s.db.First(&shift, shiftID).Error; err != nil {
		return nil, errors.New("shift not found")
	}

	if shift.Status == "closed" {
		return nil, errors.New("shift is already closed")
	}

	now := time.Now()
	shift.EndTime = &now
	shift.EndCash = &endCash
	shift.Status = "closed"

	if err := s.db.Save(&shift).Error; err != nil {
		return nil, err
	}

	return &shift, nil
}

func (s *ShiftService) GetActiveShift(businessID uint, userID uint) (*Shift, error) {
	var shift Shift
	err := s.db.Where("business_id = ? AND user_id = ? AND status = ?", businessID, userID, "open").First(&shift).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shift, nil
}

func (s *ShiftService) ListByBusiness(businessID uint) ([]Shift, error) {
	var shifts []Shift
	err := s.db.Where("business_id = ?", businessID).Order("created_at desc").Limit(50).Find(&shifts).Error
	return shifts, err
}

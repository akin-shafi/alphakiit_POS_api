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

// ShiftSummary contains detailed information about a shift
type ShiftSummary struct {
	Shift            Shift   `json:"shift"`
	TotalSales       float64 `json:"total_sales"`
	TransactionCount int     `json:"transaction_count"`
	ExpectedCash     float64 `json:"expected_cash"`
	ActualCash       float64 `json:"actual_cash"`
	Variance         float64 `json:"variance"`
}

// GetShiftSummary returns detailed summary of a shift including sales data
func (s *ShiftService) GetShiftSummary(shiftID uint) (*ShiftSummary, error) {
	var shift Shift
	if err := s.db.First(&shift, shiftID).Error; err != nil {
		return nil, errors.New("shift not found")
	}

	summary := &ShiftSummary{
		Shift:            shift,
		TotalSales:       shift.TotalSales,
		TransactionCount: shift.TransactionCount,
		ExpectedCash:     shift.StartCash + shift.TotalSales,
	}

	if shift.EndCash != nil {
		summary.ActualCash = *shift.EndCash
		summary.Variance = summary.ActualCash - summary.ExpectedCash
	}

	return summary, nil
}

// UpdateShiftMetrics updates the total sales and transaction count for a shift
// This should be called when a sale is completed
func (s *ShiftService) UpdateShiftMetrics(shiftID uint, saleAmount float64) error {
	return s.db.Model(&Shift{}).
		Where("id = ?", shiftID).
		Updates(map[string]interface{}{
			"total_sales":       gorm.Expr("total_sales + ?", saleAmount),
			"transaction_count": gorm.Expr("transaction_count + 1"),
		}).Error
}

// ValidateActiveShift checks if a user has an active shift for the business
// Returns the shift ID if active, error if not
func (s *ShiftService) ValidateActiveShift(businessID, userID uint) (*Shift, error) {
	shift, err := s.GetActiveShift(businessID, userID)
	if err != nil {
		return nil, err
	}

	if shift == nil {
		return nil, errors.New("no active shift found - please start your shift first")
	}

	return shift, nil
}

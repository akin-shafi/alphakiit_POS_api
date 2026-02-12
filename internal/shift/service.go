package shift

import (
	"errors"
	"pos-fiber-app/internal/notification"
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
	// Auto-close any old shifts before starting/checking
	s.AutoCloseOldShifts(businessID)

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

func (s *ShiftService) EndShift(shiftID uint, endCash float64, closedByName string) (*Shift, error) {
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
	shift.ClosedByName = closedByName
	shift.Status = "closed"

	// Calculate variance
	shift.ExpectedCash = shift.StartCash + shift.TotalCashSales
	shift.CashVariance = endCash - shift.ExpectedCash

	if err := s.db.Save(&shift).Error; err != nil {
		return nil, err
	}

	// Real-time Alert for Owner on shift closing
	go func() {
		notifier := notification.GetDefaultService(s.db)
		// Get business for currency
		var businessObj struct {
			Currency string
		}
		s.db.Table("businesses").Select("currency").Where("id = ?", shift.BusinessID).Scan(&businessObj)

		if shift.CashVariance != 0 {
			notifier.SendShiftVarianceAlert(
				shift.BusinessID,
				shift.ID,
				shift.UserName,
				shift.ExpectedCash,
				*shift.EndCash,
				shift.CashVariance,
				businessObj.Currency,
			)
		} else {
			// Regular shift closed report
			notifier.SendShiftClosedReport(
				shift.BusinessID,
				shift.ID,
				shift.UserName,
				shift.TotalSales,
				shift.TransactionCount,
				businessObj.Currency,
			)
		}
	}()

	return &shift, nil
}

func (s *ShiftService) GetActiveShift(businessID uint, userID uint) (*Shift, error) {
	// Auto-close any old shifts
	s.AutoCloseOldShifts(businessID)

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
		ExpectedCash:     shift.ExpectedCash,
		ActualCash:       0,
		Variance:         shift.CashVariance,
	}

	if shift.EndCash != nil {
		summary.ActualCash = *shift.EndCash
	}

	return summary, nil
}

func (s *ShiftService) UpdateShiftMetrics(shiftID uint, saleAmount float64, paymentMethod string) error {
	updates := map[string]interface{}{
		"total_sales":       gorm.Expr("total_sales + ?", saleAmount),
		"transaction_count": gorm.Expr("transaction_count + 1"),
	}

	switch paymentMethod {
	case "CASH":
		updates["total_cash_sales"] = gorm.Expr("total_cash_sales + ?", saleAmount)
	case "CARD":
		updates["total_card_sales"] = gorm.Expr("total_card_sales + ?", saleAmount)
	case "TRANSFER":
		updates["total_transfer_sales"] = gorm.Expr("total_transfer_sales + ?", saleAmount)
	case "EXTERNAL_TERMINAL":
		updates["total_external_terminal_sales"] = gorm.Expr("total_external_terminal_sales + ?", saleAmount)
	case "CREDIT":
		updates["total_credit_sales"] = gorm.Expr("total_credit_sales + ?", saleAmount)
	}

	return s.db.Model(&Shift{}).Where("id = ?", shiftID).Updates(updates).Error
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

// AutoCloseOldShifts closes any open shifts that started on previous days
func (s *ShiftService) AutoCloseOldShifts(businessID uint) {
	now := time.Now()
	// Get start of today in local time
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var oldShifts []Shift
	s.db.Where("business_id = ? AND status = ? AND start_time < ?", businessID, "open", today).Find(&oldShifts)

	for _, shift := range oldShifts {
		endTime := today.Add(-1 * time.Second) // End it at the very last second of yesterday
		endCash := shift.StartCash + shift.TotalSales

		s.db.Model(&shift).Updates(map[string]interface{}{
			"status":         "closed",
			"end_time":       endTime,
			"end_cash":       endCash,
			"expected_cash":  endCash,
			"cash_variance":  0,
			"closed_by_name": "System (Auto-closed)",
			"notes":          "Automatically closed by system at end of day",
		})
	}
}

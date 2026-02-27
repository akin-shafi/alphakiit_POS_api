package expense

import (
	"time"

	"gorm.io/gorm"
)

func Create(db *gorm.DB, businessID uint, req CreateExpenseRequest) (*Expense, error) {
	exp := &Expense{
		BusinessID:  businessID,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}

	if err := db.Create(exp).Error; err != nil {
		return nil, err
	}
	return exp, nil
}

func List(db *gorm.DB, businessID uint, from, to string) ([]Expense, error) {
	expenses := []Expense{}
	query := db.Where("business_id = ?", businessID)

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("date >= ?", t)
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("date <= ?", t)
		}
	}

	err := query.Order("date DESC").Find(&expenses).Error
	return expenses, err
}

func Delete(db *gorm.DB, id, businessID uint) error {
	return db.Where("id = ? AND business_id = ?", id, businessID).Delete(&Expense{}).Error
}

func GetSummary(db *gorm.DB, businessID uint, from, to time.Time) (float64, error) {
	var total float64
	err := db.Model(&Expense{}).
		Where("business_id = ? AND date >= ? AND date <= ?", businessID, from, to).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&total)
	return total, err
}

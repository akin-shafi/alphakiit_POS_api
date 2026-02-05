// internal/sale/activity_log.go
package sale

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ActionType string

const (
	ActionCreated     ActionType = "created"
	ActionUpdated     ActionType = "updated"
	ActionCompleted   ActionType = "completed"
	ActionVoided      ActionType = "voided"
	ActionTransferred ActionType = "transferred"
	ActionMerged      ActionType = "merged"
	ActionResumed     ActionType = "resumed"
	ActionItemAdded   ActionType = "item_added"
	ActionItemRemoved ActionType = "item_removed"
)

// SaleActivityLog tracks all actions performed on a sale for audit purposes
type SaleActivityLog struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	SaleID      uint       `gorm:"index" json:"sale_id"`
	BusinessID  uint       `gorm:"index" json:"business_id"`
	ActionType  ActionType `gorm:"type:varchar(50)" json:"action_type"`
	PerformedBy uint       `json:"performed_by"`             // user_id of person who performed action
	Details     string     `gorm:"type:text" json:"details"` // JSON with additional details
	CreatedAt   time.Time  `json:"created_at"`
}

// ActivityDetails represents the structure of the Details JSON field
type ActivityDetails struct {
	FromTable     string      `json:"from_table,omitempty"`
	ToTable       string      `json:"to_table,omitempty"`
	MergedFrom    []uint      `json:"merged_from,omitempty"`
	ProductID     uint        `json:"product_id,omitempty"`
	ProductName   string      `json:"product_name,omitempty"`
	Quantity      int         `json:"quantity,omitempty"`
	Reason        string      `json:"reason,omitempty"`
	OldValue      interface{} `json:"old_value,omitempty"`
	NewValue      interface{} `json:"new_value,omitempty"`
	AmountPaid    float64     `json:"amount_paid,omitempty"`
	PaymentMethod string      `json:"payment_method,omitempty"`
}

// LogActivity creates an activity log entry
func LogActivity(db *gorm.DB, saleID, businessID, userID uint, actionType ActionType, details ActivityDetails) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	log := &SaleActivityLog{
		SaleID:      saleID,
		BusinessID:  businessID,
		ActionType:  actionType,
		PerformedBy: userID,
		Details:     string(detailsJSON),
	}

	return db.Create(log).Error
}

// GetSaleHistory returns all activity logs for a sale
func GetSaleHistory(db *gorm.DB, saleID uint) ([]SaleActivityLog, error) {
	var logs []SaleActivityLog
	err := db.Where("sale_id = ?", saleID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetSaleHistoryWithUser returns activity logs with user information
type SaleActivityLogWithUser struct {
	SaleActivityLog
	UserName string `json:"user_name"`
}

func GetSaleHistoryWithUser(db *gorm.DB, saleID uint) ([]SaleActivityLogWithUser, error) {
	var logs []SaleActivityLogWithUser
	err := db.Table("sale_activity_logs").
		Select("sale_activity_logs.*, users.first_name || ' ' || users.last_name as user_name").
		Joins("LEFT JOIN users ON users.id = sale_activity_logs.performed_by").
		Where("sale_activity_logs.sale_id = ?", saleID).
		Order("sale_activity_logs.created_at DESC").
		Scan(&logs).Error

	return logs, err
}

// GetRecentActivityByBusiness returns recent activity logs for a business
func GetRecentActivityByBusiness(db *gorm.DB, businessID uint, limit int) ([]SaleActivityLogWithUser, error) {
	if limit <= 0 {
		limit = 50
	}

	var logs []SaleActivityLogWithUser
	err := db.Table("sale_activity_logs").
		Select("sale_activity_logs.*, users.first_name || ' ' || users.last_name as user_name").
		Joins("LEFT JOIN users ON users.id = sale_activity_logs.performed_by").
		Where("sale_activity_logs.business_id = ?", businessID).
		Order("sale_activity_logs.created_at DESC").
		Limit(limit).
		Scan(&logs).Error

	return logs, err
}

// MigrateActivityLog runs the database migration for activity logs
func MigrateActivityLog(db *gorm.DB) error {
	return db.AutoMigrate(&SaleActivityLog{})
}

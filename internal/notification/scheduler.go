package notification

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func StartNotificationScheduler(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	service := GetDefaultService(db)

	go func() {
		for range ticker.C {
			service.RunScheduledReports()
		}
	}()
}

func (n *NotificationService) RunScheduledReports() {
	// 1. Find all businesses that have reporting enabled
	var businesses []struct {
		ID               uint
		DailyReportTime  string
		LastReportSentAt *time.Time
	}
	n.db.Table("businesses").
		Where("reporting_enabled = ?", true).
		Select("id, daily_report_time, last_report_sent_at").
		Find(&businesses)

	now := time.Now()
	currentTime := now.Format("15:04") // HH:MM
	today := now.Format("2006-01-02")

	for _, b := range businesses {
		// Check if it's the right time
		if b.DailyReportTime != currentTime {
			continue
		}

		// Check if we already sent a report today
		if b.LastReportSentAt != nil && b.LastReportSentAt.Format("2006-01-02") == today {
			continue
		}

		fmt.Printf("Scheduler: Preparing daily report for business %d\n", b.ID)

		// Send report
		err := n.GenerateAndSendDailyReport(b.ID)
		if err == nil {
			// Update last sent time
			n.db.Table("businesses").Where("id = ?", b.ID).Update("last_report_sent_at", now)
		} else {
			fmt.Printf("Scheduler Error: Failed to generate report for business %d: %v\n", b.ID, err)
		}
	}
}

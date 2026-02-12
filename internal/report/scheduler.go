package report

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func StartReportScheduler(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	service := NewReportService(db)

	go func() {
		for range ticker.C {
			RunScheduledReports(db, service)
		}
	}()
}

func RunScheduledReports(db *gorm.DB, s *ReportService) {
	var businesses []struct {
		ID               uint
		DailyReportTime  string
		LastReportSentAt *time.Time
	}
	db.Table("businesses").
		Where("reporting_enabled = ?", true).
		Select("id, daily_report_time, last_report_sent_at").
		Find(&businesses)

	now := time.Now()
	currentTime := now.Format("15:04")
	today := now.Format("2006-01-02")

	for _, b := range businesses {
		if b.DailyReportTime != currentTime {
			continue
		}
		if b.LastReportSentAt != nil && b.LastReportSentAt.Format("2006-01-02") == today {
			continue
		}

		fmt.Printf("Report Scheduler: Generating report for business %d\n", b.ID)
		err := s.GenerateAndSendDailyReport(b.ID)
		if err == nil {
			db.Table("businesses").Where("id = ?", b.ID).Update("last_report_sent_at", now)
		} else {
			fmt.Printf("Report Scheduler Error: %v\n", err)
		}
	}
}

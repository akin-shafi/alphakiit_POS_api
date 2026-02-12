package archiver

import (
	"log"
	"time"

	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/sale"

	"gorm.io/gorm"
)

// StartDataLifecycleManager runs a background ticker to handle archiving and cleanup
func StartDataLifecycleManager(db *gorm.DB) {
	log.Println("[ARCHIVER] Starting Data Lifecycle Manager...")

	// Check every hour (or we could do daily)
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for range ticker.C {
			handleAllBusinesses(db)
		}
	}()
}

func handleAllBusinesses(db *gorm.DB) {
	var businesses []business.Business
	// Find businesses with archiving enabled or retention policies
	if err := db.Find(&businesses).Error; err != nil {
		return
	}

	for _, biz := range businesses {
		processBusiness(db, &biz)
	}
}

func processBusiness(db *gorm.DB, biz *business.Business) {
	// 1. Check if archiving is due (if enabled)
	if biz.AutoArchiveEnabled {
		if isArchiveDue(biz) {
			archiveBusinessData(db, biz)
		}
	}

	// 2. Perform cleanup of old records based on retention policy
	if biz.DataRetentionMonths > 0 {
		sale.PerformCleanup(db, biz.ID, biz.DataRetentionMonths, biz.Name)
	}
}

func isArchiveDue(biz *business.Business) bool {
	if biz.LastArchivedAt == nil {
		return true // Never archived
	}

	days := 30
	if biz.ArchiveFrequency == "bi-monthly" {
		days = 60
	} else if biz.ArchiveFrequency == "quarterly" {
		days = 90
	}

	nextDate := biz.LastArchivedAt.AddDate(0, 0, days)
	return time.Now().After(nextDate)
}

func archiveBusinessData(db *gorm.DB, biz *business.Business) {
	log.Printf("[ARCHIVER] Archiving data for business: %s (ID: %d)\n", biz.Name, biz.ID)

	// 1. Determine the period to backup (e.g. last month)
	to := time.Now()
	from := to.AddDate(0, -1, 0) // Default to last month

	// 2. If Google Drive is linked, upload the CSV
	if biz.GoogleDriveLinked {
		if err := UploadBackupToDrive(db, biz, from, to); err != nil {
			log.Printf("[ARCHIVER ERROR] Failed to upload to Google Drive for %s: %v\n", biz.Name, err)
			return // Don't update LastArchivedAt if upload failed
		}
	}

	// 3. Update timestamp to indicate success (even if not linked, it "processed" the manual period)
	now := time.Now()
	db.Model(biz).Update("last_archived_at", &now)

	log.Printf("[ARCHIVER] Archive completed for %s\n", biz.Name)
}

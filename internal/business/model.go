// internal/business/model.go
package business

import (
	"time"

	"pos-fiber-app/internal/common" // Import shared types

	"gorm.io/gorm"
)

// Business model
type Business struct {
	ID                 uint                `gorm:"primaryKey" json:"id"`
	TenantID           string              `gorm:"index;size:8" json:"tenant_id"`
	Name               string              `gorm:"size:255;not null" json:"name"`
	Type               common.BusinessType `gorm:"type:varchar(50);not null" json:"type"`
	Address            string              `json:"address,omitempty"`
	City               string              `json:"city,omitempty"`
	Currency           common.Currency     `gorm:"type:varchar(3);default:'NGN';not null" json:"currency"`
	IsSeeded           bool                `gorm:"default:false" json:"is_seeded"`
	SubscriptionStatus string              `gorm:"type:varchar(20);default:'TRIAL'" json:"subscription_status"`
	SubscriptionExpiry *time.Time          `json:"subscription_expiry,omitempty"`
	InstallerID        *uint               `gorm:"index" json:"installer_id,omitempty"`
	// Data Management Settings
	DataRetentionMonths int            `gorm:"default:6" json:"data_retention_months"`
	AutoArchiveEnabled  bool           `gorm:"default:false" json:"auto_archive_enabled"`
	ArchiveFrequency    string         `gorm:"type:varchar(20);default:'monthly'" json:"archive_frequency"`
	GoogleDriveLinked   bool           `gorm:"default:false" json:"google_drive_linked"`
	GoogleAccessToken   string         `json:"-"`
	GoogleRefreshToken  string         `json:"-"`
	GoogleTokenExpiry   *time.Time     `json:"-"`
	GoogleDriveFolderID string         `json:"-"`
	LastArchivedAt      *time.Time     `json:"last_archived_at,omitempty"`
	WhatsAppEnabled     bool           `gorm:"default:false" json:"whatsapp_enabled"`
	WhatsAppNumber      string         `gorm:"type:varchar(20)" json:"whatsapp_number"`
	ReportingEnabled    bool           `gorm:"default:true" json:"reporting_enabled"`
	DailyReportTime     string         `gorm:"type:varchar(5);default:'22:00'" json:"daily_report_time"` // HH:MM in 24h format
	LastReportSentAt    *time.Time     `json:"last_report_sent_at,omitempty"`
	ActiveModules       []string       `gorm:"-" json:"active_modules,omitempty"` // populated on fetch
	CreatedAt           time.Time      `json:"-"`
	UpdatedAt           time.Time      `json:"-"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// Tenant model (keep if still used elsewhere)
type Tenant struct {
	ID        string    `gorm:"primaryKey;size:8" json:"id"`
	Name      string    `gorm:"size:255" json:"name"`
	OwnerID   uint      `gorm:"not null" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

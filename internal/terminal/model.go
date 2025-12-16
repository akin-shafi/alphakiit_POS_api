package terminal

import "time"

type Terminal struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  string `gorm:"index"`
	OutletID  uint   `json:"outlet_id"`
	Code      string `gorm:"uniqueIndex" json:"code"`
	Active    bool   `json:"active"`
	CreatedAt time.Time
}

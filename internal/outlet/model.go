package outlet

import "time"

type Outlet struct {
	ID        uint      `gorm:"primaryKey"`
	TenantID  string    `gorm:"index" json:"tenant_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package advert

import (
	"time"

	"gorm.io/gorm"
)

type AdvertType string

const (
	AdvertTypeImage AdvertType = "IMAGE"
	AdvertTypeVideo AdvertType = "VIDEO"
)

type Advert struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	BusinessID *uint          `gorm:"index" json:"business_id,omitempty"` // Null for global adverts
	Title      string         `gorm:"size:200" json:"title"`
	Type       AdvertType     `gorm:"size:20;not null" json:"type"` // IMAGE or VIDEO
	URL        string         `gorm:"type:text;not null" json:"url"`
	Active     bool           `gorm:"default:true" json:"active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

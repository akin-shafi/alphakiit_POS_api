// internal/business/dto.go
package business

import (
	"pos-fiber-app/internal/common"
)

type BusinessType = common.BusinessType
type Currency = common.Currency

type CreateBusinessRequest struct {
	Name     string       `json:"name" validate:"required,min=2,max=255"`
	Type     BusinessType `json:"type" validate:"required,oneof=RESTAURANT BAR SUPERMARKET LOUNGE FUEL_STATION RETAIL HOTEL PHARMACY CLINIC BOUTIQUE OTHER"`
	Address  string       `json:"address,omitempty"`
	City     string       `json:"city,omitempty"`
	Currency Currency     `json:"currency" validate:"required,oneof=NGN USD GBP EUR"`
}

type UpdateBusinessRequest struct {
	Name     string       `json:"name,omitempty"`
	Type     BusinessType `json:"type,omitempty" validate:"omitempty,oneof=RESTAURANT BAR SUPERMARKET LOUNGE FUEL_STATION RETAIL HOTEL PHARMACY CLINIC BOUTIQUE OTHER"`
	Address  string       `json:"address,omitempty"`
	City     string       `json:"city,omitempty"`
	Currency *Currency    `json:"currency,omitempty" validate:"omitempty,oneof=NGN USD GBP EUR"`
	// Data Management
	DataRetentionMonths *int   `json:"data_retention_months,omitempty"`
	AutoArchiveEnabled  *bool  `json:"auto_archive_enabled,omitempty"`
	ArchiveFrequency    string `json:"archive_frequency,omitempty"`
	WhatsAppEnabled        *bool  `json:"whatsapp_enabled,omitempty"`
	WhatsAppNumber         string `json:"whatsapp_number,omitempty"`
	TableManagementEnabled *bool  `json:"table_management_enabled,omitempty"`
	SaveToDraftEnabled     *bool  `json:"save_to_draft_enabled,omitempty"`
}

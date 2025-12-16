// pkg/database/tenant_scope.go
package database

import "gorm.io/gorm"

func WithTenant(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

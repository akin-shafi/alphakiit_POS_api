package types

type UserClaims struct {
	UserID   uint
	TenantID string
	Role     string
	OutletID *uint
	
}

package types

type UserClaims struct {
	UserID   uint
	UserName string
	TenantID string
	Role     string
	OutletID *uint
}

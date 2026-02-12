package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// type Role string

// Generate JWT token helper
func generateToken(secret string, base Claims, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   base.UserID,
		UserName: base.UserName,
		TenantID: base.TenantID,
		Role:     base.Role,
		OutletID: base.OutletID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   "auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateAccessToken(claims Claims) (string, error) {
	ttl := 15 * time.Minute
	if claims.Role == "KITCHEN" || claims.Role == "CHEF" || claims.Role == "BARTENDER" {
		ttl = 24 * time.Hour // Long session for kitchen displays
	}
	return generateToken(os.Getenv("JWT_SECRET"), claims, ttl)
}

func GenerateRefreshToken(claims Claims) (string, error) {
	return generateToken(os.Getenv("JWT_REFRESH_SECRET"), claims, 7*24*time.Hour)
}

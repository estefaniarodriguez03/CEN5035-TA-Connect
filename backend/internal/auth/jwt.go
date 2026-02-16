package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload (user id, email, role).
type Claims struct {
	jwt.RegisteredClaims
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

const defaultExpiry = 24 * time.Hour

// NewToken creates a signed JWT for the user. Uses JWT_SECRET from env.
func NewToken(userID int, email, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: userID,
		Email:  email,
		Role:   role,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

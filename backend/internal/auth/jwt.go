package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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

// ParseToken validates the token string and returns the claims. Returns error if invalid or expired.
func ParseToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET not set")
	}
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GetClaimsFromRequest returns the JWT claims from the Authorization: Bearer <token> header.
// Returns nil, nil if no header or invalid token; caller should respond with 401.
func GetClaimsFromRequest(r *http.Request) (*Claims, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, nil
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return nil, nil
	}
	tokenString := strings.TrimSpace(auth[len(prefix):])
	if tokenString == "" {
		return nil, nil
	}
	return ParseToken(tokenString)
}

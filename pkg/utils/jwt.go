package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JWTUtil struct {
	SecretKey string
}

func NewJWTUtil(secretKey string) *JWTUtil {
	return &JWTUtil{SecretKey: secretKey}
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	TenantID uuid.UUID `json:"tenant_id"`
	Type     string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateToken creates a short-lived access token (15 minutes)
func (j *JWTUtil) GenerateToken(userID uuid.UUID, email, role string, tenantID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		TenantID: tenantID,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

// GenerateRefreshToken creates a long-lived refresh token (7 days)
func (j *JWTUtil) GenerateRefreshToken(userID uuid.UUID, tenantID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWTUtil) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Default to "access" if type not present (backward compatibility)
	if claims.Type == "" {
		claims.Type = "access"
	}

	if claims.Type != "access" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTUtil) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// GetUserIDFromToken extracts user ID from token string
func (j *JWTUtil) GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	// Use ParseWithClaims directly to avoid the "access" type check in ValidateToken
	// This is useful if we just want to know WHO the user is, regardless of token type
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}


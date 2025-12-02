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
	jwt.RegisteredClaims
}

func (j *JWTUtil) GenerateToken(userID uuid.UUID, email, role string, tenantID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		TenantID: tenantID,
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

	return claims, nil
}

// Tambahkan function ini jika test masih butuh
func (j *JWTUtil) GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

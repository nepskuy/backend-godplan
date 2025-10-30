package utils

import (
	"testing"
)

func TestJWTUtil(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret-key")

	// Test GenerateToken
	token, err := jwtUtil.GenerateToken(1, "test@example.com", "employee")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("Generated token is empty")
	}

	// Test ValidateToken
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Expected Email test@example.com, got %s", claims.Email)
	}
	if claims.Role != "employee" {
		t.Errorf("Expected Role employee, got %s", claims.Role)
	}
}

func TestJWTUtil_InvalidToken(t *testing.T) {
	jwtUtil := NewJWTUtil("test-secret-key")

	// Test with invalid token
	_, err := jwtUtil.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, but got none")
	}

	// Test with different secret key
	otherJWTUtil := NewJWTUtil("different-secret-key")
	token, _ := jwtUtil.GenerateToken(1, "test@example.com", "employee")
	_, err = otherJWTUtil.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for token with different secret, but got none")
	}
}

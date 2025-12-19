package utils

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength minimum password length
	MinPasswordLength = 12
	// BcryptCost cost for bcrypt hashing (14 is very secure)
	BcryptCost = 14
)

// HashPassword hashes a password using bcrypt with high cost
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength validates password complexity
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("password must be at least 12 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password123!", "admin123456!", "welcome12345!",
		"qwerty123456!", "letmein12345!", "password1234!",
	}
	
	passwordLower := regexp.MustCompile(`\s+`).ReplaceAllString(password, "")
	for _, weak := range weakPasswords {
		if passwordLower == weak {
			return errors.New("password is too common, please choose a stronger password")
		}
	}

	return nil
}

// SanitizeForLog removes sensitive data from logs
func SanitizeForLog(data map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{
		"password", "token", "secret", "api_key", "apikey",
		"credit_card", "creditcard", "cvv", "ssn",
		"authorization", "auth", "bearer",
	}

	sanitized := make(map[string]interface{})
	for k, v := range data {
		isSensitive := false
		for _, field := range sensitiveFields {
			if regexp.MustCompile(`(?i)`+field).MatchString(k) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

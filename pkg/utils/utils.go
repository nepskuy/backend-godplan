package utils

import "time"

// GetCurrentTimestamp returns current timestamp in RFC3339 format
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// Atau jika Anda ingin format yang lebih sederhana:
func GetCurrentTimestampSimple() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

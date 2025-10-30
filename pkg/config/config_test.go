package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "godplan_test")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("JWT_SECRET", "test-secret-key")

	cfg := Load()

	if cfg.DBHost != "localhost" {
		t.Errorf("Expected DB_HOST 'localhost', got '%s'", cfg.DBHost)
	}
	if cfg.JWTSecret != "test-secret-key" {
		t.Errorf("Expected JWT_SECRET 'test-secret-key', got '%s'", cfg.JWTSecret)
	}
}

func TestGetDBConnectionString(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "godplan_test",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=godplan_test sslmode=disable"
	actual := cfg.GetDBConnectionString()

	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}

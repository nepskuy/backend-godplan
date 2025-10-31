package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	DBSSLRootCert string
	ServerPort    string
	JWTSecret     string
	DatabaseURL   string
}

// Load memuat konfigurasi dari environment variables
func Load() *Config {
	log.Println("🔧 Loading configuration...")

	cfg := &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "godplan"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		DBSSLRootCert: getEnv("DB_SSLROOTCERT", ""),
		ServerPort:    getEnv("PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
		DatabaseURL:   os.Getenv("DATABASE_URL"), // gunakan DATABASE_URL jika ada
	}

	log.Println("✅ Configuration loaded successfully")
	return cfg
}

// GetDBConnectionString mengembalikan connection string untuk database
func (c *Config) GetDBConnectionString() string {
	// PRIORITAS 1: Gunakan DATABASE_URL jika ada (production)
	if c.DatabaseURL != "" {
		log.Println("✅ Using DATABASE_URL from environment")

		// Tambahkan search_path jika belum ada
		if !strings.Contains(c.DatabaseURL, "search_path") && !strings.Contains(c.DatabaseURL, "options") {
			separator := "?"
			if strings.Contains(c.DatabaseURL, "?") {
				separator = "&"
			}
			return c.DatabaseURL + separator + "search_path=godplan,public"
		}

		return c.DatabaseURL
	}

	// PRIORITAS 2: Bangun dari individual env vars
	log.Println("⚠️ DATABASE_URL not found, building from individual variables")
	log.Printf("   DB_HOST: %s", c.DBHost)
	log.Printf("   DB_PORT: %s", c.DBPort)
	log.Printf("   DB_USER: %s", c.DBUser)
	log.Printf("   DB_NAME: %s", c.DBName)
	log.Printf("   DB_SSLMODE: %s", c.DBSSLMode)

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)

	// Tambahkan sslrootcert jika ada
	if c.DBSSLRootCert != "" {
		connStr += fmt.Sprintf(" sslrootcert=%s", c.DBSSLRootCert)
		log.Println("📝 Added sslrootcert to connection string")
	}

	// Tambahkan search_path untuk development
	if !strings.Contains(connStr, "search_path") {
		connStr += " search_path=godplan,public"
		log.Println("📝 Added search_path to connection string")
	}

	return connStr
}

// getEnv mendapatkan environment variable dengan default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue != "" && key != "DB_PASSWORD" && key != "JWT_SECRET" {
			log.Printf("⚙️ Using default for %s: %s", key, defaultValue)
		}
		return defaultValue
	}

	// Jangan log password/secret
	if key == "DB_PASSWORD" || key == "JWT_SECRET" {
		log.Printf("✅ %s is set (hidden)", key)
	} else {
		log.Printf("✅ %s: %s", key, value)
	}

	return value
}

// IsProduction mengecek apakah running di production
func IsProduction() bool {
	return os.Getenv("VERCEL") != "" || os.Getenv("ENVIRONMENT") == "production"
}

// IsDevelopment mengecek apakah running di development
func IsDevelopment() bool {
	return !IsProduction()
}

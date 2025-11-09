package config

import (
	"fmt"
	"io/ioutil"
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
	Env           string
}

// Variabel cache untuk environment
var (
	isProduction  *bool
	isDevelopment *bool
)

// Load memuat konfigurasi dari environment variables
func Load() *Config {
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
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Env:           getEnv("ENV", "development"),
	}

	// Hanya log di development
	if IsDevelopment() {
		logConfig(cfg)
	}

	return cfg
}

// logConfig hanya dijalankan di development
func logConfig(cfg *Config) {
	fmt.Println("🔧 Loading configuration...")

	if cfg.DatabaseURL != "" {
		fmt.Println("✅ DATABASE_URL is available")
		maskedURL := maskPassword(cfg.DatabaseURL)
		fmt.Printf("📝 Using DATABASE_URL: %s\n", maskedURL)
	} else {
		fmt.Println("⚠️ DATABASE_URL not found, using individual DB config")
		fmt.Printf("📝 DB Host: %s, Port: %s, User: %s, Name: %s\n",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	}

	fmt.Println("✅ Configuration loaded successfully")
}

// writeCACert menulis DB_CA_PEM ke file sementara jika ada
func writeCACert() (string, error) {
	pemContent := os.Getenv("DB_CA_PEM")
	if pemContent == "" {
		return "", fmt.Errorf("DB_CA_PEM environment variable is empty")
	}

	tmpFile, err := ioutil.TempFile("", "ca-*.pem")
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.Write([]byte(pemContent)); err != nil {
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

// GetDBConnectionString mengembalikan connection string untuk database
func (c *Config) GetDBConnectionString() string {
	// PRIORITAS 1: Gunakan DATABASE_URL jika ada
	if c.DatabaseURL != "" {
		if IsDevelopment() {
			fmt.Println("✅ Using DATABASE_URL from environment")
		}

		if !strings.Contains(c.DatabaseURL, "search_path") && !strings.Contains(c.DatabaseURL, "options") {
			sep := "?"
			if strings.Contains(c.DatabaseURL, "?") {
				sep = "&"
			}
			return c.DatabaseURL + sep + "search_path=godplan,public"
		}

		return c.DatabaseURL
	}

	// PRIORITAS 2: Build dari individual env vars
	if IsDevelopment() {
		fmt.Println("⚠️ DATABASE_URL not found, building from individual variables")
		fmt.Printf("   DB_HOST: %s\n", c.DBHost)
		fmt.Printf("   DB_PORT: %s\n", c.DBPort)
		fmt.Printf("   DB_USER: %s\n", c.DBUser)
		fmt.Printf("   DB_NAME: %s\n", c.DBName)
		fmt.Printf("   DB_SSLMODE: %s\n", c.DBSSLMode)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)

	// Tambahkan sslrootcert jika DB_CA_PEM ada
	if os.Getenv("DB_CA_PEM") != "" {
		pemPath, err := writeCACert()
		if err != nil {
			// Di production, fail silently atau return error yang appropriate
			if IsDevelopment() {
				fmt.Printf("❌ Failed to write CA cert: %v\n", err)
			}
			return connStr
		}
		connStr += fmt.Sprintf(" sslrootcert=%s", pemPath)
		if IsDevelopment() {
			fmt.Printf("📝 Added SSL root certificate from DB_CA_PEM: %s\n", pemPath)
		}
	} else if c.DBSSLRootCert != "" {
		connStr += fmt.Sprintf(" sslrootcert=%s", c.DBSSLRootCert)
		if IsDevelopment() {
			fmt.Printf("📝 Added SSL root certificate from DB_SSLROOTCERT: %s\n", c.DBSSLRootCert)
		}
	}

	// Tambahkan search_path
	if !strings.Contains(connStr, "search_path") {
		connStr += " search_path=godplan,public"
		if IsDevelopment() {
			fmt.Println("📝 Added search_path to connection string")
		}
	}

	return connStr
}

// getEnv mendapatkan environment variable dengan default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		// Hanya log di development untuk non-sensitive values
		// PERBAIKAN: Hindari panggilan IsDevelopment() di sini untuk mencegah rekursi
		if key != "DB_PASSWORD" && key != "JWT_SECRET" {
			// Gunakan os.Getenv langsung untuk cek environment
			env := os.Getenv("ENV")
			if env == "" || env == "development" {
				fmt.Printf("⚙️ Using default for %s: %s\n", key, defaultValue)
			}
		}
		return defaultValue
	}

	// Hanya log di development
	// PERBAIKAN: Hindari panggilan IsDevelopment() di sini
	if key != "DB_PASSWORD" && key != "JWT_SECRET" {
		env := os.Getenv("ENV")
		if env == "" || env == "development" {
			fmt.Printf("✅ %s: %s\n", key, value)
		}
	} else {
		env := os.Getenv("ENV")
		if env == "" || env == "development" {
			fmt.Printf("✅ %s is set (hidden)\n", key)
		}
	}

	return value
}

// IsProduction mengecek environment - PERBAIKAN: Hindari rekursi
func IsProduction() bool {
	if isProduction != nil {
		return *isProduction
	}

	// Gunakan os.Getenv langsung, jangan panggil getEnv()
	result := os.Getenv("VERCEL") != "" ||
		os.Getenv("ENVIRONMENT") == "production" ||
		os.Getenv("ENV") == "production"

	isProduction = &result
	return result
}

// IsDevelopment mengecek environment - PERBAIKAN: Hindari rekursi
func IsDevelopment() bool {
	if isDevelopment != nil {
		return *isDevelopment
	}

	result := !IsProduction()
	isDevelopment = &result
	return result
}

// maskPassword untuk menyembunyikan password di logs
func maskPassword(connStr string) string {
	// Mask password dalam connection string
	for _, prefix := range []string{"password=", "Password="} {
		if idx := findIndex(connStr, prefix); idx != -1 {
			end := findNextSeparator(connStr, idx+len(prefix))
			return connStr[:idx+len(prefix)] + "****" + connStr[end:]
		}
	}

	// Mask password dalam URL format (postgres://user:pass@host)
	if idx := findIndex(connStr, "://"); idx != -1 {
		if idx2 := findIndex(connStr[idx+3:], "@"); idx2 != -1 {
			start := idx + 3
			end := start + idx2
			if colonIdx := findIndex(connStr[start:end], ":"); colonIdx != -1 {
				return connStr[:start+colonIdx+1] + "****" + connStr[end:]
			}
		}
	}
	return connStr
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func findNextSeparator(s string, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '&' || s[i] == '?' || s[i] == '#' {
			return i
		}
	}
	return len(s)
}

package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nepskuy/be-godplan/pkg/config"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	connStr := cfg.GetDBConnectionString()

	if config.IsDevelopment() {
		fmt.Println("🔵 Connecting to database...")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to open database connection: %v\n", err)
		}
		return err
	}

	// Set connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// Retry ping
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}

		if config.IsDevelopment() {
			fmt.Printf("⚠️ Database connection attempt %d/%d failed: %v\n", i+1, maxRetries, err)
		}

		time.Sleep(2 * time.Second)
	}

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ All database connection attempts failed: %v\n", err)
		}
		return err
	}

	if config.IsDevelopment() {
		fmt.Println("✅ Database connected successfully")
		logConnectionInfo(cfg)
	}

	return nil
}

// logConnectionInfo hanya dijalankan di development
func logConnectionInfo(cfg *config.Config) {
	if cfg.DatabaseURL != "" {
		fmt.Println("📡 Using DATABASE_URL from environment (Vercel/Production)")
	} else {
		fmt.Printf("💻 Using individual DB config: user=%s host=%s port=%s dbname=%s sslmode=%s\n",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}

	// Warn if SSL cert file missing
	if cfg.DBSSLMode == "verify-ca" {
		if os.Getenv("DB_CA_PEM") == "" && cfg.DBSSLRootCert != "" {
			if _, err := os.Stat(cfg.DBSSLRootCert); os.IsNotExist(err) {
				fmt.Printf("⚠️ SSL root certificate not found at path: %s\n", cfg.DBSSLRootCert)
			} else {
				fmt.Printf("🔒 SSL root certificate found at path: %s\n", cfg.DBSSLRootCert)
			}
		}
	}
}

// HealthCheck memastikan koneksi database masih aktif
func HealthCheck() error {
	if DB == nil {
		return sql.ErrConnDone
	}
	return DB.Ping()
}

// GetDB mengembalikan instance database (untuk compatibility dengan GORM style)
func GetDB() *sql.DB {
	return DB
}

// CloseDB menutup koneksi database
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Stats mengembalikan statistik koneksi database
func Stats() sql.DBStats {
	if DB != nil {
		return DB.Stats()
	}
	return sql.DBStats{}
}

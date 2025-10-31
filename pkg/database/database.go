package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nepskuy/be-godplan/pkg/config"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.Load()
	connStr := cfg.GetDBConnectionString()

	log.Println("🔵 Connecting to database...")

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("❌ Failed to open database connection: %v", err)
		return err
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection with retry mechanism
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Printf("⚠️ Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Printf("❌ All database connection attempts failed: %v", err)
		return err
	}

	log.Println("✅ Database connected successfully")

	// Log info connection (without sensitive info)
	if cfg.DatabaseURL != "" {
		log.Println("📡 Using DATABASE_URL from environment (Vercel/Production)")
	} else {
		log.Printf("💻 Using individual DB config: user=%s host=%s port=%s dbname=%s sslmode=%s sslrootcert=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode, cfg.DBSSLRootCert)
	}

	// Warn if SSL cert file is missing when sslmode is verify-ca
	if cfg.DBSSLMode == "verify-ca" && cfg.DBSSLRootCert != "" {
		if _, err := os.Stat(cfg.DBSSLRootCert); os.IsNotExist(err) {
			log.Printf("⚠️ SSL root certificate not found at path: %s", cfg.DBSSLRootCert)
		} else {
			log.Printf("🔒 SSL root certificate found at path: %s", cfg.DBSSLRootCert)
		}
	}

	return nil
}

// HealthCheck memastikan koneksi database masih aktif
func HealthCheck() error {
	if DB == nil {
		return sql.ErrConnDone
	}
	return DB.Ping()
}

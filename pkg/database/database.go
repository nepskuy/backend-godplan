package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/nepskuy/be-godplan/pkg/config"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.Load()
	connStr := cfg.GetDBConnectionString()

	log.Printf("🔵 Connecting to database...")

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

	// Test connection dengan retry mechanism
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Printf("⚠️ Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Printf("❌ All database connection attempts failed: %v", err)
		return err
	}

	log.Println("✅ Database connected successfully")

	// Log info connection (tanpa password)
	if cfg.DatabaseURL != "" {
		log.Println("📡 Using DATABASE_URL (Vercel/Production)")
	} else {
		log.Printf("💻 Using individual DB config: %s@%s:%s/%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	}

	return nil
}

// HealthCheck untuk memastikan koneksi masih aktif
func HealthCheck() error {
	if DB == nil {
		return sql.ErrConnDone
	}
	return DB.Ping()
}

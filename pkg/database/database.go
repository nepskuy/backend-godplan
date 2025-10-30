package database

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nepskuy/be-godplan/pkg/config"
)

var DB *sql.DB

func InitDB() error {
	cfg := config.Load()
	connStr := cfg.GetDBConnectionString()

	// Tambahkan search_path hanya jika belum ada di connection string
	if !strings.Contains(connStr, "search_path=") {
		separator := "?"
		if strings.Contains(connStr, "?") {
			separator = "&"
		}
		connStr = connStr + separator + "search_path=godplan,public"
		log.Println("📝 Added search_path to connection string")
	}

	log.Printf("🔵 Connecting to database...")

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection dengan retry mechanism
	maxRetries := 3
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
		return err
	}

	log.Println("✅ Database connected successfully")
	return nil
}

// HealthCheck untuk memastikan koneksi masih aktif
func HealthCheck() error {
	if DB == nil {
		return sql.ErrConnDone
	}
	return DB.Ping()
}

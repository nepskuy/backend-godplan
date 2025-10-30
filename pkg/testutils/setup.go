package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB membuat koneksi database untuk testing menggunakan testcontainers
func SetupTestDB(t *testing.T) *sql.DB {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "godplan_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %s", err)
	}

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %s", err)
		}
	})

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get host: %s", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get port: %s", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=test password=test dbname=godplan_test sslmode=disable",
		host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %s", err)
	}

	// Tunggu koneksi stabil
	time.Sleep(2 * time.Second)

	// Test koneksi
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %s", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %s", err)
	}

	return db
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			full_name VARCHAR(100),
			role VARCHAR(20) DEFAULT 'employee',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			priority VARCHAR(10) DEFAULT 'medium',
			completed BOOLEAN DEFAULT FALSE,
			due_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS attendances (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			check_in TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			check_out TIMESTAMP,
			notes TEXT,
			location VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_attendances_user_date ON attendances(user_id, DATE(check_in))`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

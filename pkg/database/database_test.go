package database

import (
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Skipf("Skipping database test: %v", err)
	}
	defer DB.Close()

	// Test basic query
	var result int
	err = DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Database query failed: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}
}

func TestTableCreation(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Skipf("Skipping database test: %v", err)
	}
	defer DB.Close()

	// Check if users table exists
	var tableExists bool
	err = DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'users'
		)
	`).Scan(&tableExists)

	if err != nil {
		t.Errorf("Table check failed: %v", err)
	}

	if !tableExists {
		t.Error("Users table should exist")
	}
}

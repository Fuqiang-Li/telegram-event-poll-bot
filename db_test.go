package main

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	// Open the database
	db, err := sql.Open("sqlite", tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the activities table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create activities table: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

func TestGetDBVersion(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Test initial version
	version, err := GetDBVersion(db)
	if err != nil {
		t.Errorf("GetDBVersion failed: %v", err)
	}
	if version != 0 {
		t.Errorf("Expected initial version to be 0, got %d", version)
	}

	// Test setting and getting a specific version
	_, err = db.Exec("PRAGMA user_version = 5")
	if err != nil {
		t.Fatalf("Failed to set user version: %v", err)
	}

	version, err = GetDBVersion(db)
	if err != nil {
		t.Errorf("GetDBVersion failed: %v", err)
	}
	if version != 5 {
		t.Errorf("Expected version to be 5, got %d", version)
	}
}

func TestMigrateDB(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Test migration from version 0 to 1
	err := MigrateDB(db)
	if err != nil {
		t.Errorf("MigrateDB failed: %v", err)
	}

	// Verify the migration was successful
	version, err := GetDBVersion(db)
	if err != nil {
		t.Errorf("GetDBVersion failed: %v", err)
	}
	if version != target_db_version {
		t.Errorf("Expected version to be %d, got %d", target_db_version, version)
	}

	// Verify the column was added
	rows, err := db.Query("PRAGMA table_info(activities)")
	if err != nil {
		t.Fatalf("Failed to query table info: %v", err)
	}
	defer rows.Close()

	var columnName string
	var found bool
	for rows.Next() {
		var cid, notnull, pk int
		var dtype string
		var dflt_value sql.NullString
		err := rows.Scan(&cid, &columnName, &dtype, &notnull, &dflt_value, &pk)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		if columnName == "created_by" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Column 'created_by' was not found after migration")
	}
}

func TestMigrateDBNoMigrationNeeded(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Set version to target version
	_, err := db.Exec("PRAGMA user_version = " + strconv.Itoa(target_db_version))
	if err != nil {
		t.Fatalf("Failed to set user version: %v", err)
	}

	// Try to migrate again
	err = MigrateDB(db)
	if err != nil {
		t.Errorf("MigrateDB failed: %v", err)
	}

	// Verify version hasn't changed
	version, err := GetDBVersion(db)
	if err != nil {
		t.Errorf("GetDBVersion failed: %v", err)
	}
	if version != target_db_version {
		t.Errorf("Expected version to be %d, got %d", target_db_version, version)
	}
}

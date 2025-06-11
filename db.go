package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

const (
	target_db_version = 2
)

var (
	dbMigrationMap = map[int][]string{
		1: {
			`ALTER TABLE activities ADD COLUMN created_by TEXT DEFAULT 'S'`,
		},
		2: {
			`ALTER TABLE activities ADD COLUMN created_by_id INTEGER DEFAULT 0`,
			`ALTER TABLE events ADD COLUMN created_by_id INTEGER DEFAULT 0`,
			`ALTER TABLE event_users ADD COLUMN user_id INTEGER DEFAULT 0`,
		},
	}
)

func MigrateDB(db *sql.DB) error {
	setDbUserVersionQuery := "PRAGMA user_version = " + strconv.Itoa(target_db_version)

	// Check if the database is empty
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		return err
	}
	// If database is empty, just set the version and return
	// otherwise, check to migrate db
	if count == 0 {
		_, err = db.Exec(setDbUserVersionQuery)
		return err
	}

	userVersion, err := GetDBVersion(db)
	if err != nil {
		return err
	}
	log.Println("DB user version", userVersion, "Target version", target_db_version)
	if userVersion >= target_db_version {
		return nil
	}
	queries := []string{}
	for i := userVersion; i < target_db_version; i++ {
		queries = append(queries, dbMigrationMap[i+1]...)
	}
	queries = append(queries, setDbUserVersionQuery)
	log.Println("Migrating DB to version", target_db_version)
	log.Println("Executing queries:\n", strings.Join(queries, "\n"))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, q := range queries {
		_, err := tx.Exec(q)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func GetDBVersion(db *sql.DB) (int, error) {
	row := db.QueryRow("PRAGMA user_version")
	var userVersion int
	err := row.Scan(&userVersion)
	if err != nil {
		return 0, err
	}
	return userVersion, nil
}

package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

const (
	target_db_version = 1
)

var (
	dbMigrationMap = map[int][]string{
		1: {
			`ALTER TABLE activities ADD COLUMN created_by TEXT DEFAULT 'S'`,
		},
	}
)

func MigrateDB(db *sql.DB) error {
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
	queries = append(queries, "PRAGMA user_version = "+strconv.Itoa(target_db_version))
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

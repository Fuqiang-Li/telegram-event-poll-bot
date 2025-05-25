package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Org string

const (
	OrgCC   = "CC"
	OrgPEAK = "PEAK"
)

var AllOrgs = []Org{OrgCC, OrgPEAK}

type Activity struct {
	ID        int64
	Name      string
	Org       Org
	Lead      string
	CoLeads   []string
	CreatedBy string
	StartedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a Activity) string() string {
	return fmt.Sprintf("<b>%s %s - (Org: %s) - (ID:%d):</b> %s(L), %s(CoL)", a.StartedAt.Format(displayTimeFormat), a.Name, a.Org, a.ID, a.Lead, strings.Join(a.CoLeads, "(CoL), "))
}

// ActivityDAO provides data access operations for activities
type ActivityDAO struct {
	db *sql.DB
}

// NewActivityDAO creates a new ActivityDAO instance
func NewActivityDAO(db *sql.DB) *ActivityDAO {
	return &ActivityDAO{db: db}
}

// Initialize creates the necessary tables if they don't exist
func (dao *ActivityDAO) Initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS activities (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            org TEXT NOT NULL,
            lead TEXT NOT NULL,
            co_leads TEXT,
            started_at DATETIME NOT NULL,
            created_by TEXT DEFAULT 'S',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
	}

	for _, q := range queries {
		_, err := dao.db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create inserts a new activity into the database
func (dao *ActivityDAO) Save(activity *Activity) (int64, error) {
	query := `
		INSERT INTO activities (name, org, lead, co_leads, started_at, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	coLeadsStr := strings.Join(activity.CoLeads, ",")

	result, err := dao.db.Exec(
		query,
		activity.Name,
		activity.Org,
		activity.Lead,
		coLeadsStr,
		activity.StartedAt,
		activity.CreatedBy,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return id, err
	}
	activity.ID = id
	return id, nil
}

// GetByID retrieves an activity by its ID
func (dao *ActivityDAO) GetByID(id int64) (*Activity, error) {
	query := `
		SELECT id, name, org, lead, co_leads, started_at, created_by, created_at, updated_at
		FROM activities
		WHERE id = ?
	`

	row := dao.db.QueryRow(query, id)

	var a Activity
	var coLeadsStr string

	err := row.Scan(
		&a.ID,
		&a.Name,
		&a.Org,
		&a.Lead,
		&coLeadsStr,
		&a.StartedAt,
		&a.CreatedBy,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if coLeadsStr != "" {
		a.CoLeads = strings.Split(coLeadsStr, ",")
	} else {
		a.CoLeads = []string{}
	}

	return &a, nil
}

// GetByDuration retrieves activities within a specific time range
func (dao *ActivityDAO) GetByDuration(startTime, endTime time.Time) ([]Activity, error) {
	query := `
		SELECT id, name, org, lead, co_leads, started_at, created_by, created_at, updated_at
		FROM activities
		WHERE started_at BETWEEN ? AND ?
		ORDER BY started_at ASC
	`

	rows, err := dao.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var a Activity
		var coLeadsStr string

		err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Org,
			&a.Lead,
			&coLeadsStr,
			&a.StartedAt,
			&a.CreatedBy,
			&a.CreatedAt,
			&a.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if coLeadsStr != "" {
			a.CoLeads = strings.Split(coLeadsStr, ",")
		} else {
			a.CoLeads = []string{}
		}

		activities = append(activities, a)
	}

	return activities, nil
}

// Update updates an existing activity
func (dao *ActivityDAO) Update(activity *Activity) error {
	query := `
		UPDATE activities
		SET name = ?, org = ?, lead = ?, co_leads = ?, started_at = ?, created_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	coLeadsStr := strings.Join(activity.CoLeads, ",")

	_, err := dao.db.Exec(
		query,
		activity.Name,
		activity.Org,
		activity.Lead,
		coLeadsStr,
		activity.StartedAt,
		activity.CreatedBy,
		activity.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete removes an activity by its ID
func (dao *ActivityDAO) Delete(id int64) (int64, error) {
	query := `DELETE FROM activities WHERE id = ?`

	res, err := dao.db.Exec(query, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

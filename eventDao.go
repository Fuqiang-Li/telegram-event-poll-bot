package main

import (
	"database/sql"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Struct to store the event details
type Event struct {
	ID          int64
	Description string
	DesiredPax  int
	MaxPax      int
	Options     []string
	ChatID      int64
	MessageID   int
	CreatedBy   string
	StartedAt   *time.Time // Using pointer to allow NULL values
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventUser struct {
	EventID int64
	User    string
	Option  string
}

// DAO layer for Event and EventUser
type EventDAO struct {
	db *sql.DB
}

func NewEventDAO(db *sql.DB) *EventDAO {
	return &EventDAO{db: db}
}

func (dao *EventDAO) Initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT,
			desired_pax INTEGER,
			max_pax INTEGER,
			options TEXT,
			chat_id INTEGER,
			message_id INTEGER,
			created_by TEXT,
			started_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS event_users (
			event_id INTEGER,
			user TEXT,
			option TEXT,
			FOREIGN KEY(event_id) REFERENCES events(id),
			UNIQUE(event_id, user, option)
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

func (dao *EventDAO) GetEventByID(eventID int64) (*Event, error) {
	query := `SELECT id, description, desired_pax, max_pax, options, chat_id, message_id, created_by,
		started_at, created_at, updated_at FROM events WHERE id = ?`
	row := dao.db.QueryRow(query, eventID)
	event := &Event{}
	var optionsStr string // Temporary variable to hold the options string
	err := row.Scan(
		&event.ID,
		&event.Description,
		&event.DesiredPax,
		&event.MaxPax,
		&optionsStr, // Scan into string first
		&event.ChatID,
		&event.MessageID,
		&event.CreatedBy,
		&event.StartedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Convert string back to slice
	if optionsStr != "" {
		event.Options = strings.Split(optionsStr, ";")
	}
	return event, nil
}

func (dao *EventDAO) SaveEvent(event *Event) (int64, error) {
	// Convert options slice to string
	optionsStr := strings.Join(event.Options, ";")

	query := `INSERT INTO events (
		description, desired_pax, max_pax, options, chat_id, message_id, created_by,
		started_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := dao.db.Exec(
		query,
		event.Description,
		event.DesiredPax,
		event.MaxPax,
		optionsStr, // Use the converted string instead of slice
		event.ChatID,
		event.MessageID,
		event.CreatedBy,
		event.StartedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (dao *EventDAO) UpdateEvent(event *Event) error {
	// Convert options slice to string
	optionsStr := strings.Join(event.Options, ";")

	query := `UPDATE events 
		SET description = ?, desired_pax = ?, max_pax = ?, options = ?, chat_id = ?, message_id = ?, created_by = ?,
		started_at = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query,
		event.Description,
		event.DesiredPax,
		event.MaxPax,
		optionsStr, // Use the converted string instead of slice
		event.ChatID,
		event.MessageID,
		event.CreatedBy,
		event.StartedAt,
		event.ID,
	)
	return err
}

func (dao *EventDAO) UpdateMessageID(eventID int64, messageID int) error {
	query := `UPDATE events 
		SET message_id = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query, messageID, eventID)
	return err
}

func (dao *EventDAO) GetEventByMessageID(messageID int) (*Event, error) {
	query := `SELECT id, description, desired_pax, max_pax, options, chat_id, message_id, created_by,
		started_at, created_at, updated_at FROM events WHERE message_id = ?`
	row := dao.db.QueryRow(query, messageID)
	event := &Event{}
	var optionsStr string // Temporary variable to hold the options string
	err := row.Scan(
		&event.ID,
		&event.Description,
		&event.DesiredPax,
		&event.MaxPax,
		&optionsStr, // Scan into string first
		&event.ChatID,
		&event.MessageID,
		&event.CreatedBy,
		&event.StartedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Convert string back to slice
	if optionsStr != "" {
		event.Options = strings.Split(optionsStr, ";")
	}
	return event, nil
}

func (dao *EventDAO) GetEventUsers(eventID int64) ([]EventUser, error) {
	query := "SELECT event_id, user, option FROM event_users WHERE event_id = ?"
	rows, err := dao.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []EventUser
	for rows.Next() {
		var eventUser EventUser
		err := rows.Scan(&eventUser.EventID, &eventUser.User, &eventUser.Option)
		if err != nil {
			return nil, err
		}
		users = append(users, eventUser)
	}
	return users, nil
}

func (dao *EventDAO) SaveEventUser(eventUser *EventUser) error {
	query := "INSERT INTO event_users (event_id, user, option) VALUES (?, ?, ?)"
	_, err := dao.db.Exec(query, eventUser.EventID, eventUser.User, eventUser.Option)
	return err
}

func (dao *EventDAO) DeleteEventUser(eventUser *EventUser) (int64, error) {
	query := "DELETE FROM event_users WHERE event_id = ? AND user = ? AND option = ?"
	result, err := dao.db.Exec(query, eventUser.EventID, eventUser.User, eventUser.Option)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Add a new method to update the start time
func (dao *EventDAO) UpdateStartTime(eventID int64, startTime time.Time) error {
	query := `UPDATE events 
		SET started_at = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query, startTime, eventID)
	return err
}

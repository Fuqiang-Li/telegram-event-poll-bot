package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Struct to store the event details
type Event struct {
	ID          int64
	Description string
	DesiredPax  int
	MaxPax      int
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
			FOREIGN KEY(event_id) REFERENCES events(id),
			UNIQUE(event_id, user)
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
	query := `SELECT id, description, desired_pax, max_pax, chat_id, message_id, created_by,
		started_at, created_at, updated_at FROM events WHERE id = ?`
	row := dao.db.QueryRow(query, eventID)
	event := &Event{}
	err := row.Scan(
		&event.ID,
		&event.Description,
		&event.DesiredPax,
		&event.MaxPax,
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
	return event, nil
}

func (dao *EventDAO) SaveEvent(event *Event) (int64, error) {
	query := `INSERT INTO events (
		description, desired_pax, max_pax, chat_id, message_id, created_by,
		started_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, err := dao.db.Exec(
		query,
		event.Description,
		event.DesiredPax,
		event.MaxPax,
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
	query := `UPDATE events 
		SET description = ?, desired_pax = ?, max_pax = ?, chat_id = ?, message_id = ?, created_by = ?,
		started_at = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query, event.Description, event.DesiredPax, event.MaxPax, event.ChatID, event.MessageID, event.CreatedBy,
		event.StartedAt, event.ID)
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
	query := `SELECT id, description, desired_pax, max_pax, chat_id, message_id, created_by,
		created_at, updated_at FROM events WHERE message_id = ?`
	row := dao.db.QueryRow(query, messageID)
	event := &Event{}
	err := row.Scan(
		&event.ID,
		&event.Description,
		&event.DesiredPax,
		&event.MaxPax,
		&event.ChatID,
		&event.MessageID,
		&event.CreatedBy,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (dao *EventDAO) GetEventUsers(eventID int64) ([]string, error) {
	query := "SELECT event_id, user FROM event_users WHERE event_id = ?"
	rows, err := dao.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var eventUser EventUser
		err := rows.Scan(&eventUser.EventID, &eventUser.User)
		if err != nil {
			return nil, err
		}
		users = append(users, eventUser.User)
	}
	return users, nil
}

func (dao *EventDAO) SaveEventUser(eventUser *EventUser) error {
	query := "INSERT INTO event_users (event_id, user) VALUES (?, ?)"
	_, err := dao.db.Exec(query, eventUser.EventID, eventUser.User)
	return err
}

func (dao *EventDAO) DeleteEventUser(eventUser *EventUser) (int64, error) {
	query := "DELETE FROM event_users WHERE event_id = ? AND user = ?"
	result, err := dao.db.Exec(query, eventUser.EventID, eventUser.User)
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

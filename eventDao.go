package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Struct to store the event details
type Event struct {
	ID          int64
	Description string
	Options     []string
	ChatID      int64
	MessageID   int
	CreatedBy   string
	CreatedByID int64
	StartedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventUser struct {
	EventID int64
	User    string
	UserID  int64
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
			options TEXT,
			chat_id INTEGER,
			message_id INTEGER,
			created_by TEXT,
			created_by_id,
			started_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS event_users (
			event_id INTEGER,
			user TEXT,
			user_id INTEGER,
			option TEXT,
			deleted BOOLEAN DEFAULT FALSE,
			FOREIGN KEY(event_id) REFERENCES events(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_event_users ON event_users (event_id, user_id, user, option)`,
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
	query := `SELECT id, description, options, chat_id, message_id, created_by, created_by_id,
		started_at, created_at, updated_at FROM events WHERE id = ?`
	row := dao.db.QueryRow(query, eventID)
	event := &Event{}
	var optionsStr string
	err := row.Scan(
		&event.ID,
		&event.Description,
		&optionsStr,
		&event.ChatID,
		&event.MessageID,
		&event.CreatedBy,
		&event.CreatedByID,
		&event.StartedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if optionsStr != "" {
		event.Options = strings.Split(optionsStr, ";")
	}
	return event, nil
}

func (dao *EventDAO) GetEventsByIDs(eventIDs []int64) ([]*Event, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}
	// Build the correct query for variable number of eventIDs
	placeholders := make([]string, len(eventIDs))
	args := make([]interface{}, len(eventIDs))
	for i, id := range eventIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	placeholderStr := strings.Join(placeholders, ",")
	query := `SELECT id, description, options, chat_id, message_id, created_by, created_by_id,
		started_at, created_at, updated_at FROM events WHERE id IN (` + placeholderStr + `)`
	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		var optionsStr string
		err := rows.Scan(
			&event.ID,
			&event.Description,
			&optionsStr,
			&event.ChatID,
			&event.MessageID,
			&event.CreatedBy,
			&event.CreatedByID,
			&event.StartedAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if optionsStr != "" {
			event.Options = strings.Split(optionsStr, ";")
		}
		events = append(events, &event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (dao *EventDAO) SaveEvent(event *Event) (int64, error) {
	optionsStr := strings.Join(event.Options, ";")

	query := `INSERT INTO events (
		description, options, chat_id, message_id, created_by, created_by_id,
		started_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := dao.db.Exec(
		query,
		event.Description,
		optionsStr,
		event.ChatID,
		event.MessageID,
		event.CreatedBy,
		event.CreatedByID,
		event.StartedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (dao *EventDAO) UpdateEvent(event *Event) error {
	optionsStr := strings.Join(event.Options, ";")

	query := `UPDATE events 
		SET description = ?, options = ?, chat_id = ?, message_id = ?, created_by = ?, created_by_id = ?,
		started_at = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query,
		event.Description,
		optionsStr,
		event.ChatID,
		event.MessageID,
		event.CreatedBy,
		event.CreatedByID,
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
	query := `SELECT id, description, options, chat_id, message_id, created_by, created_by_id,
		started_at, created_at, updated_at FROM events WHERE message_id = ?`
	row := dao.db.QueryRow(query, messageID)
	event := &Event{}
	var optionsStr string
	err := row.Scan(
		&event.ID,
		&event.Description,
		&optionsStr,
		&event.ChatID,
		&event.MessageID,
		&event.CreatedBy,
		&event.CreatedByID,
		&event.StartedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if optionsStr != "" {
		event.Options = strings.Split(optionsStr, ";")
	}
	return event, nil
}

func (dao *EventDAO) GetEventUsers(eventID int64) ([]EventUser, error) {
	query := "SELECT event_id, user, option, user_id FROM event_users WHERE event_id = ? and deleted = FALSE"
	rows, err := dao.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []EventUser
	for rows.Next() {
		var eventUser EventUser
		err := rows.Scan(&eventUser.EventID, &eventUser.User, &eventUser.Option, &eventUser.UserID)
		if err != nil {
			return nil, err
		}
		users = append(users, eventUser)
	}
	return users, nil
}

func (dao *EventDAO) SaveEventUser(eventUser *EventUser) error {
	query := "INSERT INTO event_users (event_id, user, option, user_id) VALUES (?, ?, ?, ?)"
	_, err := dao.db.Exec(query, eventUser.EventID, eventUser.User, eventUser.Option, eventUser.UserID)
	return err
}

func (dao *EventDAO) ToggleEventUser(eventUser *EventUser) error {
	// old unique index in db event_id, user, option
	query := "INSERT INTO event_users (event_id, user, option, user_id) VALUES (?, ?, ?, ?) ON CONFLICT(event_id, user, option) DO UPDATE SET deleted = NOT deleted, user_id = excluded.user_id"
	_, err := dao.db.Exec(query, eventUser.EventID, eventUser.User, eventUser.Option, eventUser.UserID)
	if err == nil {
		return nil
	}
	// new unique index in db event_id, user_id, user, option
	query = "INSERT INTO event_users (event_id, user, option, user_id) VALUES (?, ?, ?, ?) ON CONFLICT(event_id, user_id, user, option) DO UPDATE SET deleted = NOT deleted, user_id = excluded.user_id"
	_, err = dao.db.Exec(query, eventUser.EventID, eventUser.User, eventUser.Option, eventUser.UserID)
	return err
}

func (dao *EventDAO) DeleteEventUser(eventUser *EventUser) (int64, error) {
	query := "DELETE FROM event_users WHERE event_id = ? AND option = ? AND ((user_id = ?) OR (user = ? AND user_id = 0))"
	result, err := dao.db.Exec(query, eventUser.EventID, eventUser.Option, eventUser.UserID, eventUser.User)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (dao *EventDAO) GetEventUsersByUser(user string, userID int64) ([]EventUser, error) {
	query := `
		SELECT event_id, user, option, user_id
		FROM event_users
		WHERE (user_id = ? AND user_id != 0) OR (user = ? AND user_id = 0)
		AND not deleted
	`
	rows, err := dao.db.Query(query, userID, userID, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventUsers []EventUser
	for rows.Next() {
		var eventUser EventUser
		err := rows.Scan(&eventUser.EventID, &eventUser.User, &eventUser.Option, &eventUser.UserID)
		if err != nil {
			return nil, err
		}
		eventUsers = append(eventUsers, eventUser)
	}
	return eventUsers, nil
}

func (dao *EventDAO) GetEventsVotedByUser(user string, userID int64) ([]*Event, []EventUser, error) {
	fmt.Println(user, userID)
	eventUsers, err := dao.GetEventUsersByUser(user, userID)
	fmt.Println(eventUsers)
	if err != nil {
		return nil, nil, err
	}
	var eventIDs []int64
	for _, eu := range eventUsers {
		eventIDs = append(eventIDs, eu.EventID)
	}
	fmt.Println(eventIDs)
	events, err := dao.GetEventsByIDs(eventIDs)
	return events, eventUsers, err
}

// Add a new method to update the start time
func (dao *EventDAO) UpdateStartTime(eventID int64, startTime time.Time) error {
	query := `UPDATE events 
		SET started_at = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := dao.db.Exec(query, startTime, eventID)
	return err
}

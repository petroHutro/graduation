package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Image struct {
	Filename   string
	Base64Data []byte
}

type Event struct {
	ID              int
	UserID          int
	Title           string
	Description     string
	Place           string
	Participants    int
	MaxParticipants int
	Date            time.Time
	Active          bool
	Images          []Image
}

type EventUsers struct {
	ID     int
	Date   time.Time
	UserID []int
}

func (s *Storage) inTransaction(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin: %w", err)
	}

	if err := f(ctx, tx); err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("transaction Rollback failed: %w", err)
		}
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *Storage) GetImage(ctx context.Context, filename string) ([]byte, error) {
	var data []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT data
		FROM photo
		WHERE url = $1
	`, filename).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("photo not found: %w", &RepError{Err: err, UniqueViolation: true})
		}
		return nil, fmt.Errorf("cannot get photo: %w", err)
	}

	return data, nil
}

func (s *Storage) GetEventsToday(ctx context.Context, date time.Time) ([]EventUsers, error) {
	rowsEvent, err := s.db.QueryContext(ctx, `
		SELECT id, date
		FROM event
		WHERE EXTRACT(YEAR FROM date) = $1 AND EXTRACT(MONTH FROM date) = $2 AND EXTRACT(DAY FROM date) = $3
		AND active = true
		ORDER BY date
	`, date.Year(), date.Month(), date.Day())
	if err != nil || rowsEvent.Err() != nil {
		return nil, fmt.Errorf("cannot get events: %w", err)
	}
	defer rowsEvent.Close()

	var today []EventUsers
	for rowsEvent.Next() {
		var eventDate time.Time
		var eventID int
		err := rowsEvent.Scan(&eventID, &eventDate)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}

		rowsUser, err := s.db.QueryContext(ctx, `
			SELECT user_id
			FROM record
			WHERE event_id = $1
		`, eventID)
		if err != nil || rowsUser.Err() != nil {
			return nil, fmt.Errorf("cannot get record: %w", err)
		}
		defer rowsUser.Close()

		var users []int
		for rowsUser.Next() {
			var UserID int
			err := rowsUser.Scan(&UserID)
			if err != nil {
				return nil, fmt.Errorf("cannot scan: %w", err)
			}
			users = append(users, UserID)
		}
		today = append(today, EventUsers{ID: eventID, Date: eventDate, UserID: users})
	}

	return today, nil
}

func (s *Storage) EventsToday(ctx context.Context, date time.Time) error {
	events, err := s.GetEventsToday(ctx, date)
	if err != nil {
		return fmt.Errorf("cannot get events today: %w", err)
	}

	for _, event := range events {
		for _, user := range event.UserID {
			_, err = s.db.ExecContext(ctx, `
				INSERT INTO today (event_id, user_id, date)
				VALUES ($1, $2, $3)
				ON CONFLICT (event_id, user_id) DO NOTHING;
			`, event.ID, user, event.Date)
			if err != nil {
				return fmt.Errorf("cannot set: %w", err)
			}
		}
	}

	return nil
}

func (s *Storage) GetUserToday(ctx context.Context, date time.Time) (map[int]int, error) {
	day := date.Day()
	hour := date.Hour()
	rowsEvent, err := s.db.QueryContext(ctx, `
		SELECT event_id, user_id
		FROM today
		WHERE EXTRACT(DAY FROM date) <= $1 AND EXTRACT(HOUR FROM date) <= $2
	`, day, hour)
	if err != nil || rowsEvent.Err() != nil {
		return nil, fmt.Errorf("cannot get user_id: %w", err)
	}
	defer rowsEvent.Close()

	userEvent := make(map[int]int)
	for rowsEvent.Next() {
		var userID, eventID int
		err := rowsEvent.Scan(&userID, &eventID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}
		userEvent[userID] = eventID
	}

	return userEvent, nil
}

func send(eventID, userID int) error {
	fmt.Println("send message", eventID, userID)
	return nil
}

func (s *Storage) SendMessage(ctx context.Context, date time.Time) error {
	err := s.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		userEvent, err := s.GetUserToday(ctx, date)
		if err != nil {
			return fmt.Errorf("cannot get user today: %w", err)
		}

		for eventID, userID := range userEvent {
			err := send(eventID, userID)
			if err != nil {
				return fmt.Errorf("cannot send message: %w", err)
			}

			_, err = s.db.ExecContext(ctx, `
				DELETE FROM today WHERE event_id = $1 AND user_id = $2
			`, eventID, userID)
			if err != nil {
				return fmt.Errorf("cannot dell today: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot send message: %w", err)
	}

	return nil
}

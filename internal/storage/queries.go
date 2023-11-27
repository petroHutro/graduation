package storage

import (
	"context"
	"database/sql"
	"fmt"
	"graduation/internal/entity"
	"graduation/internal/utils"
	"time"
)

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

func (s *Storage) GetImage(ctx context.Context, filename string) (string, error) {
	// var data []byte

	// err := s.db.QueryRowContext(ctx, `
	// 	SELECT data
	// 	FROM photo
	// 	WHERE url = $1
	// `, filename).Scan(&data)
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		return nil, fmt.Errorf("photo not found: %w", &RepError{Err: err, UniqueViolation: true})
	// 	}
	// 	return nil, fmt.Errorf("cannot get photo: %w", err)
	// }

	// return data, nil
	url, err := s.ost.Get(filename)
	if err != nil {
		return "", fmt.Errorf("photo not found: %w", err)
	}
	return url, nil
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

func (s *Storage) getUserToday(ctx context.Context, date time.Time) (map[int]int, error) {
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

func (s *Storage) SendMessage(ctx context.Context, date time.Time, send func(mail, body string, urls []string) error) error {
	messages, err := s.getMessage(ctx, date)
	if err != nil {
		return fmt.Errorf("cannot get message: %w", err)
	}

	for _, message := range messages {
		err := send(message.Mail, message.Body, message.Urls)
		if err != nil {
			return fmt.Errorf("cannot send message: %w", err)
		}

		// _, err = s.db.ExecContext(ctx, `
		// 	DELETE FROM today WHERE event_id = $1 AND user_id = $2
		// `, message.EventID, message.UserID)
		// if err != nil {
		// 	return fmt.Errorf("cannot dell today: %w", err)
		// }
	}

	return nil
}

func (s *Storage) getMessage(ctx context.Context, date time.Time) ([]entity.Message, error) {
	userEvent, err := s.getUserToday(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("cannot get user today: %w", err)
	}

	var messages []entity.Message
	for eventID, userID := range userEvent {
		mail, err := s.getMail(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("cannot get mail: %w", err)
		}

		event, err := s.GetEvent(ctx, eventID)
		if err != nil {
			return nil, fmt.Errorf("cannot get event: %w", err)
		}

		body, err := s.getBody(ctx, event)
		if err != nil {
			return nil, fmt.Errorf("cannot get body: %w", err)
		}

		var urls []string
		for _, url := range event.Images {
			urls = append(urls, url.Filename)
		}

		messages = append(messages, entity.Message{
			EventID: eventID,
			UserID:  userID,
			Mail:    mail,
			Body:    body,
			Urls:    urls})
	}

	return messages, nil
}

func (s *Storage) getMail(ctx context.Context, userID int) (string, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT mail 
		FROM users WHERE id = $1;
	`, userID)

	var mail string
	err := row.Scan(&mail)
	if err != nil {
		return "", fmt.Errorf("cannot scan: %w", err)
	}

	return mail, nil
}

func (s *Storage) getBody(ctx context.Context, event *entity.Event) (string, error) {
	body, err := utils.GenerateHTML(event)
	if err != nil {
		return "", fmt.Errorf("cannot get event: %w", err)
	}

	return body, nil
}

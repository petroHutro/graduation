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

func (s *storageData) inTransaction(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error {
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

func (s *storageData) GetImage(ctx context.Context, filename string) (string, error) {
	url, err := s.ost.Get(filename)
	if err != nil {
		return "", fmt.Errorf("photo not found: %w", err)
	}
	return url, nil
}

func (s *storageData) EventsToday(ctx context.Context, date time.Time) error {
	events, err := s.getEventsToday(ctx, date)
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

	if err := s.dellEventToday(ctx, date); err != nil {
		return fmt.Errorf("cannot dell old event: %w", err)
	}

	return nil
}

func (s *storageData) SendMessage(ctx context.Context, date time.Time, send func(mail, body string, urls []string) error) error {
	messages, err := s.getMessage(ctx, date)
	if err != nil {
		return fmt.Errorf("cannot get message: %w", err)
	}

	for _, message := range messages {
		err := send(message.Mail, message.Body, message.Urls)
		if err != nil {
			return fmt.Errorf("cannot send message: %w", err)
		}

		_, err = s.db.ExecContext(ctx, `
			UPDATE today
				SET send = true
				WHERE event_id = $1 AND user_id = $2
		`, message.EventID, message.UserID)
		if err != nil {
			return fmt.Errorf("cannot update today: %w", err)
		}
	}

	return nil
}

func (s *storageData) getEventsToday(ctx context.Context, date time.Time) ([]EventUsers, error) {
	rowsEvent, err := s.db.QueryContext(ctx, `
		SELECT id, date
		FROM event
		WHERE EXTRACT(YEAR FROM date) = $1 AND EXTRACT(MONTH FROM date) = $2 AND EXTRACT(DAY FROM date) = $3
		AND active = true
		AND NOT EXISTS (
			SELECT 1
			FROM today
			WHERE today.event_id = event.id
		)
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

func (s *storageData) dellEventToday(ctx context.Context, date time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM today 
		WHERE send = TRUE
		AND (EXTRACT(YEAR FROM date) < $1 or EXTRACT(MONTH FROM date) < $2 or EXTRACT(DAY FROM date) < $3)
	`, date.Year(), date.Month(), date.Day())
	if err != nil {
		return fmt.Errorf("cannot dell today: %w", err)
	}

	return nil
}

func (s *storageData) getUserToday(ctx context.Context, date time.Time) (map[int]int, error) {
	formattedTime := date.Format("2006-01-02 15:04:05")
	rowsEvent, err := s.db.QueryContext(ctx, `
		SELECT event_id, user_id
		FROM today
		WHERE send = FALSE 
		AND ABS(EXTRACT(EPOCH FROM ($1 - date)) / 3600) <= 3
		ORDER BY date;
	`, formattedTime)
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

func (s *storageData) getMessage(ctx context.Context, date time.Time) ([]entity.Message, error) {
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

		var urls []string
		for index, image := range event.Images {
			url, err := s.GetImage(ctx, image.Filename)
			if err != nil {
				return nil, fmt.Errorf("cannot get url: %w", err)
			}
			event.Images[index].Filename = url
			urls = append(urls, url)
		}

		body, err := s.getBody(ctx, event)
		if err != nil {
			return nil, fmt.Errorf("cannot get body: %w", err)
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

func (s *storageData) getMail(ctx context.Context, userID int) (string, error) {
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

func (s *storageData) getBody(ctx context.Context, event *entity.Event) (string, error) {
	body, err := utils.GenerateHTML(event)
	if err != nil {
		return "", fmt.Errorf("cannot get event: %w", err)
	}

	return body, nil
}

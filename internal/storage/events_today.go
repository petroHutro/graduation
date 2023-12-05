package storage

import (
	"context"
	"fmt"
	"time"
)

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

	stmt, err := s.db.PrepareContext(ctx, `
		SELECT user_id
		FROM record
		WHERE event_id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("cannot creat Prepare: %w", err)
	}
	defer stmt.Close()

	var today []EventUsers
	for rowsEvent.Next() {
		var eventDate time.Time
		var eventID int
		err := rowsEvent.Scan(&eventID, &eventDate)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}

		rowsUser, err := stmt.QueryContext(ctx, eventID)
		if err != nil || rowsUser.Err() != nil {
			return nil, fmt.Errorf("cannot get record: %w", err)
		}
		defer rowsUser.Close()

		var users []int
		for rowsUser.Next() {
			var userID int
			err := rowsUser.Scan(&userID)
			if err != nil {
				return nil, fmt.Errorf("cannot scan: %w", err)
			}
			users = append(users, userID)
		}
		today = append(today, EventUsers{ID: eventID, Date: eventDate, UserID: users})
	}

	return today, nil
}

func (s *storageData) addEventsToday(ctx context.Context, date time.Time) error {
	events, err := s.getEventsToday(ctx, date)
	if err != nil {
		return fmt.Errorf("cannot get events today: %w", err)
	}

	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO today (event_id, user_id, date)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id) DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("cannot creat Prepare: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		for _, user := range event.UserID {
			_, err := stmt.ExecContext(ctx, event.ID, user, event.Date)
			if err != nil {
				return fmt.Errorf("cannot set: %w", err)
			}
		}
	}

	return nil
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

func (s *storageData) closeEventToday(ctx context.Context, date time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE event
		SET active = false
		WHERE active = TRUE
		AND (EXTRACT(YEAR FROM date) < $1 or EXTRACT(MONTH FROM date) < $2 or EXTRACT(DAY FROM date) < $3)
	`, date.Year(), date.Month(), date.Day())
	if err != nil {
		return fmt.Errorf("cannot close event today: %w", err)
	}

	return nil
}

func (s *storageData) EventsToday(ctx context.Context, date time.Time) error {
	if err := s.addEventsToday(ctx, date); err != nil {
		return fmt.Errorf("cannot add events today: %w", err)
	}

	if err := s.dellEventToday(ctx, date); err != nil {
		return fmt.Errorf("cannot dell old event: %w", err)
	}

	if err := s.closeEventToday(ctx, date.Add(-6*time.Hour)); err != nil {
		return fmt.Errorf("cannot dell old event: %w", err)
	}

	return nil
}

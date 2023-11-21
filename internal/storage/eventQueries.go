package storage

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

func (s *Storage) CreateEvent(ctx context.Context, e *Event) error {
	err := s.SetEvent(ctx, e)
	if err != nil {
		return fmt.Errorf("cannot set event: %w", err)
	}

	for _, image := range e.Images {
		err := s.SetEventPhoto(ctx, e.ID, image.Filename, image.Base64Data)
		if err != nil {
			return fmt.Errorf("cannot set url: %w", err)
		}
	}

	return nil
}

func (s *Storage) SetEvent(ctx context.Context, e *Event) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO event (user_id, title, description, place, participants, max_participants, date, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, e.UserID, e.Title, e.Description, e.Place, e.Participants, e.MaxParticipants, e.Date, e.Active).Scan(&e.ID)

	return err
}

func (s *Storage) SetEventPhoto(ctx context.Context, eventID int, url string, data []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO photo (event_id, url, data)
		VALUES ($1, $2, $3)
	`, eventID, url, data)

	return err
}

func (s *Storage) GetEvent(ctx context.Context, eventID int) (*Event, error) {
	event := &Event{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, description, place, participants, max_participants, date, active
		FROM event
		WHERE id = $1
	`, eventID).Scan(
		&event.ID,
		&event.UserID,
		&event.Title,
		&event.Description,
		&event.Place,
		&event.Participants,
		&event.MaxParticipants,
		&event.Date,
		&event.Active)
	if err != nil {
		return nil, fmt.Errorf("cannot get event: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT url
		FROM photo
		WHERE event_id = $1
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("cannot get urls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}
		event.Images = append(event.Images, Image{Filename: url})
	}
	return event, nil
}

func (s *Storage) GetEvents(ctx context.Context, from, to time.Time, limit, page int) ([]Event, int, error) {
	offset := (page - 1) * limit
	rowsE, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, place, participants, max_participants, date, active
		FROM event
		WHERE date BETWEEN $1 AND $2 AND active = true
		ORDER BY date LIMIT $3 OFFSET $4
	`, from, to, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot get events: %w", err)
	}
	defer rowsE.Close()

	var events []Event
	for rowsE.Next() {
		var event Event
		err := rowsE.Scan(
			&event.ID,
			&event.UserID,
			&event.Title,
			&event.Description,
			&event.Place,
			&event.Participants,
			&event.MaxParticipants,
			&event.Date,
			&event.Active)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot scan: %w", err)
		}

		rowsP, err := s.db.QueryContext(ctx, `
			SELECT url
			FROM photo
			WHERE event_id = $1
		`, event.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot get urls: %w", err)
		}
		defer rowsP.Close()

		for rowsP.Next() {
			var url string
			err := rowsP.Scan(&url)
			if err != nil {
				return nil, 0, fmt.Errorf("cannot scan: %w", err)
			}
			event.Images = append(event.Images, Image{Filename: url})
		}
		events = append(events, event)
	}

	var count int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT id)
		FROM event
		WHERE date BETWEEN $1 AND $2;
	`, from, to).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot get count event: %w", err)
	}

	count = int(math.Ceil(float64(count) / float64(limit)))

	return events, count, nil
}

func (s *Storage) DellPhoto(ctx context.Context, eventID int) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM photo WHERE event_id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("cannot dell photo from db: %w", err)
	}

	return nil
}

func (s *Storage) DellEvent(ctx context.Context, userID, eventID int) error {
	var id int
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM event
		WHERE id = $1 AND user_id = $2
	`, eventID, userID).Scan(&id)
	if err != nil {
		var flag bool

		err := s.db.QueryRowContext(ctx, `
			SELECT 1 FROM event
			WHERE id = $1
		`, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT event: %w", err), ForeignKeyViolation: true}
		}

		err = s.db.QueryRowContext(ctx, `
			SELECT 1 FROM event
			WHERE id = $1 AND user_id = $2
		`, userID, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT record: %w", err), UniqueViolation: true}
		}

		return fmt.Errorf("event not for user: %w", err)
	}

	if err := s.DellPhoto(ctx, eventID); err != nil {
		return fmt.Errorf("cannot dell photo: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM event WHERE id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("cannot dell event: %w", err)
	}

	return nil
}

func (s *Storage) CloseEvent(ctx context.Context, userID, eventID int) error {
	rows, err := s.db.ExecContext(ctx, `
		UPDATE event
			SET active = false
			WHERE id = $1 AND user_id = $2
	`, eventID, userID)
	if err != nil {
		return fmt.Errorf("cannot close event: %w", err)
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot get rows: %w", err)
	}

	if rowsAffected == 0 {
		var flag bool

		err := s.db.QueryRowContext(ctx, `
			SELECT 1 FROM event
			WHERE id = $1
		`, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT event: %w", err), ForeignKeyViolation: true}
		}

		err = s.db.QueryRowContext(ctx, `
			SELECT 1 FROM event
			WHERE id = $1 AND user_id = $2
		`, userID, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT record: %w", err), UniqueViolation: true}
		}

		return errors.New("0 close")
	}

	return nil
}

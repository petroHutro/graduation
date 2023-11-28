package storage

import (
	"context"
	"errors"
	"fmt"
	"graduation/internal/entity"
	"math"
	"time"
)

func (s *storageData) setEvent(ctx context.Context, e *entity.Event) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO event (user_id, title, description, place, participants, max_participants, date, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, e.UserID, e.Title, e.Description, e.Place, e.Participants, e.MaxParticipants, e.Date, e.Active).Scan(&e.ID)

	return err
}

func (s *storageData) setEventPhoto(ctx context.Context, eventID int, name string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO photo (event_id, name)
		VALUES ($1, $2)
	`, eventID, name)

	return err
}

func (s *storageData) CreateEvent(ctx context.Context, e *entity.Event) error {
	err := s.setEvent(ctx, e)
	if err != nil {
		return fmt.Errorf("cannot set event: %w", err)
	}

	for _, image := range e.Images {
		err := s.setEventPhoto(ctx, e.ID, image.Filename)
		if err != nil {
			return fmt.Errorf("cannot set photo db: %w", err)
		}
		if err := s.ost.Set(image.Filename, image.Base64Data); err != nil {
			return fmt.Errorf("cannot set photo ost: %w", err)
		}
	}

	return nil
}

func (s *storageData) GetImages(ctx context.Context, eventID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name
		FROM photo
		WHERE event_id = $1
	`, eventID)
	if err != nil || rows.Err() != nil {
		return nil, fmt.Errorf("cannot get urls: %w", err)
	}
	defer rows.Close()

	var filenames []string
	for rows.Next() {
		var filename string
		err := rows.Scan(&filename)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

func (s *storageData) GetEvent(ctx context.Context, eventID int) (*entity.Event, error) {
	event := &entity.Event{}
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

	urls, err := s.GetImages(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("cannot get images: %w", err)
	}
	for _, url := range urls {
		event.Images = append(event.Images, entity.Image{Filename: url})
	}

	return event, nil
}

func (s *storageData) GetEvents(ctx context.Context, from, to time.Time, limit, page int) ([]entity.Event, int, error) {
	offset := (page - 1) * limit
	rowsE, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, place, participants, max_participants, date, active
		FROM event
		WHERE date BETWEEN $1 AND $2 AND active = true
		ORDER BY date LIMIT $3 OFFSET $4
	`, from, to, limit, offset)
	if err != nil || rowsE.Err() != nil {
		return nil, 0, fmt.Errorf("cannot get events: %w", err)
	}
	defer rowsE.Close()

	var events []entity.Event
	for rowsE.Next() {
		var event entity.Event
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

		urls, err := s.GetImages(ctx, event.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot get images: %w", err)
		}
		for _, url := range urls {
			event.Images = append(event.Images, entity.Image{Filename: url})
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

func (s *storageData) dellPhoto(ctx context.Context, eventID int) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM photo WHERE event_id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("cannot dell photo from db: %w", err)
	}

	urls, err := s.GetImages(ctx, eventID)
	if err != nil {
		return fmt.Errorf("cannot get images: %w", err)
	}
	for _, url := range urls {
		if err := s.ost.Delete(url); err != nil {
			return fmt.Errorf("cannot dell ost: %w", err)
		}
	}

	return nil
}

func (s *storageData) DellEvent(ctx context.Context, userID, eventID int) error {
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

	if err := s.dellPhoto(ctx, eventID); err != nil {
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

func (s *storageData) CloseEvent(ctx context.Context, userID, eventID int) error {
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

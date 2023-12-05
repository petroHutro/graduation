package storage

import (
	"context"
	"errors"
	"fmt"
	"graduation/internal/entity"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *storageData) SetUser(ctx context.Context, login, password, mail string) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (login, password, mail)
		VALUES ($1, $2, $3)
		RETURNING id
	`, login, password, mail).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return 0, &RepError{Err: err, Repetition: true}
		}
		return 0, fmt.Errorf("cannot set database: %w", err)
	}

	return id, nil
}

func (s *storageData) GetUser(ctx context.Context, login, password string) (int, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id 
		FROM users WHERE login = $1 AND password = $2;
	`, login, password)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("cannot scan: %w", err)
	}

	return id, nil
}

func (s *storageData) GetUserEvents(ctx context.Context, userID int) ([]entity.Event, error) {
	rowsE, err := s.db.QueryContext(ctx, `
		SELECT event.id, event.user_id, event.title, event.description, event.place, event.participants, event.max_participants, event.date, event.active
		FROM event
		JOIN record ON event.id = record.event_id
		WHERE record.user_id = $1
		ORDER BY event.date
	`, userID)
	if err != nil || rowsE.Err() != nil {
		return nil, fmt.Errorf("cannot get events: %w", err)
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
			return nil, fmt.Errorf("cannot scan: %w", err)
		}

		urls, err := s.GetImages(ctx, event.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get images: %w", err)
		}
		for _, url := range urls {
			event.Images = append(event.Images, entity.Image{Filename: url})
		}

		events = append(events, event)
	}

	return events, nil
}

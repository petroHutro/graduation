package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"graduation/internal/entity"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) SetUser(ctx context.Context, login, password, mail string) (int, error) {
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

func (s *Storage) GetUser(ctx context.Context, login, password string) (int, error) {
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

func (s *Storage) AddEventUser(ctx context.Context, eventID, userID int) error {
	err := s.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO record (event_id, user_id)
			VALUES ($1, $2)
		`, eventID, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.UniqueViolation:
					if pgErr.ConstraintName == "record_event_id_user_id_key" {
						return &RepError{Err: err, UniqueViolation: true}
					}
				case pgerrcode.ForeignKeyViolation:
					if pgErr.ConstraintName == "record_event_id_fkey" {
						return &RepError{Err: err, ForeignKeyViolation: true}
					}
				default:
					return fmt.Errorf("cannot add user: %w", err)
				}
			}
		}

		rows, err := tx.ExecContext(ctx, `
			UPDATE event
				SET participants = participants + 1
				WHERE id = $1 AND participants < max_participants AND active = true
		`, eventID)
		if err != nil {
			return fmt.Errorf("cannot UPDATE event: %w", err)
		}

		rowsAffected, err := rows.RowsAffected()
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf("0 UPDATE: %w", &RepError{Err: err, UniqueViolation: true})
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot add evnt user: %w", err)
	}

	return nil
}

func (s *Storage) DellEventUser(ctx context.Context, eventID, userID int) error {
	err := s.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.ExecContext(ctx, `
			DELETE FROM record WHERE user_id = $1 AND event_id = $2
		`, userID, eventID)
		if err != nil {
			return fmt.Errorf("cannot dell record: %w", err)
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
				SELECT 1 FROM record
				WHERE user_id = $1 AND event_id = $2
			`, userID, eventID).Scan(&flag)
			if err != nil {
				return &RepError{Err: fmt.Errorf("cannot SELECT record: %w", err), UniqueViolation: true}
			}

			return errors.New("0 DELETE")
		}

		rows, err = tx.ExecContext(ctx, `
			UPDATE event
				SET participants = participants - 1
				WHERE id = $1
		`, eventID)
		if err != nil {
			return fmt.Errorf("cannot UPDATE event: %w", err)
		}

		rowsAffected, err = rows.RowsAffected()
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf("0 UPDATE: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot send message: %w", err)
	}

	return nil
}

func (s *Storage) GetUserEvents(ctx context.Context, userID int) ([]entity.Event, error) {
	rowsE, err := s.db.QueryContext(ctx, `
		SELECT event.id, event.user_id, event.title, event.description, event.place, event.participants, event.max_participants, event.date, event.active
		FROM event
		JOIN record ON event.id = record.event_id
		WHERE record.user_id = $1
		ORDER BY event.date
	`, userID)
	if err != nil {
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

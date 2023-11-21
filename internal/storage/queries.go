package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) SetUser(ctx context.Context, login, password string) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id
	`, login, password).Scan(&id)

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

type Image struct {
	Filename   string
	Base64Data []byte
}

type Event struct {
	ID              int
	User_id         int
	Title           string
	Description     string
	Place           string
	Participants    int
	MaxParticipants int
	Date            time.Time
	Active          bool
	Images          []Image
}

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
	`, e.User_id, e.Title, e.Description, e.Place, e.Participants, e.MaxParticipants, e.Date, e.Active).Scan(&e.ID)

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
		&event.User_id,
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
			&event.User_id,
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

func (s *Storage) AddEventUser(ctx context.Context, eventID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
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

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *Storage) DellEventUser(ctx context.Context, eventID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin: %w", err)
	}

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

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *Storage) GetUserEvents(ctx context.Context, userID int) ([]Event, error) {
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

	var events []Event
	for rowsE.Next() {
		var event Event
		err := rowsE.Scan(
			&event.ID,
			&event.User_id,
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

		rowsP, err := s.db.QueryContext(ctx, `
			SELECT url
			FROM photo
			WHERE event_id = $1
		`, event.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get urls: %w", err)
		}
		defer rowsP.Close()

		for rowsP.Next() {
			var url string
			err := rowsP.Scan(&url)
			if err != nil {
				return nil, fmt.Errorf("cannot scan: %w", err)
			}
			event.Images = append(event.Images, Image{Filename: url})
		}
		events = append(events, event)
	}
	return events, nil
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

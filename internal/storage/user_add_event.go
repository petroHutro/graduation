package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"graduation/internal/entity"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func creatTicket(ctx context.Context, tx *sql.Tx, ticket *entity.Ticket) error {
	date := time.Now()
	date = date.Add(time.Hour * time.Duration(ticket.Exp))
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ticket (token, event_id, user_id, date)
		VALUES ($1, $2, $3, $4)
	`, ticket.Token, ticket.EventID, ticket.UserID, date)
	if err != nil {
		return fmt.Errorf("cannot INSERT ticket: %w", err)
	}

	return nil
}

func addRecord(ctx context.Context, tx *sql.Tx, eventID, userID int) error {
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
				return fmt.Errorf("cannot INSERT record: %w", err)
			}
		}
	}

	return nil
}

func addCountUser(ctx context.Context, tx *sql.Tx, eventID int) error {
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
		return fmt.Errorf("0 UPDATE: %w", err)
	}

	return nil
}

func (s *storageData) AddEventUser(ctx context.Context, tick *entity.Ticket) error {
	err := s.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := addRecord(ctx, tx, tick.EventID, tick.UserID); err != nil {
			return fmt.Errorf("cannot addRecord: %w", err)
		}

		if err := addCountUser(ctx, tx, tick.EventID); err != nil {
			return fmt.Errorf("cannot addCountUser: %w", err)
		}

		if err := creatTicket(ctx, tx, tick); err != nil {
			return fmt.Errorf("cannot creatTicket: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot add evnt user: %w", err)
	}

	return nil
}

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func dellTicket(ctx context.Context, tx *sql.Tx, userID, eventID int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM ticket WHERE user_id = $1 AND event_id = $2
	`, userID, eventID)
	if err != nil {
		return fmt.Errorf("cannot dell ticket: %w", err)
	}

	return nil
}

func dellRecoed(ctx context.Context, tx *sql.Tx, userID, eventID int) error {
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

		err := tx.QueryRowContext(ctx, `
				SELECT 1 FROM event
				WHERE id = $1
			`, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT event: %w", err), ForeignKeyViolation: true}
		}

		err = tx.QueryRowContext(ctx, `
				SELECT 1 FROM record
				WHERE user_id = $1 AND event_id = $2
			`, userID, eventID).Scan(&flag)
		if err != nil {
			return &RepError{Err: fmt.Errorf("cannot SELECT record: %w", err), UniqueViolation: true}
		}

		return errors.New("0 DELETE")
	}

	return nil
}

func dellCountUser(ctx context.Context, tx *sql.Tx, eventID int) error {
	rows, err := tx.ExecContext(ctx, `
		UPDATE event
			SET participants = participants - 1
			WHERE id = $1
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

func (s *storageData) DellEventUser(ctx context.Context, eventID, userID int) error {
	err := s.inTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {

		if err := dellRecoed(ctx, tx, userID, eventID); err != nil {
			return fmt.Errorf("cannot dell record: %w", err)
		}

		if err := dellCountUser(ctx, tx, eventID); err != nil {
			return fmt.Errorf("cannot dell count user: %w", err)
		}

		if err := dellTicket(ctx, tx, userID, eventID); err != nil {
			return fmt.Errorf("cannot dell ticket: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cannot dell: %w", err)
	}

	return nil
}

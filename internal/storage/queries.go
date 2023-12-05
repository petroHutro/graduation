package storage

import (
	"context"
	"database/sql"
	"fmt"
	"graduation/internal/entity"
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

func (s *storageData) UserTickets(ctx context.Context, userID int) ([]entity.Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT token, event_id, active
		FROM ticket
		WHERE user_id = $1
	`, userID)
	if err != nil || rows.Err() != nil {
		return nil, fmt.Errorf("cannot get record: %w", err)
	}
	defer rows.Close()

	var tickets []entity.Ticket
	for rows.Next() {
		var ticket entity.Ticket
		err := rows.Scan(&ticket.Token, &ticket.EventID, &ticket.Status)
		if err != nil {
			return nil, fmt.Errorf("cannot scan: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (s *storageData) GetImage(ctx context.Context, filename string) (string, error) {
	url, err := s.ost.Get(filename)
	if err != nil {
		return "", fmt.Errorf("photo not found: %w", err)
	}
	return url, nil
}

func (s *storageData) MessageUpdate(ctx context.Context, eventID, userID int) error {
	_, err := s.db.ExecContext(ctx, `
			UPDATE today
			SET send = true
			WHERE event_id = $1 AND user_id = $2
		`, eventID, userID)
	if err != nil {
		return fmt.Errorf("cannot update today: %w", err)
	}

	return nil
}

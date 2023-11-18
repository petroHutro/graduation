package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Storage struct {
	db *sql.DB
}

type RepError struct {
	Err                 error
	Repetition          bool
	UniqueViolation     bool
	ForeignKeyViolation bool
}

func (e *RepError) Error() string {
	return e.Err.Error()
}

func newStorage(databaseDSN string) (*Storage, error) {
	db, err := Connection(databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot connection database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

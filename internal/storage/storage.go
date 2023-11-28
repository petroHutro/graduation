package storage

import (
	"context"
	"database/sql"
	"fmt"
	"graduation/internal/entity"
	"graduation/internal/objectstorage"

	"time"
)

type Storage interface {
	SetUser(ctx context.Context, login, password, mail string) (int, error)
	GetUser(ctx context.Context, login, password string) (int, error)
	AddEventUser(ctx context.Context, eventID, userID int) error
	DellEventUser(ctx context.Context, eventID, userID int) error
	GetUserEvents(ctx context.Context, userID int) ([]entity.Event, error)
	GetEvent(ctx context.Context, eventID int) (*entity.Event, error)
	GetEvents(ctx context.Context, from, to time.Time, limit, page int) ([]entity.Event, int, error)
	GetImage(ctx context.Context, filename string) (string, error)
	DellEvent(ctx context.Context, userID, eventID int) error
	CreateEvent(ctx context.Context, e *entity.Event) error
	CloseEvent(ctx context.Context, userID, eventID int) error
	SendMessage(ctx context.Context, date time.Time, send func(mail, body string, urls []string) error) error
}

type storageData struct {
	db  *sql.DB
	ost *objectstorage.Storage
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

func newStorage(databaseDSN string) (*storageData, error) {
	db, err := Connection(databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot connection database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	ost, err := objectstorage.Connect()
	if err != nil {
		return nil, fmt.Errorf("cannot connection object storage: %w", err)
	}

	return &storageData{db: db, ost: ost}, nil
}

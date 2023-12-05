package storage

import (
	"context"
	"database/sql"
	"graduation/internal/entity"
	"graduation/internal/ostorage"

	"time"
)

//go:generate mockgen -source=storage.go -destination=mock/mock.go -package=mock

type UserStorage interface {
	SetUser(ctx context.Context, login, password, mail string) (int, error)
	GetUser(ctx context.Context, login, password string) (int, error)
	AddEventUser(ctx context.Context, tick *entity.Ticket) error
	DellEventUser(ctx context.Context, eventID, userID int) error
	GetUserEvents(ctx context.Context, userID int) ([]entity.Event, error)
	UserTickets(ctx context.Context, userID int) ([]entity.Ticket, error)
}

type EventStorage interface {
	GetEvent(ctx context.Context, eventID int) (*entity.Event, error)
	GetEvents(ctx context.Context, from, to time.Time, limit, page int) ([]entity.Event, int, error)
	GetImage(ctx context.Context, filename string) (string, error)
	DellEvent(ctx context.Context, userID, eventID int) error
	CreateEvent(ctx context.Context, e *entity.Event) error
	CloseEvent(ctx context.Context, userID, eventID int) error
	GetDateEvent(ctx context.Context, eventID int) (int, error)
}

type NotificationStorage interface {
	GetMessages(ctx context.Context, date time.Time) ([]entity.Message, error)
	MessageUpdate(ctx context.Context, eventID, userID int) error
	EventsToday(ctx context.Context, date time.Time) error
}

type Storage interface {
	UserStorage
	EventStorage
	NotificationStorage
}

type storageData struct {
	db  *sql.DB
	ost *ostorage.Storage
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

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"graduation/internal/config"
	"graduation/internal/ostorage"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func newStorage(conf *config.Storage) (*storageData, error) {
	db, err := Connection(conf.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot connection database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	ost, err := ostorage.Connect(&conf.ObjectStorage)
	if err != nil {
		return nil, fmt.Errorf("cannot connection object storage: %w", err)
	}

	return &storageData{db: db, ost: ost}, nil
}

func Connection(databaseDSN string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open DataBase: %w", err)
	}

	return db, nil
}

func InitStorage(conf *config.Storage) (Storage, error) {
	st, err := newStorage(conf)
	if err != nil {
		return nil, fmt.Errorf("cannot create data base: %w", err)
	}

	_, err = goose.GetDBVersion(st.db)
	if err != nil {
		return nil, err
	}

	err = goose.Up(st.db, "./internal/migration")
	if err != nil && err != goose.ErrNoNextVersion {
		return nil, err
	}

	return st, nil
}

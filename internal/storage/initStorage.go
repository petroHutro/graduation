package storage

import (
	"database/sql"
	"fmt"

	"graduation/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func Connection(databaseDSN string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open DataBase: %w", err)
	}

	return db, nil
}

func InitStorage(conf *config.Storage) (*Storage, error) {
	st, err := newStorage(conf.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot create data base: %w", err)
	}

	_, err = goose.GetDBVersion(st.db)
	if err != nil {
		return nil, err
	}

	err = goose.Up(st.db, "/Users/petro/GoProjects/graduation/internal/migration")
	if err != nil && err != goose.ErrNoNextVersion {
		return nil, err
	}

	return st, nil
}

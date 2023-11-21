package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"graduation/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func (s *Storage) createTable(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN

				CREATE TABLE users (
					id 			SERIAL PRIMARY KEY,
					login		TEXT NOT NULL,
					password	TEXT NOT NULL,
					UNIQUE 		(login)
				);

				CREATE TABLE event (
					id 					SERIAL PRIMARY KEY,
					user_id				INT REFERENCES users(id) ON DELETE CASCADE,
					title 				TEXT NOT NULL,
					description 		TEXT NOT NULL,
					place 				TEXT NOT NULL,
					participants		INT DEFAULT 0,
					max_participants	INT DEFAULT 0,
					date 				timestamp,
					active 				BOOLEAN DEFAULT FALSE
				);

				CREATE TABLE record (
					id 			SERIAL PRIMARY KEY,
					event_id	INT REFERENCES event(id) ON DELETE CASCADE,
					user_id		INT REFERENCES users(id) ON DELETE CASCADE,
					UNIQUE 		(event_id, user_id)
				);

				CREATE TABLE today (
					id 			SERIAL PRIMARY KEY,
					event_id	INT REFERENCES event(id) ON DELETE CASCADE,
					user_id		INT REFERENCES users(id) ON DELETE CASCADE,
					date 		timestamp,
					UNIQUE 		(event_id, user_id)
				);

				CREATE TABLE photo (
					id 			SERIAL PRIMARY KEY,
					event_id	INT REFERENCES event(id) ON DELETE CASCADE,
					url 		TEXT NOT NULL,
					data 		BYTEA NOT NULL
				);

			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("cannot request create table: %w", err)
	}

	// _, _ = tx.ExecContext(ctx, `
	// 	DROP TABLE user ;
	// `)

	return tx.Commit()
}

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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := st.createTable(ctx); err != nil {
		return nil, fmt.Errorf("cannot create table: %w", err)
	}

	return st, nil
}

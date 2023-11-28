-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS users (
	id 			SERIAL PRIMARY KEY,
	login		TEXT NOT NULL,
	password	TEXT NOT NULL,
	mail		TEXT NOT NULL,
	UNIQUE 		(login)
);

CREATE TABLE IF NOT EXISTS event (
	id 					SERIAL PRIMARY KEY,
	user_id				INT REFERENCES users(id) ON DELETE CASCADE,
	title 				TEXT NOT NULL,
	description 		TEXT NOT NULL,
	place 				TEXT NOT NULL,
	participants		INT DEFAULT 0,
	max_participants	INT DEFAULT 0,
	date 				timestamp,
	active 				BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS record (
	id 			SERIAL PRIMARY KEY,
	event_id	INT REFERENCES event(id) ON DELETE CASCADE,
	user_id		INT REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE 		(event_id, user_id)
);

CREATE TABLE IF NOT EXISTS today (
	id 			SERIAL PRIMARY KEY,
	event_id	INT REFERENCES event(id) ON DELETE CASCADE,
	user_id		INT REFERENCES users(id) ON DELETE CASCADE,
	date 		timestamp,
	send 		BOOLEAN DEFAULT FALSE,
	UNIQUE 		(event_id, user_id)
);

CREATE TABLE IF NOT EXISTS photo (
	id 			SERIAL PRIMARY KEY,
	event_id	INT REFERENCES event(id) ON DELETE CASCADE,
	name 		TEXT NOT NULL
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

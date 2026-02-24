package db

import "database/sql"

// Migrate ensures required database tables exist.
func Migrate(db *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	classcode TEXT[] NOT NULL DEFAULT '{}',
	role TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS queues (
	id SERIAL PRIMARY KEY,
	course_id INT NOT NULL DEFAULT 0,
	ta_id INT NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'paused', 'closed')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS queue_entries (
	id SERIAL PRIMARY KEY,
	queue_id INT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
	student_id INT NOT NULL REFERENCES users(id),
	position INT NOT NULL,
	joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(queue_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_queue_entries_queue_position ON queue_entries(queue_id, position);
`
	_, err := db.Exec(query)
	return err
}


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
);`

	_, err := db.Exec(query)
	return err
}


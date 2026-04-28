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

CREATE TABLE IF NOT EXISTS office_hours (
	id SERIAL PRIMARY KEY,
	ta_id INT NOT NULL REFERENCES users(id),
	course_id INT NOT NULL,
	day_of_week SMALLINT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
	start_time TIME NOT NULL,
	end_time TIME NOT NULL,
	location TEXT NOT NULL DEFAULT '',
	CHECK (start_time < end_time)
);

CREATE INDEX IF NOT EXISTS idx_office_hours_ta_id ON office_hours(ta_id);
CREATE INDEX IF NOT EXISTS idx_office_hours_course_id ON office_hours(course_id);

CREATE TABLE IF NOT EXISTS queue_announcements (
	id SERIAL PRIMARY KEY,
	queue_id INT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
	ta_id INT NOT NULL REFERENCES users(id),
	message TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_queue_announcements_queue_created ON queue_announcements(queue_id, created_at DESC);

CREATE TABLE IF NOT EXISTS sessions (
	id SERIAL PRIMARY KEY,
	queue_id INT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
	ta_id INT NOT NULL REFERENCES users(id),
	student_id INT NOT NULL REFERENCES users(id),
	zoom_meeting_id TEXT NOT NULL DEFAULT '',
	zoom_join_url TEXT NOT NULL DEFAULT '',
	zoom_start_url TEXT NOT NULL DEFAULT '',
	zoom_passcode TEXT NOT NULL DEFAULT '',
	started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_queue_started ON sessions(queue_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_ta ON sessions(ta_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_student ON sessions(student_id, started_at DESC);
`
	_, err := db.Exec(query)
	return err
}

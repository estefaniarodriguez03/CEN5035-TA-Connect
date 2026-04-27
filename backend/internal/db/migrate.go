package db

import "database/sql"

// Migrate ensures required database tables exist.
func Migrate(db *sql.DB) error {
	const createQuery = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	classcode TEXT[] NOT NULL DEFAULT '{}',
	role TEXT NOT NULL,
	average_session_duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
	session_sample_count INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS queues (
	id SERIAL PRIMARY KEY,
	course_id INT NOT NULL DEFAULT 0,
	ta_id INT NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'paused', 'closed')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	average_session_duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
	session_sample_count INT NOT NULL DEFAULT 0,
	last_served_at TIMESTAMPTZ
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

CREATE TABLE IF NOT EXISTS session_logs (
	id SERIAL PRIMARY KEY,
	student_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	ta_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	start_time TIMESTAMPTZ NOT NULL,
	end_time TIMESTAMPTZ NOT NULL,
	duration_seconds DOUBLE PRECISION NOT NULL,
	CHECK (end_time >= start_time),
	CHECK (duration_seconds >= 0)
);

CREATE INDEX IF NOT EXISTS idx_session_logs_ta_id ON session_logs(ta_id);
CREATE INDEX IF NOT EXISTS idx_session_logs_student_id ON session_logs(student_id);
CREATE INDEX IF NOT EXISTS idx_session_logs_start_time ON session_logs(start_time DESC);
`
	_, err := db.Exec(createQuery)
	if err != nil {
		return err
	}
	// Idempotent column adds for existing deployments created before these fields.
	const alter = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS average_session_duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS session_sample_count INT NOT NULL DEFAULT 0;
ALTER TABLE queues ADD COLUMN IF NOT EXISTS average_session_duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE queues ADD COLUMN IF NOT EXISTS session_sample_count INT NOT NULL DEFAULT 0;
ALTER TABLE queues ADD COLUMN IF NOT EXISTS last_served_at TIMESTAMPTZ;
`
	_, err = db.Exec(alter)
	return err
}

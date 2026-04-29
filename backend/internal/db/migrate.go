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

CREATE TABLE IF NOT EXISTS courses (
	id SERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ta_courses (
	ta_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (ta_id, course_id)
);

CREATE INDEX IF NOT EXISTS idx_ta_courses_ta_id ON ta_courses(ta_id);
CREATE INDEX IF NOT EXISTS idx_ta_courses_course_id ON ta_courses(course_id);

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
ALTER TABLE queues ADD COLUMN IF NOT EXISTS serving_student_id INT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE courses ADD COLUMN IF NOT EXISTS code TEXT;
ALTER TABLE courses ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE courses ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE courses
SET code = 'COURSE-' || id::text
WHERE code IS NULL OR code = '';

ALTER TABLE courses ALTER COLUMN code SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_courses_code_unique ON courses(code);

-- Ensure a legacy default course row exists because older codepaths used course_id=0.
INSERT INTO courses (id, code, name)
VALUES (0, 'LEGACY-0', 'Legacy Course')
ON CONFLICT (id) DO NOTHING;

-- Backfill referenced course ids seen in existing rows so FK constraints can be added safely.
INSERT INTO courses (id, code, name)
SELECT DISTINCT q.course_id, 'AUTO-' || q.course_id::text, 'Auto-created course ' || q.course_id::text
FROM queues q
LEFT JOIN courses c ON c.id = q.course_id
WHERE q.course_id IS NOT NULL AND c.id IS NULL
ON CONFLICT (id) DO NOTHING;

INSERT INTO courses (id, code, name)
SELECT DISTINCT o.course_id, 'AUTO-' || o.course_id::text, 'Auto-created course ' || o.course_id::text
FROM office_hours o
LEFT JOIN courses c ON c.id = o.course_id
WHERE o.course_id IS NOT NULL AND c.id IS NULL
ON CONFLICT (id) DO NOTHING;

SELECT setval(
	pg_get_serial_sequence('courses', 'id'),
	COALESCE((SELECT MAX(id) FROM courses), 1),
	true
);

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_queues_course_id'
	) THEN
		ALTER TABLE queues
		ADD CONSTRAINT fk_queues_course_id
		FOREIGN KEY (course_id) REFERENCES courses(id);
	END IF;
END $$;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_office_hours_course_id'
	) THEN
		ALTER TABLE office_hours
		ADD CONSTRAINT fk_office_hours_course_id
		FOREIGN KEY (course_id) REFERENCES courses(id);
	END IF;
END $$;

CREATE OR REPLACE FUNCTION ensure_ta_role_for_ta_courses() RETURNS trigger AS $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM users u
		WHERE u.id = NEW.ta_id AND u.role = 'ta'
	) THEN
		RAISE EXCEPTION 'ta_courses.ta_id % does not reference a TA user', NEW.ta_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_trigger WHERE tgname = 'trg_ta_courses_ta_role'
	) THEN
		CREATE TRIGGER trg_ta_courses_ta_role
		BEFORE INSERT OR UPDATE ON ta_courses
		FOR EACH ROW
		EXECUTE FUNCTION ensure_ta_role_for_ta_courses();
	END IF;
END $$;
`
	_, err = db.Exec(alter)
	return err
}

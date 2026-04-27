package sessionlog

import (
	"context"
	"database/sql"
	"time"
)

// InsertCompleted appends a finished help session (one /next to the next /next for the same queue).
func InsertCompleted(ctx context.Context, tx *sql.Tx, studentID, taID int, start, end time.Time, durationSec float64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO session_logs (student_id, ta_id, start_time, end_time, duration_seconds)
		VALUES ($1, $2, $3, $4, $5)
	`, studentID, taID, start, end, durationSec)
	return err
}

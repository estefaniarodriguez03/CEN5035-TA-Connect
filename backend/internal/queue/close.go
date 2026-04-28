package queue

import (
	"context"
	"database/sql"
	"log"
	"time"

	"backend/internal/sessionlog"
)

// applyClosedQueueFinalization runs in a transaction when the queue is set to closed:
// - removes all waiting queue entries;
// - if a student was in an active help session, writes session_logs and updates queue/TA rolling stats;
// - clears last_served_at and serving_student_id; sets status to closed.
func applyClosedQueueFinalization(ctx context.Context, tx *sql.Tx, queueID, taID int) error {
	var lastServed sql.NullTime
	var serving sql.NullInt64
	err := tx.QueryRowContext(ctx,
		`SELECT last_served_at, serving_student_id FROM queues WHERE id = $1 AND ta_id = $2 FOR UPDATE`,
		queueID, taID,
	).Scan(&lastServed, &serving)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM queue_entries WHERE queue_id = $1`, queueID); err != nil {
		return err
	}

	var sessionEnd time.Time
	var sessionSec float64
	if lastServed.Valid {
		sessionEnd = time.Now()
		sessionSec = sessionEnd.Sub(lastServed.Time).Seconds()
	}

	if lastServed.Valid && serving.Valid {
		if err := sessionlog.InsertCompleted(ctx, tx, int(serving.Int64), taID, lastServed.Time, sessionEnd, sessionSec); err != nil {
			return err
		}
		var ulock int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, taID).Scan(&ulock); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET
				session_sample_count = session_sample_count + 1,
				average_session_duration_seconds = CASE
					WHEN session_sample_count = 0 THEN $2
					ELSE (average_session_duration_seconds * session_sample_count + $2) / (session_sample_count + 1)
				END
			WHERE id = $1
		`, taID, sessionSec); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE queues SET
				session_sample_count = session_sample_count + 1,
				average_session_duration_seconds = CASE
					WHEN session_sample_count = 0 THEN $2
					ELSE (average_session_duration_seconds * session_sample_count + $2) / (session_sample_count + 1)
				END
			WHERE id = $1
		`, queueID, sessionSec); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE queues SET
			status = 'closed',
			last_served_at = NULL,
			serving_student_id = NULL
		WHERE id = $1
	`, queueID); err != nil {
		return err
	}
	log.Printf("queue closed: queue_id=%d ta_id=%d", queueID, taID)
	return nil
}

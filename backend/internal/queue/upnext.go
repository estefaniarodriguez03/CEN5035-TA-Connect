package queue

import (
	"context"
	"database/sql"
)

// PublishUpNextForQueue emits STUDENT_UP_NEXT for the single student at the front of the queue
// (position 1). Clients can show “you’re next” only when payload.student_id matches the viewer.
func PublishUpNextForQueue(ctx context.Context, db *sql.DB, queueID int) {
	var studentID int
	var position int
	err := db.QueryRowContext(ctx, `
		SELECT student_id, position
		FROM queue_entries
		WHERE queue_id = $1 AND position = 1
		ORDER BY joined_at ASC
		LIMIT 1
	`, queueID).Scan(&studentID, &position)
	if err != nil {
		// No one at the head (empty queue or scan error).
		return
	}

	DefaultHub.Publish(queueID, QueueEvent{
		Type:    EventStudentUpNext,
		QueueID: queueID,
		Payload: map[string]any{
			"student_id": studentID,
			"position":   position,
		},
	})
}

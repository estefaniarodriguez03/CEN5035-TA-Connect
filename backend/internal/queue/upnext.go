package queue

import (
	"context"
	"database/sql"
	"os"
	"strconv"
)

// upNextThreshold returns how many positions from the front count as "about to be your turn".
// Configured with QUEUE_UP_NEXT_THRESHOLD (default 3).
func upNextThreshold() int {
	s := os.Getenv("QUEUE_UP_NEXT_THRESHOLD")
	if s == "" {
		return 3
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 3
	}
	return n
}

// PublishUpNextForQueue emits STUDENT_UP_NEXT when at least one entry has position <= threshold.
func PublishUpNextForQueue(ctx context.Context, db *sql.DB, queueID int) {
	th := upNextThreshold()
	rows, err := db.QueryContext(ctx, `
		SELECT student_id, position
		FROM queue_entries
		WHERE queue_id = $1 AND position <= $2
		ORDER BY position ASC, joined_at ASC
	`, queueID, th)
	if err != nil {
		return
	}
	defer rows.Close()

	var students []map[string]any
	for rows.Next() {
		var studentID, position int
		if err := rows.Scan(&studentID, &position); err != nil {
			return
		}
		students = append(students, map[string]any{
			"student_id": studentID,
			"position":   position,
		})
	}
	if len(students) == 0 {
		return
	}

	DefaultHub.Publish(queueID, QueueEvent{
		Type:    EventStudentUpNext,
		QueueID: queueID,
		Payload: map[string]any{
			"threshold": th,
			"students":  students,
		},
	})
}

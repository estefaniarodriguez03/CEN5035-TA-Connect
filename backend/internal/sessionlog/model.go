package sessionlog

import "time"

// SessionLog is a completed office-hour help session (one row in session_logs).
// Duration is stored in seconds in the database as duration_seconds.
type SessionLog struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	TAID      int       `json:"ta_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  float64   `json:"duration"`
}

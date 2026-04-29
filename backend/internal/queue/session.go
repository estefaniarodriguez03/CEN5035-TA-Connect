package queue

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/auth"
	"backend/internal/zoom"
)

// StartSessionRequest is the JSON body for POST /api/queues/{id}/session.
// student_id is required and must reference a user that is currently in the queue.
type StartSessionRequest struct {
	StudentID int    `json:"student_id"`
	Topic     string `json:"topic,omitempty"`
}

// Session is one persisted TA<->student office-hour session backed by a Zoom meeting.
type Session struct {
	ID            int    `json:"id"`
	QueueID       int    `json:"queue_id"`
	TAID          int    `json:"ta_id"`
	StudentID     int    `json:"student_id"`
	StudentName   string `json:"student_name,omitempty"`
	ZoomMeetingID string `json:"zoom_meeting_id"`
	ZoomJoinURL   string `json:"zoom_join_url"`
	ZoomStartURL  string `json:"zoom_start_url"`
	ZoomPasscode  string `json:"zoom_passcode"`
	StartedAt     string `json:"started_at"`
}

// StartSession handles POST /api/queues/{id}/session.
//
// Behavior:
//   - Auth: TA only, must own the queue.
//   - Atomically removes the requested student from the queue (so they can't be
//     served twice and the queue stays consistent), generates a Zoom meeting
//     link, and persists a sessions row.
//   - Broadcasts SESSION_STARTED so the student's UI can pick up the link in
//     real time, plus QUEUE_UPDATED so other clients refresh.
//
// Responses:
//   - 201 with the session JSON on success
//   - 400 invalid body / missing student_id
//   - 401 missing/invalid auth
//   - 403 wrong role or non-owning TA
//   - 404 queue not found, or student is not in this queue
//   - 409 queue is not open
//   - 500 database/zoom error
func StartSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		claims, err := auth.GetClaimsFromRequest(r)
		if err != nil || claims == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization required"})
			return
		}
		if claims.Role != "ta" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only TAs can start sessions"})
			return
		}

		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		var req StartSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if req.StudentID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "student_id is required"})
			return
		}

		ctx := r.Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		// Safe to call after commit: a no-op rollback is what we want on the error paths.
		defer tx.Rollback()

		// Lock the queue row so its status / ownership cannot change mid-operation.
		var taID int
		var status string
		err = tx.QueryRowContext(ctx,
			`SELECT ta_id, status FROM queues WHERE id = $1 FOR UPDATE`,
			queueID,
		).Scan(&taID, &status)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "queue not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		if taID != claims.UserID {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the owning TA can start sessions on this queue"})
			return
		}
		if QueueStatus(status) != QueueStatusOpen {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "queue is not open"})
			return
		}

		// Confirm the student is in the queue and lock that entry.
		var entryID int
		var studentName string
		err = tx.QueryRowContext(ctx, `
			SELECT qe.id, u.username
			FROM queue_entries qe
			JOIN users u ON u.id = qe.student_id
			WHERE qe.queue_id = $1 AND qe.student_id = $2
			FOR UPDATE
		`, queueID, req.StudentID).Scan(&entryID, &studentName)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "student is not in this queue"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		// Provision the Zoom meeting BEFORE inserting the session row so a Zoom
		// failure doesn't leave us with a session row that has no link.
		meeting, err := zoom.DefaultClient.CreateMeeting(ctx, req.Topic)
		if err != nil {
			if errors.Is(err, zoom.ErrNotConfigured) {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "zoom is not configured on the server"})
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to create zoom meeting"})
			return
		}

		// Remove the served student from the queue and renumber positions.
		if _, err := tx.ExecContext(ctx, `DELETE FROM queue_entries WHERE id = $1`, entryID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE queue_entries q SET position = t.rn
			FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY joined_at) AS rn
				FROM queue_entries WHERE queue_id = $1
			) t WHERE q.queue_id = $1 AND q.id = t.id
		`, queueID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		// Persist the session with the Zoom link.
		var sess Session
		sess.QueueID = queueID
		sess.TAID = claims.UserID
		sess.StudentID = req.StudentID
		sess.StudentName = studentName
		sess.ZoomMeetingID = meeting.MeetingID
		sess.ZoomJoinURL = meeting.JoinURL
		sess.ZoomStartURL = meeting.StartURL
		sess.ZoomPasscode = meeting.Passcode

		err = tx.QueryRowContext(ctx, `
			INSERT INTO sessions (queue_id, ta_id, student_id, zoom_meeting_id, zoom_join_url, zoom_start_url, zoom_passcode)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, started_at::text
		`,
			sess.QueueID, sess.TAID, sess.StudentID,
			sess.ZoomMeetingID, sess.ZoomJoinURL, sess.ZoomStartURL, sess.ZoomPasscode,
		).Scan(&sess.ID, &sess.StartedAt)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		if err := tx.Commit(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		// Broadcast: a session started for this student, and the queue changed.
		// student_zoom_join_url is the only Zoom field exposed on the bus so the
		// start_url (host link) doesn't leak to non-TA subscribers.
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventSessionStarted,
			QueueID: queueID,
			Payload: map[string]any{
				"session_id":      sess.ID,
				"student_id":      sess.StudentID,
				"student_name":    sess.StudentName,
				"ta_id":           sess.TAID,
				"zoom_join_url":   sess.ZoomJoinURL,
				"zoom_meeting_id": sess.ZoomMeetingID,
				"zoom_passcode":   sess.ZoomPasscode,
				"started_at":      sess.StartedAt,
			},
		})
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventQueueUpdated,
			QueueID: queueID,
		})
		PublishUpNextForQueue(ctx, db, queueID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(sess)
	}
}

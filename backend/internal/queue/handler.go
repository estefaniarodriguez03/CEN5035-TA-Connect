package queue

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/httperr"
	"backend/internal/sessionlog"

	"github.com/go-chi/chi/v5"
)

// CreateQueueRequest is the JSON body for POST /api/queues.
type CreateQueueRequest struct {
	CourseID int `json:"course_id"`
}

// UpdateQueueStateRequest is the JSON body for PATCH /api/queues/{id}/state
// and POST /api/queues/{id}/status (status: open | paused | closed).
type UpdateQueueStateRequest struct {
	Status string `json:"status"`
}

// UpdateQueueStatusRequest is an alias for backwards compatibility.
type UpdateQueueStatusRequest = UpdateQueueStateRequest

// PostAnnouncementRequest is the JSON body for POST /api/queues/{id}/announcement.
type PostAnnouncementRequest struct {
	Message string `json:"message"`
}

// CreateQueue handles POST /api/queues. TA creates a new queue.
// Auth and role=ta enforced by middleware.
func CreateQueue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		var req CreateQueueRequest
		_ = json.NewDecoder(r.Body).Decode(&req) // optional body

		var id int
		var createdAt string
		err := db.QueryRowContext(r.Context(), `
			INSERT INTO queues (course_id, ta_id, status) VALUES ($1, $2, 'open')
			RETURNING id, created_at::text
		`, req.CourseID, claims.UserID).Scan(&id, &createdAt)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         id,
			"course_id":  req.CourseID,
			"ta_id":      claims.UserID,
			"status":     "open",
			"created_at": createdAt,
		})
	}
}

// Join handles POST /api/queues/{id}/join. Student joins the queue.
// Auth and role=student enforced by middleware.
// Uses a transaction to safely handle simultaneous join requests.
func Join(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		ctx := r.Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer tx.Rollback()

		// Lock the queue row to serialize concurrent joins and verify status atomically.
		var status string
		err = tx.QueryRowContext(ctx, `SELECT status FROM queues WHERE id = $1 FOR UPDATE`, queueID).Scan(&status)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_not_found", "queue not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if QueueStatus(status) != QueueStatusOpen {
			msg := "queue is not open for new entries"
			switch QueueStatus(status) {
			case QueueStatusPaused:
				msg = "queue is paused"
			case QueueStatusClosed:
				msg = "queue is closed"
			}
			httperr.Write(w, http.StatusConflict, "queue_unavailable", msg, nil)
			return
		}

		// Insert entry with next position (safe under the queue row lock).
		var entryID, position int
		var joinedAt string
		err = tx.QueryRowContext(ctx, `
			INSERT INTO queue_entries (queue_id, student_id, position, joined_at)
			SELECT $1, $2, COALESCE(MAX(position), 0) + 1, NOW()
			FROM queue_entries WHERE queue_id = $1
			RETURNING id, position, joined_at::text
		`, queueID, claims.UserID).Scan(&entryID, &position, &joinedAt)
		if err != nil {
			if auth.IsUniqueViolation(err) {
				httperr.Write(w, http.StatusConflict, "already_in_queue", "already in queue", nil)
				return
			}
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		if err := tx.Commit(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventStudentJoined,
			QueueID: queueID,
			Payload: map[string]any{
				"id":         entryID,
				"queue_id":   queueID,
				"position":   position,
				"joined_at":  joinedAt,
				"student_id": claims.UserID,
			},
		})
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventQueueUpdated,
			QueueID: queueID,
		})
		PublishUpNextForQueue(r.Context(), db, queueID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        entryID,
			"queue_id":  queueID,
			"position":  position,
			"joined_at": joinedAt,
		})
	}
}

// Leave handles POST /api/queues/{id}/leave. Student leaves the queue.
// Auth and role=student enforced by middleware.
func Leave(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		res, err := db.ExecContext(r.Context(), `DELETE FROM queue_entries WHERE queue_id = $1 AND student_id = $2`, queueID, claims.UserID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			httperr.Write(w, http.StatusNotFound, "not_in_queue", "not in queue", nil)
			return
		}

		// Renumber positions so they are 1,2,3,...
		_, _ = db.ExecContext(r.Context(), `
			UPDATE queue_entries q SET position = t.rn
			FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY joined_at) AS rn
				FROM queue_entries WHERE queue_id = $1
			) t WHERE q.queue_id = $1 AND q.id = t.id
		`, queueID)

		// Notify subscribers that a student left and the queue was updated.
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventStudentLeft,
			QueueID: queueID,
			Payload: map[string]any{
				"student_id": claims.UserID,
			},
		})
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventQueueUpdated,
			QueueID: queueID,
		})
		PublishUpNextForQueue(r.Context(), db, queueID)

		w.WriteHeader(http.StatusNoContent)
	}
}

// Next handles POST /api/queues/{id}/next.
// It atomically removes the first student in the queue and returns them as "in session".
// Uses row-level locking to avoid double-serving the same student under concurrent requests.
// Auth and role=ta enforced by middleware.
func Next(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		ctx := r.Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer tx.Rollback()

		// Lock the queue row so status/ownership cannot change mid-operation.
		var status string
		var taID int
		var lastServed sql.NullTime
		var prevServing sql.NullInt64
		err = tx.QueryRowContext(ctx,
			`SELECT ta_id, status, last_served_at, serving_student_id FROM queues WHERE id = $1 FOR UPDATE`,
			queueID,
		).Scan(&taID, &status, &lastServed, &prevServing)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_not_found", "queue not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if taID != claims.UserID {
			httperr.Write(w, http.StatusForbidden, "not_queue_owner", "only the owning TA can advance this queue", nil)
			return
		}
		if QueueStatus(status) != QueueStatusOpen {
			httperr.Write(w, http.StatusConflict, "queue_not_open", "queue is not open", nil)
			return
		}

		// Lock and fetch the first queue entry for this queue.
		var e Entry
		var username string
		err = tx.QueryRowContext(ctx, `
			SELECT qe.id, qe.queue_id, qe.student_id, qe.position, qe.joined_at::text, u.username
			FROM queue_entries qe
			JOIN users u ON u.id = qe.student_id
			WHERE qe.queue_id = $1
			ORDER BY qe.position ASC, qe.joined_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`, queueID).Scan(&e.ID, &e.QueueID, &e.StudentID, &e.Position, &e.JoinedAt, &username)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_empty", "queue is empty", map[string]int{"queue_id": queueID})
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		e.Username = username

		// Remove that entry from the queue.
		if _, err := tx.ExecContext(ctx, `DELETE FROM queue_entries WHERE id = $1`, e.ID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		// Renumber remaining entries so positions stay 1,2,3,...
		if _, err := tx.ExecContext(ctx, `
			UPDATE queue_entries q SET position = t.rn
			FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY joined_at) AS rn
				FROM queue_entries WHERE queue_id = $1
			) t WHERE q.queue_id = $1 AND q.id = t.id
		`, queueID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		var haveSessionDuration bool
		var sessionEnd time.Time
		var sessionDurationSec float64
		if lastServed.Valid {
			haveSessionDuration = true
			sessionEnd = time.Now()
			sessionDurationSec = sessionEnd.Sub(lastServed.Time).Seconds()
		}

		// Persist a completed help session: previous serving_student since last_served_at, ended now.
		if lastServed.Valid && prevServing.Valid {
			if err := sessionlog.InsertCompleted(ctx, tx, int(prevServing.Int64), taID, lastServed.Time, sessionEnd, sessionDurationSec); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
		}

		// Record time between consecutive /next calls as a completed help session for rolling average duration
		// (per queue and per owning TA). Track who is currently being helped for the next completion.
		if _, err := tx.ExecContext(ctx, `
			UPDATE queues SET
				last_served_at = NOW(),
				serving_student_id = $2,
				session_sample_count = session_sample_count + CASE WHEN last_served_at IS NULL THEN 0 ELSE 1 END,
				average_session_duration_seconds = CASE
					WHEN last_served_at IS NULL THEN average_session_duration_seconds
					WHEN session_sample_count = 0 THEN EXTRACT(EPOCH FROM (NOW() - last_served_at))
					ELSE (average_session_duration_seconds * session_sample_count + EXTRACT(EPOCH FROM (NOW() - last_served_at))) / (session_sample_count + 1)
				END
			WHERE id = $1
		`, queueID, e.StudentID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		if haveSessionDuration {
			var ulock int
			if err := tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, taID).Scan(&ulock); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE users SET
					session_sample_count = session_sample_count + 1,
					average_session_duration_seconds = CASE
						WHEN session_sample_count = 0 THEN $2
						ELSE (average_session_duration_seconds * session_sample_count + $2) / (session_sample_count + 1)
					END
				WHERE id = $1
			`, taID, sessionDurationSec); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		log.Printf("queue next: queue_id=%d ta_id=%d dequeued_student_id=%d", queueID, claims.UserID, e.StudentID)

		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventStudentServed,
			QueueID: queueID,
			Payload: e,
		})
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventQueueUpdated,
			QueueID: queueID,
		})
		PublishUpNextForQueue(ctx, db, queueID)

		// Mark the returned student as "in session" in the API response.
		writeJSON(w, http.StatusOK, map[string]any{
			"queue_id": queueID,
			"status":   "in_session",
			"student":  e,
		})
	}
}

// UpdateQueueState handles PATCH /api/queues/{id}/state.
// TA owner sets queue status to open, paused, or closed.
func UpdateQueueState(db *sql.DB) http.HandlerFunc {
	return updateQueueState(db, http.MethodPatch)
}

// UpdateStatus handles POST /api/queues/{id}/status (legacy; prefer PATCH .../state).
func UpdateStatus(db *sql.DB) http.HandlerFunc {
	return updateQueueState(db, http.MethodPost)
}

// Auth and role=ta enforced by middleware.
func updateQueueState(db *sql.DB, allowedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		var req UpdateQueueStateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
			return
		}

		newStatus := QueueStatus(req.Status)
		if !newStatus.Valid() {
			httperr.Write(w, http.StatusBadRequest, "invalid_status", "invalid status", nil)
			return
		}

		var taID int
		var previous string
		err = db.QueryRowContext(r.Context(), `SELECT ta_id, status FROM queues WHERE id = $1`, queueID).Scan(&taID, &previous)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_not_found", "queue not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if taID != claims.UserID {
			httperr.Write(w, http.StatusForbidden, "not_queue_owner", "only the owning TA can update this queue", nil)
			return
		}

		fromStatus := QueueStatus(previous)
		if !fromStatus.Valid() {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "invalid queue data", nil)
			return
		}
		if !CanTransitionTo(fromStatus, newStatus) {
			msg := `invalid state transition: from "` + string(fromStatus) + `" to "` + string(newStatus) + `"`
			if fromStatus == QueueStatusClosed {
				msg += `; reopen the queue to "open" first`
			}
			httperr.Write(w, http.StatusConflict, "invalid_state_transition", msg, map[string]any{
				"from":    string(fromStatus),
				"to":      string(newStatus),
				"allowed": AllowedTargetStatuses(fromStatus),
			})
			return
		}

		if previous == string(newStatus) {
			writeJSON(w, http.StatusOK, map[string]any{
				"id":     queueID,
				"status": string(newStatus),
			})
			return
		}

		if _, err := db.ExecContext(r.Context(), `UPDATE queues SET status = $2 WHERE id = $1`, queueID, string(newStatus)); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		if previous != string(newStatus) {
			DefaultHub.Publish(queueID, QueueEvent{
				Type:    EventQueueStateChanged,
				QueueID: queueID,
				Payload: map[string]any{
					"previous_status": previous,
					"status":          string(newStatus),
				},
			})
			DefaultHub.Publish(queueID, QueueEvent{
				Type:    EventQueueUpdated,
				QueueID: queueID,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"id":     queueID,
			"status": string(newStatus),
		})
	}
}

const maxAnnouncementLen = 4000

// PostAnnouncement handles POST /api/queues/{id}/announcement. TA owner posts a message; stored and broadcast.
// Auth and role=ta enforced by middleware.
func PostAnnouncement(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		var req PostAnnouncementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
			return
		}
		msg := strings.TrimSpace(req.Message)
		if msg == "" {
			httperr.Write(w, http.StatusBadRequest, "message_required", "message is required", nil)
			return
		}
		if len(msg) > maxAnnouncementLen {
			httperr.Write(w, http.StatusBadRequest, "message_too_long", "message too long", nil)
			return
		}

		var taID int
		err = db.QueryRowContext(r.Context(), `SELECT ta_id FROM queues WHERE id = $1`, queueID).Scan(&taID)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_not_found", "queue not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if taID != claims.UserID {
			httperr.Write(w, http.StatusForbidden, "not_queue_owner", "only the owning TA can announce on this queue", nil)
			return
		}

		var annID int
		var createdAt string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO queue_announcements (queue_id, ta_id, message)
			VALUES ($1, $2, $3)
			RETURNING id, created_at::text
		`, queueID, claims.UserID, msg).Scan(&annID, &createdAt)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventAnnouncementSent,
			QueueID: queueID,
			Payload: map[string]any{
				"id":         annID,
				"message":    msg,
				"ta_id":      claims.UserID,
				"created_at": createdAt,
			},
		})
		DefaultHub.Publish(queueID, QueueEvent{
			Type:    EventQueueUpdated,
			QueueID: queueID,
		})

		writeJSON(w, http.StatusCreated, map[string]any{
			"id":         annID,
			"queue_id":   queueID,
			"message":    msg,
			"created_at": createdAt,
		})
	}
}

// Entry is one queue entry for API response.
type Entry struct {
	ID                   int    `json:"id"`
	QueueID              int    `json:"queue_id"`
	StudentID            int    `json:"student_id"`
	Position             int    `json:"position"`
	JoinedAt             string `json:"joined_at"`
	Username             string `json:"username,omitempty"`
	EstimatedWaitSeconds int64  `json:"estimated_wait_seconds"`
}

// GetQueue handles GET /api/queues/{id}. Returns queue metadata (public).
func GetQueue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueID, err := parseQueueID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_queue_id", "invalid queue id", nil)
			return
		}

		var id, courseID, taID int
		var status, createdAt string
		var qAvg, taAvg float64
		var qN, taN int
		err = db.QueryRowContext(r.Context(), `
			SELECT q.id, q.course_id, q.ta_id, q.status, q.created_at::text,
				q.average_session_duration_seconds, q.session_sample_count,
				u.average_session_duration_seconds, u.session_sample_count
			FROM queues q
			JOIN users u ON u.id = q.ta_id
			WHERE q.id = $1
		`, queueID).Scan(&id, &courseID, &taID, &status, &createdAt, &qAvg, &qN, &taAvg, &taN)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "queue_not_found", "queue not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		// Load ordered queue entries (students) for this queue.
		rows, err := db.QueryContext(r.Context(), `
			SELECT qe.id, qe.queue_id, qe.student_id, qe.position, qe.joined_at::text, u.username
			FROM queue_entries qe
			JOIN users u ON u.id = qe.student_id
			WHERE qe.queue_id = $1
			ORDER BY qe.position ASC, qe.joined_at ASC
		`, queueID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		var entries []Entry
		for rows.Next() {
			var e Entry
			var username string
			if err := rows.Scan(&e.ID, &e.QueueID, &e.StudentID, &e.Position, &e.JoinedAt, &username); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			e.Username = username
			entries = append(entries, e)
		}
		if entries == nil {
			entries = []Entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(queueStateJSON(id, courseID, taID, status, createdAt, entries, qAvg, qN, taAvg, taN))
	}
}

// GetActiveQueueByCourse handles GET /api/queues/active?course_id={id}.
// Returns the latest open queue for a course, including entries (public).
func GetActiveQueueByCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID, err := parseCourseIDFromQuery(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course id", nil)
			return
		}

		var id, qCourseID, taID int
		var status, createdAt string
		var qAvg, taAvg float64
		var qN, taN int
		err = db.QueryRowContext(r.Context(), `
			SELECT q.id, q.course_id, q.ta_id, q.status, q.created_at::text,
				q.average_session_duration_seconds, q.session_sample_count,
				u.average_session_duration_seconds, u.session_sample_count
			FROM queues q
			JOIN users u ON u.id = q.ta_id
			WHERE q.course_id = $1 AND q.status = 'open'
			ORDER BY q.created_at DESC
			LIMIT 1
		`, courseID).Scan(&id, &qCourseID, &taID, &status, &createdAt, &qAvg, &qN, &taAvg, &taN)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "no_active_queue", "no active queue for course", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT qe.id, qe.queue_id, qe.student_id, qe.position, qe.joined_at::text, u.username
			FROM queue_entries qe
			JOIN users u ON u.id = qe.student_id
			WHERE qe.queue_id = $1
			ORDER BY qe.position ASC, qe.joined_at ASC
		`, id)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		var entries []Entry
		for rows.Next() {
			var e Entry
			var username string
			if err := rows.Scan(&e.ID, &e.QueueID, &e.StudentID, &e.Position, &e.JoinedAt, &username); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			e.Username = username
			entries = append(entries, e)
		}
		if entries == nil {
			entries = []Entry{}
		}

		writeJSON(w, http.StatusOK, queueStateJSON(id, qCourseID, taID, status, createdAt, entries, qAvg, qN, taAvg, taN))
	}
}

func parseQueueID(r *http.Request) (int, error) {
	s := chi.URLParam(r, "id")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(s)
}

func parseCourseIDFromQuery(r *http.Request) (int, error) {
	s := r.URL.Query().Get("course_id")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(s)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

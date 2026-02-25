package queue

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/auth"

	"github.com/go-chi/chi/v5"
)

// CreateQueueRequest is the JSON body for POST /api/queues.
type CreateQueueRequest struct {
	CourseID int `json:"course_id"`
}

// CreateQueue handles POST /api/queues. TA creates a new queue; requires auth and role=ta.
func CreateQueue(db *sql.DB) http.HandlerFunc {
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
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only TAs can create queues"})
			return
		}
		var req CreateQueueRequest
		_ = json.NewDecoder(r.Body).Decode(&req) // optional body

		var id int
		var createdAt string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO queues (course_id, ta_id, status) VALUES ($1, $2, 'open')
			RETURNING id, created_at::text
		`, req.CourseID, claims.UserID).Scan(&id, &createdAt)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
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

// Join handles POST /api/queues/{id}/join. Student joins the queue; requires auth.
func Join(db *sql.DB) http.HandlerFunc {
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
		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		// Check queue exists and is open
		var status string
		err = db.QueryRowContext(r.Context(), `SELECT status FROM queues WHERE id = $1`, queueID).Scan(&status)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "queue not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		if status != "open" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "queue is not open for new entries"})
			return
		}

		// Insert entry with next position (atomic)
		var entryID, position int
		var joinedAt string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO queue_entries (queue_id, student_id, position, joined_at)
			SELECT $1, $2, COALESCE(MAX(position), 0) + 1, NOW()
			FROM queue_entries WHERE queue_id = $1
			RETURNING id, position, joined_at::text
		`, queueID, claims.UserID).Scan(&entryID, &position, &joinedAt)
		if err != nil {
			if auth.IsUniqueViolation(err) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "already in queue"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

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

// Leave handles POST /api/queues/{id}/leave. Student leaves the queue; requires auth.
func Leave(db *sql.DB) http.HandlerFunc {
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
		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		res, err := db.ExecContext(r.Context(), `DELETE FROM queue_entries WHERE queue_id = $1 AND student_id = $2`, queueID, claims.UserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not in queue"})
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

		w.WriteHeader(http.StatusNoContent)
	}
}

// Next handles POST /api/queues/{id}/next.
// It atomically removes the first student in the queue and returns them as "in session".
// Uses row-level locking to avoid double-serving the same student under concurrent requests.
func Next(db *sql.DB) http.HandlerFunc {
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
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only TAs can advance the queue"})
			return
		}

		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		ctx := r.Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		defer tx.Rollback()

		// Lock the queue row so status/ownership cannot change mid-operation.
		var status string
		var taID int
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
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the owning TA can advance this queue"})
			return
		}
		if status != "open" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "queue is not open"})
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
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "queue is empty"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		e.Username = username

		// Remove that entry from the queue.
		if _, err := tx.ExecContext(ctx, `DELETE FROM queue_entries WHERE id = $1`, e.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
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
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		if err := tx.Commit(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		// Mark the returned student as "in session" in the API response.
		writeJSON(w, http.StatusOK, map[string]any{
			"queue_id": queueID,
			"status":   "in_session",
			"student":  e,
		})
	}
}

// Entry is one queue entry for API response.
type Entry struct {
	ID        int    `json:"id"`
	QueueID   int    `json:"queue_id"`
	StudentID int    `json:"student_id"`
	Position  int    `json:"position"`
	JoinedAt  string `json:"joined_at"`
	Username  string `json:"username,omitempty"`
}

// GetQueue handles GET /api/queues/{id}. Returns queue metadata.
func GetQueue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		var id, courseID, taID int
		var status, createdAt string
		err = db.QueryRowContext(r.Context(),
			`SELECT id, course_id, ta_id, status, created_at::text FROM queues WHERE id = $1`,
			queueID).Scan(&id, &courseID, &taID, &status, &createdAt)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "queue not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
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
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		defer rows.Close()

		var entries []Entry
		for rows.Next() {
			var e Entry
			var username string
			if err := rows.Scan(&e.ID, &e.QueueID, &e.StudentID, &e.Position, &e.JoinedAt, &username); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         id,
			"course_id":  courseID,
			"ta_id":      taID,
			"status":     status,
			"created_at": createdAt,
			"entries":    entries,
		})
	}
}

func parseQueueID(r *http.Request) (int, error) {
	s := chi.URLParam(r, "id")
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

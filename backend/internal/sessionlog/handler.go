package sessionlog

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/auth"
	"backend/internal/httperr"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 200
)

// SessionHistoryEntry is one row for GET /api/session-history (includes display names).
type SessionHistoryEntry struct {
	ID              int     `json:"id"`
	StudentID       int     `json:"student_id"`
	TAID            int     `json:"ta_id"`
	StudentUsername string  `json:"student_username"`
	TAUsername      string  `json:"ta_username"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	DurationSeconds float64 `json:"duration"`
}

// ListHistory handles GET /api/session-history.
// TAs see sessions they taught; students see sessions they attended.
// Query: limit (default 50, max 200), offset (default 0).
func ListHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "auth_required", "authorization required", nil)
			return
		}
		switch claims.Role {
		case "ta", "student":
		default:
			httperr.Write(w, http.StatusForbidden, "role_forbidden", "session history is only available for ta and student accounts", nil)
			return
		}

		limit := defaultHistoryLimit
		if s := r.URL.Query().Get("limit"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				limit = v
			}
		}
		if limit > maxHistoryLimit {
			limit = maxHistoryLimit
		}
		offset := 0
		if s := r.URL.Query().Get("offset"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v >= 0 {
				offset = v
			}
		}

		var rows *sql.Rows
		var err error
		if claims.Role == "ta" {
			rows, err = db.QueryContext(r.Context(), `
				SELECT sl.id, sl.student_id, sl.ta_id,
					stu.username, ta_u.username,
					sl.start_time::text, sl.end_time::text, sl.duration_seconds
				FROM session_logs sl
				JOIN users stu ON stu.id = sl.student_id
				JOIN users ta_u ON ta_u.id = sl.ta_id
				WHERE sl.ta_id = $1
				ORDER BY sl.start_time DESC
				LIMIT $2 OFFSET $3
			`, claims.UserID, limit, offset)
		} else {
			rows, err = db.QueryContext(r.Context(), `
				SELECT sl.id, sl.student_id, sl.ta_id,
					stu.username, ta_u.username,
					sl.start_time::text, sl.end_time::text, sl.duration_seconds
				FROM session_logs sl
				JOIN users stu ON stu.id = sl.student_id
				JOIN users ta_u ON ta_u.id = sl.ta_id
				WHERE sl.student_id = $1
				ORDER BY sl.start_time DESC
				LIMIT $2 OFFSET $3
			`, claims.UserID, limit, offset)
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		var out []SessionHistoryEntry
		for rows.Next() {
			var e SessionHistoryEntry
			if err := rows.Scan(
				&e.ID, &e.StudentID, &e.TAID,
				&e.StudentUsername, &e.TAUsername,
				&e.StartTime, &e.EndTime, &e.DurationSeconds,
			); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			out = append(out, e)
		}
		if err := rows.Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if out == nil {
			out = []SessionHistoryEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sessions": out,
		})
	}
}

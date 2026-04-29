package course

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/auth"
	"backend/internal/httperr"
	"github.com/go-chi/chi/v5"
)

type AddTACourseRequest struct {
	CourseID int    `json:"course_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
}

type Course struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// AddForTA handles POST /api/ta/courses.
// TA-only route: links the authenticated TA to a course.
// Accepts either course_id or code (+ optional name).
func AddForTA(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		var req AddTACourseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
			return
		}

		code := strings.ToUpper(strings.TrimSpace(req.Code))
		name := strings.TrimSpace(req.Name)

		if req.CourseID <= 0 && code == "" {
			httperr.Write(w, http.StatusBadRequest, "missing_course_reference", "provide course_id or code", nil)
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer tx.Rollback()

		courseID := req.CourseID
		if courseID > 0 {
			var exists int
			if err := tx.QueryRowContext(r.Context(), `SELECT 1 FROM courses WHERE id = $1`, courseID).Scan(&exists); err == sql.ErrNoRows {
				httperr.Write(w, http.StatusNotFound, "course_not_found", "course not found", nil)
				return
			} else if err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
		} else {
			if name == "" {
				name = code
			}
			if err := tx.QueryRowContext(r.Context(), `
				INSERT INTO courses (code, name)
				VALUES ($1, $2)
				ON CONFLICT (code) DO UPDATE SET name = CASE WHEN courses.name = '' THEN EXCLUDED.name ELSE courses.name END
				RETURNING id
			`, code, name).Scan(&courseID); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
		}

		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO ta_courses (ta_id, course_id)
			VALUES ($1, $2)
			ON CONFLICT (ta_id, course_id) DO NOTHING
		`, claims.UserID, courseID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		var course Course
		if err := tx.QueryRowContext(r.Context(), `
			SELECT id, code, name FROM courses WHERE id = $1
		`, courseID).Scan(&course.ID, &course.Code, &course.Name); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		if err := tx.Commit(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"ta_id":  claims.UserID,
			"course": course,
		})
	}
}

// ListForTA handles GET /api/ta/courses.
// TA-only route: returns the authenticated TA's linked courses.
func ListForTA(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		rows, err := db.QueryContext(r.Context(), `
			SELECT c.id, c.code, c.name
			FROM ta_courses tc
			JOIN courses c ON c.id = tc.course_id
			WHERE tc.ta_id = $1
			ORDER BY c.code ASC
		`, claims.UserID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		out := make([]Course, 0)
		for rows.Next() {
			var c Course
			if err := rows.Scan(&c.ID, &c.Code, &c.Name); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			out = append(out, c)
		}
		if err := rows.Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"courses": out})
	}
}

// DeleteForTA handles DELETE /api/ta/courses/{id}.
// TA-only route: removes the authenticated TA's link to a course.
func DeleteForTA(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		courseIDStr := chi.URLParam(r, "id")
		courseID, err := strconv.Atoi(courseIDStr)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course ID", nil)
			return
		}

		result, err := db.ExecContext(r.Context(), `
			DELETE FROM ta_courses
			WHERE ta_id = $1 AND course_id = $2
		`, claims.UserID, courseID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		if rowsAffected == 0 {
			httperr.Write(w, http.StatusNotFound, "not_found", "course not linked to this TA", nil)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

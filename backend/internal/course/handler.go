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
    Color    string `json:"color"`
}

type Course struct {
    ID    int    `json:"id"`
    Code  string `json:"code"`
    Name  string `json:"name"`
    Color string `json:"color"`
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
        color := strings.ToLower(strings.TrimSpace(req.Color))
        if color == "" {
            color = "orange"
        }

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
                INSERT INTO courses (code, name, color)
                VALUES ($1, $2, $3)
                ON CONFLICT (code) DO UPDATE SET name = CASE WHEN courses.name = '' THEN EXCLUDED.name ELSE courses.name END, color = CASE WHEN courses.color = 'orange' THEN EXCLUDED.color ELSE courses.color END
                RETURNING id
            `, code, name, color).Scan(&courseID); err != nil {
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
            SELECT id, code, name, COALESCE(color, 'orange') FROM courses WHERE id = $1
        `, courseID).Scan(&course.ID, &course.Code, &course.Name, &course.Color); err != nil {
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
            SELECT c.id, c.code, c.name, COALESCE(c.color, 'orange')
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
            if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Color); err != nil {
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
// If no other TAs have this course, deletes it from the database entirely.
func DeleteForTA(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := auth.ClaimsFromContext(r.Context())

        courseIDStr := chi.URLParam(r, "id")
        courseID, err := strconv.Atoi(courseIDStr)
        if err != nil {
            httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course ID", nil)
            return
        }

        // Start transaction to ensure atomic operations
        tx, err := db.BeginTx(r.Context(), nil)
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }
        defer tx.Rollback()

        // Delete the TA-course link
        result, err := tx.ExecContext(r.Context(), `
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

        // Check if any other TAs have this course
        var otherTACount int
        if err := tx.QueryRowContext(r.Context(), `
            SELECT COUNT(*) FROM ta_courses WHERE course_id = $1
        `, courseID).Scan(&otherTACount); err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }

        // If no other TAs have this course, delete it from the database
        if otherTACount == 0 {
            // Delete student office hours for this course
            if _, err := tx.ExecContext(r.Context(), `
                DELETE FROM student_office_hours
                WHERE office_hour_id IN (
                    SELECT id FROM office_hours WHERE course_id = $1
                )
            `, courseID); err != nil {
                httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
                return
            }

            // Delete office hours for this course
            if _, err := tx.ExecContext(r.Context(), `
                DELETE FROM office_hours WHERE course_id = $1
            `, courseID); err != nil {
                httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
                return
            }

            // Delete student courses for this course
            if _, err := tx.ExecContext(r.Context(), `
                DELETE FROM student_courses WHERE course_id = $1
            `, courseID); err != nil {
                httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
                return
            }

            // Delete the course itself
            if _, err := tx.ExecContext(r.Context(), `
                DELETE FROM courses WHERE id = $1
            `, courseID); err != nil {
                httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
                return
            }
        }

        if err := tx.Commit(); err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

// ListAll handles GET /api/courses. Public route.
// Returns only courses that have at least one TA assigned.
func ListAll(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.QueryContext(r.Context(), `
            SELECT DISTINCT c.id, c.code, c.name, COALESCE(c.color, 'orange')
            FROM courses c
            JOIN ta_courses tc ON tc.course_id = c.id
            ORDER BY c.code ASC
        `)
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }
        defer rows.Close()

        out := make([]Course, 0)
        for rows.Next() {
            var c Course
            if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Color); err != nil {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}
package studentcourses

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/auth"
	"backend/internal/httperr"

	"github.com/go-chi/chi/v5"
)

type StudentCourse struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// List handles GET /api/student/courses
func List(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "not authenticated", nil)
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT c.id, c.code, c.name, COALESCE(c.color, 'orange')
			FROM student_courses sc
			JOIN courses c ON c.id = sc.course_id
			WHERE sc.student_id = $1
			ORDER BY c.code ASC
		`, claims.UserID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}
		defer rows.Close()

		courses := []StudentCourse{}
		for rows.Next() {
			var course StudentCourse
			if err := rows.Scan(&course.ID, &course.Code, &course.Name, &course.Color); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
				return
			}
			courses = append(courses, course)
		}

		if err := rows.Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"courses": courses})
	}
}

// Add handles POST /api/student/courses
func Add(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "not authenticated", nil)
			return
		}

		var req struct {
			CourseID int `json:"course_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CourseID <= 0 {
			httperr.Write(w, http.StatusBadRequest, "invalid_request", "course_id is required", nil)
			return
		}

		// Verify course exists
		var courseCode, courseName, courseColor string
		err := db.QueryRowContext(r.Context(),
			"SELECT code, name, COALESCE(color, 'orange') FROM courses WHERE id = $1", req.CourseID).Scan(&courseCode, &courseName, &courseColor)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "not_found", "course not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		// Add course to student (ignore duplicate)
		_, err = db.ExecContext(r.Context(), `
			INSERT INTO student_courses (student_id, course_id)
			VALUES ($1, $2)
			ON CONFLICT (student_id, course_id) DO NOTHING
		`, claims.UserID, req.CourseID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"course": StudentCourse{
				ID:    req.CourseID,
				Code:  courseCode,
				Name:  courseName,
				Color: courseColor,
			},
		})
	}
}

// Remove handles DELETE /api/student/courses/{id}
func Remove(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "not authenticated", nil)
			return
		}

		courseIDStr := chi.URLParam(r, "id")
		courseID, err := strconv.Atoi(courseIDStr)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course ID", nil)
			return
		}

		result, err := db.ExecContext(r.Context(), `
			DELETE FROM student_courses
			WHERE student_id = $1 AND course_id = $2
		`, claims.UserID, courseID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		if rowsAffected == 0 {
			httperr.Write(w, http.StatusNotFound, "not_found", "student course not found", nil)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateBatch handles PUT /api/student/courses/batch
func UpdateBatch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "not authenticated", nil)
			return
		}

		var req struct {
			CourseIDs []int `json:"course_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request", "invalid request body", nil)
			return
		}

		// Start transaction
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}
		defer tx.Rollback()

		// Delete all existing student courses
		if _, err := tx.ExecContext(r.Context(),
			"DELETE FROM student_courses WHERE student_id = $1", claims.UserID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		// Insert new courses
		for _, courseID := range req.CourseIDs {
			if _, err := tx.ExecContext(r.Context(), `
				INSERT INTO student_courses (student_id, course_id)
				VALUES ($1, $2)
			`, claims.UserID, courseID); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}

		// Fetch and return updated courses
		rows, err := db.QueryContext(r.Context(), `
			SELECT c.id, c.code, c.name, COALESCE(c.color, 'orange')
			FROM student_courses sc
			JOIN courses c ON c.id = sc.course_id
			WHERE sc.student_id = $1
			ORDER BY c.code ASC
		`, claims.UserID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
			return
		}
		defer rows.Close()

		courses := []StudentCourse{}
		for rows.Next() {
			var course StudentCourse
			if err := rows.Scan(&course.ID, &course.Code, &course.Name, &course.Color); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "db_error", "database error", nil)
				return
			}
			courses = append(courses, course)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"courses": courses})
	}
}

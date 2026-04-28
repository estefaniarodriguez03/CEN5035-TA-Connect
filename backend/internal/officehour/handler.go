package officehour

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/httperr"

	"github.com/go-chi/chi/v5"
)

// CreateOfficeHourRequest is the JSON body for POST /api/office-hours.
type CreateOfficeHourRequest struct {
	CourseID  int    `json:"course_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Location  string `json:"location"`
}

func ensureCourseExists(ctx context.Context, db *sql.DB, courseID int) error {
	code := "AUTO-" + strconv.Itoa(courseID)
	name := "Auto-created course " + strconv.Itoa(courseID)
	_, err := db.ExecContext(ctx, `
		INSERT INTO courses (id, code, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, courseID, code, name)
	return err
}

// Create handles POST /api/office-hours. TA-only; rejects overlapping slots for the same TA and day.
// Auth and role=ta enforced by middleware.
func Create(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		var req CreateOfficeHourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
			return
		}

		if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
			httperr.Write(w, http.StatusBadRequest, "invalid_day_of_week", "day_of_week must be 0-6 (Sunday-Saturday)", nil)
			return
		}

		startClock, err := parseClock(req.StartTime)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_start_time", "invalid start_time", nil)
			return
		}
		endClock, err := parseClock(req.EndTime)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_end_time", "invalid end_time", nil)
			return
		}
		if !startClock.Before(endClock) {
			httperr.Write(w, http.StatusBadRequest, "invalid_time_range", "start_time must be before end_time", nil)
			return
		}
		if err := ensureCourseExists(r.Context(), db, req.CourseID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		var overlaps bool
		err = db.QueryRowContext(r.Context(), `
			SELECT EXISTS (
				SELECT 1 FROM office_hours
				WHERE ta_id = $1 AND day_of_week = $2
				AND start_time < $4::time AND end_time > $3::time
			)
		`, claims.UserID, req.DayOfWeek, startClock.Format("15:04:05"), endClock.Format("15:04:05")).Scan(&overlaps)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if overlaps {
			httperr.Write(w, http.StatusConflict, "office_hour_overlap", "overlapping office hour", nil)
			return
		}

		var id int
		var startOut, endOut string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO office_hours (ta_id, course_id, day_of_week, start_time, end_time, location)
			VALUES ($1, $2, $3, $4::time, $5::time, $6)
			RETURNING id, start_time::text, end_time::text
		`, claims.UserID, req.CourseID, req.DayOfWeek, startClock.Format("15:04:05"), endClock.Format("15:04:05"), strings.TrimSpace(req.Location)).Scan(&id, &startOut, &endOut)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(OfficeHour{
			ID:        id,
			TAID:      claims.UserID,
			CourseID:  req.CourseID,
			DayOfWeek: req.DayOfWeek,
			StartTime: trimTimeSuffix(startOut),
			EndTime:   trimTimeSuffix(endOut),
			Location:  strings.TrimSpace(req.Location),
		})
	}
}

// Update handles PUT /api/office-hours/{id}. TA-only; owning TA only; rejects overlaps with other slots on the same day.
// Auth and role=ta enforced by middleware.
func Update(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		id, err := parseOfficeHourID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_office_hour_id", "invalid office hour id", nil)
			return
		}

		var ownerID int
		err = db.QueryRowContext(r.Context(), `SELECT ta_id FROM office_hours WHERE id = $1`, id).Scan(&ownerID)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "office_hour_not_found", "office hour not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if ownerID != claims.UserID {
			httperr.Write(w, http.StatusForbidden, "not_resource_owner", "only the owning TA can update this office hour", nil)
			return
		}

		var req CreateOfficeHourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
			return
		}

		if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
			httperr.Write(w, http.StatusBadRequest, "invalid_day_of_week", "day_of_week must be 0-6 (Sunday-Saturday)", nil)
			return
		}

		startClock, err := parseClock(req.StartTime)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_start_time", "invalid start_time", nil)
			return
		}
		endClock, err := parseClock(req.EndTime)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_end_time", "invalid end_time", nil)
			return
		}
		if !startClock.Before(endClock) {
			httperr.Write(w, http.StatusBadRequest, "invalid_time_range", "start_time must be before end_time", nil)
			return
		}

		startStr := startClock.Format("15:04:05")
		endStr := endClock.Format("15:04:05")
		if err := ensureCourseExists(r.Context(), db, req.CourseID); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		var overlaps bool
		err = db.QueryRowContext(r.Context(), `
			SELECT EXISTS (
				SELECT 1 FROM office_hours
				WHERE ta_id = $1 AND day_of_week = $2 AND id <> $5
				AND start_time < $4::time AND end_time > $3::time
			)
		`, claims.UserID, req.DayOfWeek, startStr, endStr, id).Scan(&overlaps)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if overlaps {
			httperr.Write(w, http.StatusConflict, "office_hour_overlap", "overlapping office hour", nil)
			return
		}

		var startOut, endOut string
		err = db.QueryRowContext(r.Context(), `
			UPDATE office_hours
			SET course_id = $2, day_of_week = $3, start_time = $4::time, end_time = $5::time, location = $6
			WHERE id = $1
			RETURNING start_time::text, end_time::text
		`, id, req.CourseID, req.DayOfWeek, startStr, endStr, strings.TrimSpace(req.Location)).Scan(&startOut, &endOut)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "office_hour_not_found", "office hour not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		writeJSON(w, http.StatusOK, OfficeHour{
			ID:        id,
			TAID:      claims.UserID,
			CourseID:  req.CourseID,
			DayOfWeek: req.DayOfWeek,
			StartTime: trimTimeSuffix(startOut),
			EndTime:   trimTimeSuffix(endOut),
			Location:  strings.TrimSpace(req.Location),
		})
	}
}

// Delete handles DELETE /api/office-hours/{id}. TA-only; owning TA only.
// Auth and role=ta enforced by middleware.
func Delete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())

		id, err := parseOfficeHourID(r)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_office_hour_id", "invalid office hour id", nil)
			return
		}

		var ownerID int
		err = db.QueryRowContext(r.Context(), `SELECT ta_id FROM office_hours WHERE id = $1`, id).Scan(&ownerID)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusNotFound, "office_hour_not_found", "office hour not found", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if ownerID != claims.UserID {
			httperr.Write(w, http.StatusForbidden, "not_resource_owner", "only the owning TA can delete this office hour", nil)
			return
		}

		res, err := db.ExecContext(r.Context(), `DELETE FROM office_hours WHERE id = $1`, id)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		n, err := res.RowsAffected()
		if err != nil || n == 0 {
			httperr.Write(w, http.StatusNotFound, "office_hour_not_found", "office hour not found", nil)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ListByTA handles GET /api/office-hours/ta/{ta_id}. Public; returns all office hours for that TA, ordered by day then start time.
func ListByTA(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "ta_id")
		if s == "" {
			httperr.Write(w, http.StatusBadRequest, "invalid_ta_id", "invalid ta id", nil)
			return
		}
		taID, err := strconv.Atoi(s)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_ta_id", "invalid ta id", nil)
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, ta_id, course_id, day_of_week, start_time::text, end_time::text, location
			FROM office_hours
			WHERE ta_id = $1
			ORDER BY day_of_week ASC, start_time ASC
		`, taID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		list := make([]OfficeHour, 0)
		for rows.Next() {
			var oh OfficeHour
			var startOut, endOut string
			if err := rows.Scan(&oh.ID, &oh.TAID, &oh.CourseID, &oh.DayOfWeek, &startOut, &endOut, &oh.Location); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			oh.StartTime = trimTimeSuffix(startOut)
			oh.EndTime = trimTimeSuffix(endOut)
			list = append(list, oh)
		}
		if err := rows.Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(list)
	}
}

// ListByCourse handles GET /api/office-hours/course/{course_id}. Public; returns all office hours for that course, ordered by day, start time, then TA.
func ListByCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "course_id")
		if s == "" {
			httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course id", nil)
			return
		}
		courseID, err := strconv.Atoi(s)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_course_id", "invalid course id", nil)
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, ta_id, course_id, day_of_week, start_time::text, end_time::text, location
			FROM office_hours
			WHERE course_id = $1
			ORDER BY day_of_week ASC, start_time ASC, ta_id ASC
		`, courseID)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		defer rows.Close()

		list := make([]OfficeHour, 0)
		for rows.Next() {
			var oh OfficeHour
			var startOut, endOut string
			if err := rows.Scan(&oh.ID, &oh.TAID, &oh.CourseID, &oh.DayOfWeek, &startOut, &endOut, &oh.Location); err != nil {
				httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
				return
			}
			oh.StartTime = trimTimeSuffix(startOut)
			oh.EndTime = trimTimeSuffix(endOut)
			list = append(list, oh)
		}
		if err := rows.Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(list)
	}
}

func parseOfficeHourID(r *http.Request) (int, error) {
	s := chi.URLParam(r, "id")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(s)
}

// parseClock parses a wall-clock time (HH:MM or HH:MM:SS).
func parseClock(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	ref := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, layout := range []string{"15:04:05", "15:04"} {
		t, err := time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			return time.Date(ref.Year(), ref.Month(), ref.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
		}
	}
	return time.Time{}, errors.New("invalid time")
}

// trimTimeSuffix drops fractional seconds if Postgres returns them (e.g. 11:00:00.123456).
func trimTimeSuffix(s string) string {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return s[:i]
	}
	return s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

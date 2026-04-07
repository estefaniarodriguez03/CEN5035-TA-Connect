package officehour

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
)

// CreateOfficeHourRequest is the JSON body for POST /api/office-hours.
type CreateOfficeHourRequest struct {
	CourseID  int    `json:"course_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Location  string `json:"location"`
}

// Create handles POST /api/office-hours. TA-only; rejects overlapping slots for the same TA and day.
func Create(db *sql.DB) http.HandlerFunc {
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
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only TAs can create office hours"})
			return
		}

		var req CreateOfficeHourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "day_of_week must be 0-6 (Sunday-Saturday)"})
			return
		}

		startClock, err := parseClock(req.StartTime)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start_time"})
			return
		}
		endClock, err := parseClock(req.EndTime)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid end_time"})
			return
		}
		if !startClock.Before(endClock) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start_time must be before end_time"})
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
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		if overlaps {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "overlapping office hour"})
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
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
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

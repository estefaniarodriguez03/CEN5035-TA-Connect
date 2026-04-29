package studentschedule

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"

    "backend/internal/auth"
    "backend/internal/httperr"

    "github.com/go-chi/chi/v5"
)

type ScheduleEntry struct {
    ID            int    `json:"id"`
    OfficeHourID  int    `json:"office_hour_id"`
    CourseID      int    `json:"course_id"`
    CourseCode    string `json:"course_code"`
    CourseName    string `json:"course_name"`
    CourseColor   string `json:"course_color"`
    TAUsername    string `json:"ta_username"`
    TAID          int    `json:"ta_id"`
    DayOfWeek     int    `json:"day_of_week"`
    StartTime     string `json:"start_time"`
    EndTime       string `json:"end_time"`
    Location      string `json:"location"`
}

// Add handles POST /api/student/schedule
func Add(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := auth.ClaimsFromContext(r.Context())

        var req struct {
            OfficeHourID int `json:"office_hour_id"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OfficeHourID <= 0 {
            httperr.Write(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
            return
        }

        // Verify the office hour exists
        var exists int
        err := db.QueryRowContext(r.Context(),
            `SELECT 1 FROM office_hours WHERE id = $1`, req.OfficeHourID).Scan(&exists)
        if err == sql.ErrNoRows {
            httperr.Write(w, http.StatusNotFound, "office_hour_not_found", "office hour not found", nil)
            return
        }
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }

        var id int
        err = db.QueryRowContext(r.Context(), `
            INSERT INTO student_office_hours (student_id, office_hour_id)
            VALUES ($1, $2)
            ON CONFLICT (student_id, office_hour_id) DO NOTHING
            RETURNING id
        `, claims.UserID, req.OfficeHourID).Scan(&id)
        if err == sql.ErrNoRows {
            // Already exists — not an error, just return 200
            w.WriteHeader(http.StatusOK)
            return
        }
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }

        writeJSON(w, http.StatusCreated, map[string]any{"id": id})
    }
}

// List handles GET /api/student/schedule
func List(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := auth.ClaimsFromContext(r.Context())

        rows, err := db.QueryContext(r.Context(), `
            SELECT soh.id, oh.id, oh.course_id, c.code, c.name, COALESCE(c.color, 'orange'),
                   u.username, oh.ta_id, oh.day_of_week,
                   oh.start_time::text, oh.end_time::text, oh.location
            FROM student_office_hours soh
            JOIN office_hours oh ON oh.id = soh.office_hour_id
            JOIN courses c ON c.id = oh.course_id
            JOIN users u ON u.id = oh.ta_id
            WHERE soh.student_id = $1
            ORDER BY oh.day_of_week ASC, oh.start_time ASC
        `, claims.UserID)
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }
        defer rows.Close()

        out := make([]ScheduleEntry, 0)
        for rows.Next() {
            var e ScheduleEntry
            var startOut, endOut string
            if err := rows.Scan(
                &e.ID, &e.OfficeHourID, &e.CourseID, &e.CourseCode, &e.CourseName, &e.CourseColor,
                &e.TAUsername, &e.TAID, &e.DayOfWeek,
                &startOut, &endOut, &e.Location,
            ); err != nil {
                httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
                return
            }
            e.StartTime = trimTimeSuffix(startOut)
            e.EndTime = trimTimeSuffix(endOut)
            out = append(out, e)
        }
        if err := rows.Err(); err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }

        writeJSON(w, http.StatusOK, map[string]any{"schedule": out})
    }
}

// Remove handles DELETE /api/student/schedule/{id}
func Remove(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := auth.ClaimsFromContext(r.Context())

        s := chi.URLParam(r, "id")
        id, err := strconv.Atoi(s)
        if err != nil || id <= 0 {
            httperr.Write(w, http.StatusBadRequest, "invalid_id", "invalid id", nil)
            return
        }

        res, err := db.ExecContext(r.Context(),
            `DELETE FROM student_office_hours WHERE id = $1 AND student_id = $2`,
            id, claims.UserID)
        if err != nil {
            httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
            return
        }
        n, _ := res.RowsAffected()
        if n == 0 {
            httperr.Write(w, http.StatusNotFound, "not_found", "schedule entry not found", nil)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}

func trimTimeSuffix(s string) string {
    for i, c := range s {
        if c == '.' {
            return s[:i]
        }
    }
    return s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}
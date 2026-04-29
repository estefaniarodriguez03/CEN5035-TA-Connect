package officehour

// OfficeHour is a recurring weekly office hour slot stored in office_hours.
// DayOfWeek follows Go's time.Weekday: 0 = Sunday, 6 = Saturday.
// StartTime and EndTime are wall-clock times in "HH:MM:SS" form as returned by PostgreSQL TIME.
type OfficeHour struct {
    ID         int    `json:"id"`
    TAID       int    `json:"ta_id"`
    TAUsername string `json:"ta_username,omitempty"`
    CourseID   int    `json:"course_id"`
    DayOfWeek  int    `json:"day_of_week"`
    StartTime  string `json:"start_time"`
    EndTime    string `json:"end_time"`
    Location   string `json:"location"`
}
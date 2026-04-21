package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TA can create an office hour; response echoes all fields (time values include seconds).
func TestCreateOfficeHour_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_oh_%d", suffix),
		"email":    fmt.Sprintf("ta_oh_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	body := map[string]any{
		"course_id":   2,
		"day_of_week": 1,
		"start_time":  "11:00",
		"end_time":    "13:00",
		"location":    "CSE 220",
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", body, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create office hour: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var oh struct {
		ID        int    `json:"id"`
		TAID      int    `json:"ta_id"`
		CourseID  int    `json:"course_id"`
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Location  string `json:"location"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&oh); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if oh.ID == 0 || oh.TAID != ta.User.ID || oh.CourseID != 2 || oh.DayOfWeek != 1 {
		t.Fatalf("unexpected office hour: %+v", oh)
	}
	if oh.StartTime != "11:00:00" || oh.EndTime != "13:00:00" || oh.Location != "CSE 220" {
		t.Fatalf("unexpected times/location: %+v", oh)
	}
}

// /api/office-hours POST requires a TA token: no token => 401, student => 403.
func TestCreateOfficeHour_UnauthorizedAndStudentForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "10:00", "end_time": "11:00",
	}, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_oh_%d", suffix),
		"email":    fmt.Sprintf("st_oh_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "10:00", "end_time": "11:00",
	}, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// Overlapping time windows on the same day for the same TA are rejected as 409.
// Adjacent windows (touching boundaries) are allowed.
func TestCreateOfficeHour_OverlapRejected(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_oh2_%d", suffix),
		"email":    fmt.Sprintf("ta_oh2_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	slot := map[string]any{
		"course_id": 1, "day_of_week": 3, "start_time": "10:00", "end_time": "12:00", "location": "A",
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", slot, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first slot: expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	overlap := map[string]any{
		"course_id": 2, "day_of_week": 3, "start_time": "11:00", "end_time": "13:00", "location": "B",
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", overlap, ta.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("overlap: expected 409 got %d body=%s", rr.Code, rr.Body.String())
	}

	adjacent := map[string]any{
		"course_id": 1, "day_of_week": 3, "start_time": "12:00", "end_time": "14:00", "location": "C",
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", adjacent, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("adjacent slot: expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// Owning TA can fully update their office hour.
func TestUpdateOfficeHour_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_up_%d", suffix),
		"email":    fmt.Sprintf("ta_up_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 2, "start_time": "09:00", "end_time": "10:00", "location": "Old",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", created.ID), map[string]any{
		"course_id": 5, "day_of_week": 4, "start_time": "14:00", "end_time": "16:00", "location": "New",
	}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("update: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var oh struct {
		ID        int    `json:"id"`
		TAID      int    `json:"ta_id"`
		CourseID  int    `json:"course_id"`
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Location  string `json:"location"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&oh); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if oh.ID != created.ID || oh.TAID != ta.User.ID || oh.CourseID != 5 || oh.DayOfWeek != 4 {
		t.Fatalf("unexpected office hour: %+v", oh)
	}
	if oh.StartTime != "14:00:00" || oh.EndTime != "16:00:00" || oh.Location != "New" {
		t.Fatalf("unexpected times/location: %+v", oh)
	}
}

// Updating one OH to overlap another on the same day is rejected; in-place shrink is allowed.
func TestUpdateOfficeHour_OverlapRejected(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_up2_%d", suffix),
		"email":    fmt.Sprintf("ta_up2_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "10:00", "end_time": "12:00",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("slot a: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "14:00", "end_time": "15:00",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("slot b: %d %s", rr.Code, rr.Body.String())
	}
	var b struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&b); err != nil {
		t.Fatalf("decode b: %v", err)
	}

	// Move b into a's window => conflict
	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", b.ID), map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "11:00", "end_time": "13:00",
	}, ta.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("overlap update: expected 409 got %d body=%s", rr.Code, rr.Body.String())
	}

	// Shrink b in place (no touch with a) => OK
	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", b.ID), map[string]any{
		"course_id": 1, "day_of_week": 1, "start_time": "14:30", "end_time": "15:00",
	}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("shrink in place: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// Update returns 404 for missing ids, and 403 when the caller doesn't own the OH
// (or isn't a TA at all).
func TestUpdateOfficeHour_NotFoundAndForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_o_%d", suffix),
		"email":    fmt.Sprintf("ta_o_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ta1 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_p_%d", suffix),
		"email":    fmt.Sprintf("ta_p_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	ta2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 0, "start_time": "10:00", "end_time": "11:00",
	}, ta1.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPut, "/api/office-hours/999999", map[string]any{
		"course_id": 1, "day_of_week": 0, "start_time": "10:00", "end_time": "11:00",
	}, ta1.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing id: expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", created.ID), map[string]any{
		"course_id": 1, "day_of_week": 0, "start_time": "10:00", "end_time": "11:00",
	}, ta2.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("other TA: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_up_%d", suffix),
		"email":    fmt.Sprintf("st_up_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", created.ID), map[string]any{
		"course_id": 1, "day_of_week": 0, "start_time": "10:00", "end_time": "11:00",
	}, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// Owning TA can delete their office hour; a second delete returns 404.
func TestDeleteOfficeHour_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_del_%d", suffix),
		"email":    fmt.Sprintf("ta_del_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 5, "start_time": "10:00", "end_time": "11:00",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rr = doJSON(t, ts, http.MethodDelete, fmt.Sprintf("/api/office-hours/%d", created.ID), nil, ta.Token)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodDelete, fmt.Sprintf("/api/office-hours/%d", created.ID), nil, ta.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("delete again: expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// DELETE enforces: 401 (no token), 404 (missing id), 403 (wrong TA / student).
func TestDeleteOfficeHour_UnauthorizedNotFoundForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodDelete, "/api/office-hours/1", nil, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_d1_%d", suffix),
		"email":    fmt.Sprintf("ta_d1_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ta1 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_d2_%d", suffix),
		"email":    fmt.Sprintf("ta_d2_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	ta2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 2, "start_time": "09:00", "end_time": "10:00",
	}, ta1.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rr = doJSON(t, ts, http.MethodDelete, "/api/office-hours/999999", nil, ta1.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing id: expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodDelete, fmt.Sprintf("/api/office-hours/%d", created.ID), nil, ta2.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("other TA: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_del_%d", suffix),
		"email":    fmt.Sprintf("st_del_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodDelete, fmt.Sprintf("/api/office-hours/%d", created.ID), nil, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// GET by TA: empty-case returns []; after creating rows they come back ordered by day_of_week.
func TestListOfficeHoursByTA_HappyPathAndEmpty(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_list_%d", suffix),
		"email":    fmt.Sprintf("ta_list_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	path := fmt.Sprintf("/api/office-hours/ta/%d", ta.User.ID)
	rr = doJSON(t, ts, http.MethodGet, path, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list empty: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var empty []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&empty); err != nil {
		t.Fatalf("decode empty list: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no office hours, got %d", len(empty))
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 3, "day_of_week": 2, "start_time": "10:00", "end_time": "11:00", "location": "L1",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 3, "day_of_week": 1, "start_time": "14:00", "end_time": "15:00", "location": "L2",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create 2: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodGet, path, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var list []struct {
		ID        int    `json:"id"`
		TAID      int    `json:"ta_id"`
		CourseID  int    `json:"course_id"`
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Location  string `json:"location"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 office hours, got %d", len(list))
	}
	// Ordered by day_of_week: Monday (1) before Tuesday (2)
	if list[0].DayOfWeek != 1 || list[1].DayOfWeek != 2 {
		t.Fatalf("expected order Mon then Tue, got %+v", list)
	}
	if list[0].TAID != ta.User.ID || list[1].TAID != ta.User.ID {
		t.Fatalf("ta_id mismatch: %+v", list)
	}
}

// Non-integer TA id in the URL => 400.
func TestListOfficeHoursByTA_InvalidID(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	rr := doJSON(t, ts, http.MethodGet, "/api/office-hours/ta/notanumber", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid ta id: expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

// GET by course returns empty initially, then all matching rows across TAs.
func TestListOfficeHoursByCourse_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()
	courseID := int(suffix % 1000000)
	if courseID < 0 {
		courseID = -courseID
	}
	if courseID == 0 {
		courseID = 1
	}

	path := fmt.Sprintf("/api/office-hours/course/%d", courseID)
	rr := doJSON(t, ts, http.MethodGet, path, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list empty course: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var empty []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&empty); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(empty))
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_c1_%d", suffix),
		"email":    fmt.Sprintf("ta_c1_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ta1 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_c2_%d", suffix),
		"email":    fmt.Sprintf("ta_c2_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	ta2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": courseID, "day_of_week": 3, "start_time": "10:00", "end_time": "11:00", "location": "A",
	}, ta1.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create ta1 oh: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": courseID, "day_of_week": 3, "start_time": "14:00", "end_time": "15:00", "location": "B",
	}, ta2.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create ta2 oh: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodGet, path, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list: %d %s", rr.Code, rr.Body.String())
	}
	var list []struct {
		TAID      int `json:"ta_id"`
		CourseID  int `json:"course_id"`
		DayOfWeek int `json:"day_of_week"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list))
	}
	for _, row := range list {
		if row.CourseID != courseID {
			t.Fatalf("wrong course_id: %+v", row)
		}
	}
	taIDs := map[int]bool{ta1.User.ID: true, ta2.User.ID: true}
	if !taIDs[list[0].TAID] || !taIDs[list[1].TAID] {
		t.Fatalf("unexpected ta_ids: %+v", list)
	}
}

// Non-integer course id in the URL => 400.
func TestListOfficeHoursByCourse_InvalidID(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	rr := doJSON(t, ts, http.MethodGet, "/api/office-hours/course/bad", nil, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid course id: expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

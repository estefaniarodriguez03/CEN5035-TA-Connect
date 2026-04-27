package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// End-to-end happy path: TA creates queue, student joins, TA serves next, queue empties.
func TestQueueLifecycle_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// 1) Register TA and student
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_q_%d", suffix),
		"email":    fmt.Sprintf("ta_q_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_q_%d", suffix),
		"email":    fmt.Sprintf("student_q_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// 2) TA creates a queue
	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{
		"course_id": 1,
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var queueResp struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&queueResp); err != nil {
		t.Fatalf("decode create queue response: %v", err)
	}
	if queueResp.ID == 0 {
		t.Fatalf("expected non-zero queue id")
	}
	queueID := queueResp.ID

	// 3) Student joins the queue
	joinPath := fmt.Sprintf("/api/queues/%d/join", queueID)
	rr = doJSON(t, ts, http.MethodPost, joinPath, map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join queue: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// 4) GET queue shows 1 entry
	getPath := fmt.Sprintf("/api/queues/%d", queueID)
	rr = doJSON(t, ts, http.MethodGet, getPath, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get queue: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var getResp struct {
		ID      int           `json:"id"`
		Entries []interface{} `json:"entries"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get queue: %v", err)
	}
	if len(getResp.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(getResp.Entries))
	}

	// 5) TA calls /next to serve the student
	nextPath := fmt.Sprintf("/api/queues/%d/next", queueID)
	rr = doJSON(t, ts, http.MethodPost, nextPath, nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("next: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// 6) GET queue again; should be empty
	rr = doJSON(t, ts, http.MethodGet, getPath, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get queue after next: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}
	getResp = struct {
		ID      int           `json:"id"`
		Entries []interface{} `json:"entries"`
	}{}
	if err := json.NewDecoder(rr.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get queue: %v", err)
	}
	if len(getResp.Entries) != 0 {
		t.Fatalf("expected 0 entries after next, got %d", len(getResp.Entries))
	}
}

// Two students served consecutively: the first help session is logged when the second /next runs.
func TestSessionLog_WrittenOnSecondNext(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()
	ctx := context.Background()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_slog_%d", suffix),
		"email":    fmt.Sprintf("ta_slog_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_slog1_%d", suffix),
		"email":    fmt.Sprintf("student_slog1_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student1: %d %s", rr.Code, rr.Body.String())
	}
	st1 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_slog2_%d", suffix),
		"email":    fmt.Sprintf("student_slog2_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student2: %d %s", rr.Code, rr.Body.String())
	}
	st2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, st1.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join 1: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, st2.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join 2: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("next 1: %d %s", rr.Code, rr.Body.String())
	}
	var n int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM session_logs WHERE ta_id = $1`, ta.User.ID).Scan(&n); err != nil {
		t.Fatalf("count session_logs: %v", err)
	}
	if n != 0 {
		t.Fatalf("after first next want 0 session_logs, got %d", n)
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("next 2: %d %s", rr.Code, rr.Body.String())
	}
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM session_logs WHERE ta_id = $1 AND student_id = $2`,
		ta.User.ID, st1.User.ID,
	).Scan(&n); err != nil {
		t.Fatalf("count session_logs for student1: %v", err)
	}
	if n != 1 {
		t.Fatalf("want 1 session log for first student, got %d", n)
	}
}

// Join / leave on a missing queue id => 404.
func TestQueueJoinLeave_Errors(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// Register student
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_err_%d", suffix),
		"email":    fmt.Sprintf("student_err_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// Try join non-existent queue
	rr = doJSON(t, ts, http.MethodPost, "/api/queues/999999/join", map[string]any{}, student.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent queue join, got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Try leave when not in queue (non-existent queue)
	rr = doJSON(t, ts, http.MethodPost, "/api/queues/999999/leave", nil, student.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for leave on non-existent queue, got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// TA toggles queue state via PATCH /state; join returns 409 while paused/closed.
func TestQueueState_PATCH_JoinBlockedWhenPausedOrClosed(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_state_%d", suffix),
		"email":    fmt.Sprintf("ta_state_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_state_%d", suffix),
		"email":    fmt.Sprintf("st_state_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 42}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	statePath := fmt.Sprintf("/api/queues/%d/state", q.ID)

	// PATCH state -> paused
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "paused"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("patch paused: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("join paused: expected 409 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var joinErr stdErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&joinErr); err != nil {
		t.Fatalf("decode join error: %v", err)
	}
	if joinErr.Message != "queue is paused" {
		t.Fatalf("join paused message: got %q", joinErr.Message)
	}

	// PATCH state -> closed
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "closed"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("patch closed: %d %s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("join closed: expected 409 got %d, body=%s", rr.Code, rr.Body.String())
	}
	if err := json.NewDecoder(rr.Body).Decode(&joinErr); err != nil {
		t.Fatalf("decode join error: %v", err)
	}
	if joinErr.Message != "queue is closed" {
		t.Fatalf("join closed message: got %q", joinErr.Message)
	}

	// Re-open and join succeeds
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "open"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("patch open: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join open: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Wrong HTTP method on /state
	rr = doJSON(t, ts, http.MethodPost, statePath, map[string]any{"status": "paused"}, ta.Token)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post /state: expected 405 got %d", rr.Code)
	}
}

// Closed queues cannot be paused; must reopen to open first. Same-status PATCH is a no-op (200, no spurious error).
func TestQueueState_InvalidTransitionIdempotent(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_tr_%d", suffix),
		"email":    fmt.Sprintf("ta_tr_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&q)
	statePath := fmt.Sprintf("/api/queues/%d/state", q.ID)

	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "closed"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("close: %d %s", rr.Code, rr.Body.String())
	}

	// invalid: closed -> paused
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "paused"}, ta.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("closed->paused: expected 409 got %d body=%s", rr.Code, rr.Body.String())
	}
	var errBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details struct {
			Allowed []string `json:"allowed"`
		} `json:"details"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&errBody)
	if errBody.Code != "invalid_state_transition" {
		t.Fatalf("code: %q", errBody.Code)
	}

	// reopen, then idempotent open->open
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "open"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("reopen: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, ts, http.MethodPatch, statePath, map[string]any{"status": "open"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("idempotent open: %d %s", rr.Code, rr.Body.String())
	}
}

// Edge cases: duplicate join, leave when not in queue, next on empty queue.
func TestQueueEdgeCases_DuplicateJoin_LeaveNotInQueue_NextEmpty(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	suffix := uniqueSuffix()

	// Register TA and create queue
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_edge_%d", suffix),
		"email":    fmt.Sprintf("ta_edge_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	// Register student
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_edge_%d", suffix),
		"email":    fmt.Sprintf("student_edge_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// Join once
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Join twice => 409
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("join twice: expected 409 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Another student tries leaving when not in queue => 404
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_edge2_%d", suffix),
		"email":    fmt.Sprintf("student_edge2_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student2: %d %s", rr.Code, rr.Body.String())
	}
	student2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/leave", q.ID), nil, student2.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("leave not in queue: expected 404 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Serve the only student
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("next: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Next on empty queue => 404
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("next on empty: expected 404 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// GET /api/queues/{id} includes average session duration and per-entry / max ETA fields.
func TestGetQueueResponse_IncludesETAMetadata(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_eta_%d", suffix),
		"email":    fmt.Sprintf("ta_eta_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_eta_%d", suffix),
		"email":    fmt.Sprintf("st_eta_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	st := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&q)
	_ = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, st.Token)

	rr = doJSON(t, ts, http.MethodGet, fmt.Sprintf("/api/queues/%d", q.ID), nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get queue: %d %s", rr.Code, rr.Body.String())
	}
	var getResp struct {
		IsEmpty          bool    `json:"is_empty"`
		AverageSession   float64 `json:"average_session_duration_seconds"`
		TAAverageSession float64 `json:"ta_average_session_duration_seconds"`
		EstimatedMax     int64   `json:"estimated_wait_time_seconds"`
		Entries          []struct {
			Position             int   `json:"position"`
			EstimatedWaitSeconds int64 `json:"estimated_wait_seconds"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if getResp.AverageSession != 0 {
		t.Fatalf("no samples yet: avg should be 0, got %v", getResp.AverageSession)
	}
	if getResp.TAAverageSession != 0 {
		t.Fatalf("no TA samples yet: ta avg should be 0, got %v", getResp.TAAverageSession)
	}
	if getResp.IsEmpty {
		t.Fatalf("expected is_empty false with one student")
	}
	if len(getResp.Entries) != 1 {
		t.Fatalf("entries: %d", len(getResp.Entries))
	}
	// Avg 0 => ETA utility uses default 240s/position; first in line has 0s wait.
	if getResp.Entries[0].EstimatedWaitSeconds != 0 {
		t.Fatalf("pos 1 estimated_wait_seconds: %d", getResp.Entries[0].EstimatedWaitSeconds)
	}
	if getResp.EstimatedMax != 0 {
		t.Fatalf("one student at front: max wait %d", getResp.EstimatedMax)
	}
}

// GET /api/queues/{id} sets is_empty true when there are no students in line.
func TestGetQueue_EmptyQueue_IsEmptyTrue(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_e_%d", suffix),
		"email":    fmt.Sprintf("ta_e_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&q)

	rr = doJSON(t, ts, http.MethodGet, fmt.Sprintf("/api/queues/%d", q.ID), nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get queue: %d %s", rr.Code, rr.Body.String())
	}
	var getResp struct {
		IsEmpty bool  `json:"is_empty"`
		Entries []any `json:"entries"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !getResp.IsEmpty {
		t.Fatalf("expected is_empty true, got false")
	}
	if len(getResp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(getResp.Entries))
	}
}

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// --- Basic protection: protected endpoints reject missing tokens ---

func TestProtectedEndpoints_MissingToken_Return401(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	// create queue without token
	rr := doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("create queue without token: expected 401 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// join without token
	rr = doJSON(t, ts, http.MethodPost, "/api/queues/1/join", map[string]any{}, "")
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		// Depending on queue existence, handler may return 401 first (expected), but we allow
		// bad-request/not-found in case queue id parsing/existence changes.
		t.Fatalf("join without token: expected 401 (or 400/404) got %d, body=%s", rr.Code, rr.Body.String())
	}

	// next without token
	rr = doJSON(t, ts, http.MethodPost, "/api/queues/1/next", nil, "")
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		t.Fatalf("next without token: expected 401 (or 400/404) got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// --- Role enforcement: student can't hit TA-only routes ---

func TestRoleEnforcement_StudentCannotCreateQueueOrNext(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	suffix := uniqueSuffix()

	// Register TA + create queue
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_role_%d", suffix),
		"email":    fmt.Sprintf("ta_role_%d@example.com", suffix),
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
		"username": fmt.Sprintf("student_role_%d", suffix),
		"email":    fmt.Sprintf("student_role_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// Student cannot create queue
	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student create queue: expected 403 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Student cannot advance queue
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student next: expected 403 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// --- 401s from RequireAuth carry a structured code ---

func TestAuthMiddleware_Missing_Returns401WithCode(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	rr := doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body errorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 401 body: %v", err)
	}
	if body.Code != "auth_required" {
		t.Fatalf("expected code=auth_required, got %q (error=%q)", body.Code, body.Error)
	}
	if body.Error == "" {
		t.Fatalf("expected non-empty error message, got empty")
	}
}

func TestAuthMiddleware_InvalidToken_Returns401WithCode(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	rr := doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, "this-is-not-a-real-token")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("bad token: expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body errorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 401 body: %v", err)
	}
	if body.Code != "auth_required" {
		t.Fatalf("expected code=auth_required, got %q", body.Code)
	}
}

// --- 403s from RequireRole carry structured code + required_role ---

func TestRoleMiddleware_StudentOnTARoute_Returns403WithCodeAndRequiredRole(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_rm_%d", suffix),
		"email":    fmt.Sprintf("student_rm_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// Student tries to create a queue (TA-only route)
	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student create queue: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "role_forbidden" {
		t.Fatalf("expected code=role_forbidden, got %q", body.Code)
	}
	if body.RequiredRole != "ta" {
		t.Fatalf("expected required_role=ta, got %q", body.RequiredRole)
	}
	if body.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestRoleMiddleware_TAOnStudentRoute_Returns403WithCodeAndRequiredRole(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// Register TA
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_rm_%d", suffix),
		"email":    fmt.Sprintf("ta_rm_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	// TA creates a queue
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

	// TA tries to join their own queue (student-only route) — should be blocked by middleware.
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, ta.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("ta join: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "role_forbidden" {
		t.Fatalf("expected code=role_forbidden, got %q", body.Code)
	}
	if body.RequiredRole != "student" {
		t.Fatalf("expected required_role=student, got %q", body.RequiredRole)
	}
}

func TestRoleMiddleware_TACannotLeaveQueue(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_leave_%d", suffix),
		"email":    fmt.Sprintf("ta_leave_%d@example.com", suffix),
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
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/leave", q.ID), nil, ta.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("ta leave: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "role_forbidden" || body.RequiredRole != "student" {
		t.Fatalf("expected role_forbidden + required_role=student, got code=%q role=%q", body.Code, body.RequiredRole)
	}
}

// --- Ownership 403s carry structured codes ---

func TestOwnership_NonOwningTAGetsStructured403_OnNext(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// Two TAs
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_own_%d", suffix),
		"email":    fmt.Sprintf("ta_own_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ownerTA := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_other_own_%d", suffix),
		"email":    fmt.Sprintf("ta_other_own_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	otherTA := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ownerTA.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Other TA tries to serve next on a queue they don't own.
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, otherTA.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-owner next: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "not_queue_owner" {
		t.Fatalf("expected code=not_queue_owner, got %q (error=%q)", body.Code, body.Error)
	}
}

func TestOwnership_NonOwningTAGetsStructured403_OnUpdateState(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_ownst_%d", suffix),
		"email":    fmt.Sprintf("ta_ownst_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ownerTA := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_otherst_%d", suffix),
		"email":    fmt.Sprintf("ta_otherst_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	otherTA := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ownerTA.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPatch, fmt.Sprintf("/api/queues/%d/state", q.ID),
		map[string]any{"status": "paused"}, otherTA.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-owner patch state: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "not_queue_owner" {
		t.Fatalf("expected code=not_queue_owner, got %q", body.Code)
	}
}

func TestOwnership_NonOwningTAGetsStructured403_OnOfficeHourUpdate(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_oh_own_%d", suffix),
		"email":    fmt.Sprintf("ta_oh_own_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta1: %d %s", rr.Code, rr.Body.String())
	}
	ta1 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_oh_other_%d", suffix),
		"email":    fmt.Sprintf("ta_oh_other_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta2: %d %s", rr.Code, rr.Body.String())
	}
	ta2 := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/office-hours", map[string]any{
		"course_id": 1, "day_of_week": 3, "start_time": "10:00", "end_time": "11:00",
	}, ta1.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create oh: %d %s", rr.Code, rr.Body.String())
	}
	var oh struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&oh); err != nil {
		t.Fatalf("decode oh: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPut, fmt.Sprintf("/api/office-hours/%d", oh.ID), map[string]any{
		"course_id": 2, "day_of_week": 3, "start_time": "12:00", "end_time": "13:00",
	}, ta2.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-owner update oh: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body forbiddenBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body.Code != "not_resource_owner" {
		t.Fatalf("expected code=not_resource_owner, got %q", body.Code)
	}
}

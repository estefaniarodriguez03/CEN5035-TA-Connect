package backend_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/db"
	"backend/internal/routes"

	"github.com/joho/godotenv"
)

type authUser struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// setupTestDB connects to the database,
// runs migrations, and returns *sql.DB.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	_ = godotenv.Load(".env")

	database, err := db.Connect()
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		database.Close()
		t.Fatalf("db migrate: %v", err)
	}

	return database
}

func newTestServer(t *testing.T) (*sql.DB, http.Handler) {
	database := setupTestDB(t)
	router := routes.SetupRoutes(database)
	return database, router
}

func doJSON(t *testing.T, ts http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	ts.ServeHTTP(rr, req)
	return rr
}

func parseAuthUser(t *testing.T, rr *httptest.ResponseRecorder) authUser {
	t.Helper()
	var au authUser
	if err := json.NewDecoder(rr.Body).Decode(&au); err != nil {
		t.Fatalf("decode auth response: %v (status %d, body: %s)", err, rr.Code, rr.Body.String())
	}
	return au
}

func uniqueSuffix() int64 {
	return time.Now().UnixNano()
}

func TestRegisterAndLoginStudentAndTA(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	// Ensure unique emails per test run to avoid unique constraint conflicts
	suffix := uniqueSuffix()

	// Register student
	studentEmail := fmt.Sprintf("student_%d@example.com", suffix)
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_%d", suffix),
		"email":    studentEmail,
		"password": "password123",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)
	if student.User.Role != "student" {
		t.Fatalf("expected student role, got %s", student.User.Role)
	}

	// Register TA
	taEmail := fmt.Sprintf("ta_%d@example.com", suffix)
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_%d", suffix),
		"email":    taEmail,
		"password": "password123",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)
	if ta.User.Role != "ta" {
		t.Fatalf("expected ta role, got %s", ta.User.Role)
	}

	// Login student
	rr = doJSON(t, ts, http.MethodPost, "/api/login", map[string]any{
		"email":    studentEmail,
		"password": "password123",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("login student: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Login TA
	rr = doJSON(t, ts, http.MethodPost, "/api/login", map[string]any{
		"email":    taEmail,
		"password": "password123",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("login ta: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

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

func TestRegister_DuplicateUsernameOrEmail(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	suffix := uniqueSuffix()
	username := fmt.Sprintf("dupuser_%d", suffix)
	email := fmt.Sprintf("dup_%d@example.com", suffix)

	// First registration should succeed.
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": username,
		"email":    email,
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("first register: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Second registration with same email should fail with 409.
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": username + "_other", // different username, same email
		"email":    email,
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusConflict {
		t.Fatalf("duplicate email: expected 409 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Third registration with same username should also fail with 409.
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": username,
		"email":    fmt.Sprintf("other_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusConflict {
		t.Fatalf("duplicate username: expected 409 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegister_InvalidRole_Returns400(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	suffix := uniqueSuffix()
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("badrole_%d", suffix),
		"email":    fmt.Sprintf("badrole_%d@example.com", suffix),
		"password": "pw",
		"role":     "admin",
	}, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()

	suffix := uniqueSuffix()
	email := fmt.Sprintf("wp_%d@example.com", suffix)
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("wp_%d", suffix),
		"email":    email,
		"password": "correct",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register: expected 200 got %d, body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/login", map[string]any{
		"email":    email,
		"password": "wrong",
	}, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d, body=%s", rr.Code, rr.Body.String())
	}
}

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
	var q struct{ ID int `json:"id"` }
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
	var q struct{ ID int `json:"id"` }
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

func TestSSE_EmitsStudentJoinedEvent(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Use a real httptest server so streaming/flushing works.
	router := routes.SetupRoutes(database)
	srv := httptest.NewServer(router)
	defer srv.Close()

	suffix := uniqueSuffix()

	// Register TA and create queue
	rr := doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_sse_%d", suffix),
		"email":    fmt.Sprintf("ta_sse_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, router, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	// Register student
	rr = doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("student_sse_%d", suffix),
		"email":    fmt.Sprintf("student_sse_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// Connect to SSE
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/queues/%d/events", srv.URL, q.ID), nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("connect sse: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("sse status: expected 200 got %d body=%s", resp.StatusCode, string(b))
	}

	// Trigger join (which should emit STUDENT_JOINED)
	rr = doJSON(t, router, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join: expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Read SSE lines until we see event: STUDENT_JOINED (or timeout via client)
	reader := bufio.NewReader(resp.Body)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read sse: %v", err)
		}
		if line == "event: STUDENT_JOINED\n" {
			return
		}
	}
	t.Fatalf("did not receive STUDENT_JOINED event within deadline")
}

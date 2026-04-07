package backend_test

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// readSSEEventUntil reads from an SSE stream until an event with the given name is received,
// returning the raw JSON from the data: line (full queue event envelope).
func readSSEEventUntil(t *testing.T, reader *bufio.Reader, want string, deadline time.Time) []byte {
	t.Helper()
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read sse line: %v", err)
		}
		if !strings.HasPrefix(line, "event: ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "event: "))
		dataLine, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read sse data line: %v", err)
		}
		dataLine = strings.TrimSpace(dataLine)
		var payload []byte
		if strings.HasPrefix(dataLine, "data: ") {
			payload = []byte(strings.TrimSpace(strings.TrimPrefix(dataLine, "data: ")))
		}
		for {
			nl, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("read sse event trailer: %v", err)
			}
			if strings.TrimSpace(nl) == "" {
				break
			}
		}
		if name == want {
			return payload
		}
	}
	t.Fatalf("timeout waiting for SSE event %q", want)
	return nil
}

func TestSSE_EmitsStudentUpNextAfterJoin(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := routes.SetupRoutes(database)
	srv := httptest.NewServer(router)
	defer srv.Close()

	suffix := uniqueSuffix()

	rr := doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_un_%d", suffix),
		"email":    fmt.Sprintf("ta_un_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, router, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	rr = doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_un_%d", suffix),
		"email":    fmt.Sprintf("st_un_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	client := &http.Client{Timeout: 8 * time.Second}
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
		t.Fatalf("sse status: %d body=%s", resp.StatusCode, string(b))
	}

	reader := bufio.NewReader(resp.Body)

	rr = doJSON(t, router, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("join: %d %s", rr.Code, rr.Body.String())
	}

	deadline := time.Now().Add(4 * time.Second)
	raw := readSSEEventUntil(t, reader, "STUDENT_UP_NEXT", deadline)

	var envelope struct {
		Type    string `json:"type"`
		QueueID int    `json:"queue_id"`
		Payload struct {
			StudentID int `json:"student_id"`
			Position  int `json:"position"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal STUDENT_UP_NEXT: %v raw=%s", err, string(raw))
	}
	if envelope.Type != "STUDENT_UP_NEXT" {
		t.Fatalf("envelope.type: got %q", envelope.Type)
	}
	if envelope.QueueID != q.ID {
		t.Fatalf("envelope.queue_id: got %d want %d", envelope.QueueID, q.ID)
	}
	if envelope.Payload.Position != 1 {
		t.Fatalf("expected head-of-queue position 1, got %d", envelope.Payload.Position)
	}
	if envelope.Payload.StudentID != student.User.ID {
		t.Fatalf("payload.student_id: got %d want %d (joining user)", envelope.Payload.StudentID, student.User.ID)
	}
}

func TestSSE_EmitsAnnouncementSent(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := routes.SetupRoutes(database)
	srv := httptest.NewServer(router)
	defer srv.Close()

	suffix := uniqueSuffix()

	rr := doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_ann_%d", suffix),
		"email":    fmt.Sprintf("ta_ann_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, router, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	client := &http.Client{Timeout: 8 * time.Second}
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
		t.Fatalf("sse status: %d body=%s", resp.StatusCode, string(b))
	}

	reader := bufio.NewReader(resp.Body)

	msg := fmt.Sprintf("Test announcement %d", suffix)
	rr = doJSON(t, router, http.MethodPost, fmt.Sprintf("/api/queues/%d/announcement", q.ID), map[string]any{
		"message": msg,
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("announcement: %d %s", rr.Code, rr.Body.String())
	}

	deadline := time.Now().Add(4 * time.Second)
	raw := readSSEEventUntil(t, reader, "ANNOUNCEMENT_SENT", deadline)

	var envelope struct {
		Type    string `json:"type"`
		QueueID int    `json:"queue_id"`
		Payload struct {
			Message string `json:"message"`
			TAID    int    `json:"ta_id"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal ANNOUNCEMENT_SENT: %v raw=%s", err, string(raw))
	}
	if envelope.Type != "ANNOUNCEMENT_SENT" {
		t.Fatalf("envelope.type: got %q", envelope.Type)
	}
	if envelope.QueueID != q.ID {
		t.Fatalf("envelope.queue_id: got %d want %d", envelope.QueueID, q.ID)
	}
	if envelope.Payload.Message != msg {
		t.Fatalf("payload.message: got %q want %q", envelope.Payload.Message, msg)
	}
	if envelope.Payload.TAID != ta.User.ID {
		t.Fatalf("payload.ta_id: got %d want %d", envelope.Payload.TAID, ta.User.ID)
	}
}

func TestPostAnnouncement_HappyPath(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_pa_%d", suffix),
		"email":    fmt.Sprintf("ta_pa_%d@example.com", suffix),
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
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	body := fmt.Sprintf("Office hours moved to room 101 (%d)", suffix)
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/announcement", q.ID), map[string]any{
		"message": body,
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("announcement: %d %s", rr.Code, rr.Body.String())
	}
	var out struct {
		ID        int    `json:"id"`
		QueueID   int    `json:"queue_id"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode announcement response: %v", err)
	}
	if out.ID == 0 || out.QueueID != q.ID || out.Message != body || out.CreatedAt == "" {
		t.Fatalf("unexpected announcement response: %+v", out)
	}
}

func TestPostAnnouncement_StudentForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_ps_%d", suffix),
		"email":    fmt.Sprintf("ta_ps_%d@example.com", suffix),
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
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_ps_%d", suffix),
		"email":    fmt.Sprintf("st_ps_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/announcement", q.ID), map[string]any{
		"message": "nope",
	}, student.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student announcement: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestPostAnnouncement_NonOwningTAForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_owner_%d", suffix),
		"email":    fmt.Sprintf("ta_owner_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta owner: %d %s", rr.Code, rr.Body.String())
	}
	taOwner := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_other_%d", suffix),
		"email":    fmt.Sprintf("ta_other_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta other: %d %s", rr.Code, rr.Body.String())
	}
	taOther := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 1}, taOwner.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct{ ID int `json:"id"` }
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/announcement", q.ID), map[string]any{
		"message": "from wrong TA",
	}, taOther.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-owner announcement: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

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

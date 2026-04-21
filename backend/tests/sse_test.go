package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/routes"
)

// PATCH /state emits a QUEUE_STATE_CHANGED envelope with previous_status + status.
func TestSSE_EmitsQueueStateChanged(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	router := routes.SetupRoutes(database)
	srv := httptest.NewServer(router)
	defer srv.Close()

	suffix := uniqueSuffix()

	rr := doJSON(t, router, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_qsc_%d", suffix),
		"email":    fmt.Sprintf("ta_qsc_%d@example.com", suffix),
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
	var q struct {
		ID int `json:"id"`
	}
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

	statePath := fmt.Sprintf("/api/queues/%d/state", q.ID)
	rr = doJSON(t, router, http.MethodPatch, statePath, map[string]any{"status": "paused"}, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("patch state: %d %s", rr.Code, rr.Body.String())
	}

	deadline := time.Now().Add(4 * time.Second)
	raw := readSSEEventUntil(t, reader, "QUEUE_STATE_CHANGED", deadline)

	var envelope struct {
		Type    string `json:"type"`
		QueueID int    `json:"queue_id"`
		Payload struct {
			PreviousStatus string `json:"previous_status"`
			Status         string `json:"status"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal QUEUE_STATE_CHANGED: %v raw=%s", err, string(raw))
	}
	if envelope.Payload.PreviousStatus != "open" || envelope.Payload.Status != "paused" {
		t.Fatalf("payload: %+v", envelope.Payload)
	}
}

// A student joining emits the STUDENT_JOINED event.
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
	var q struct {
		ID int `json:"id"`
	}
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

// STUDENT_UP_NEXT fires for whoever is at position 1 after a join.
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
	var q struct {
		ID int `json:"id"`
	}
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

// ANNOUNCEMENT_SENT payload carries the exact TA id and message that was posted.
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
	var q struct {
		ID int `json:"id"`
	}
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

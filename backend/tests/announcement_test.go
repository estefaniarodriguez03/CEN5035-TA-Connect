package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TA can post an announcement; response echoes id/queue_id/message/created_at.
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
	var q struct {
		ID int `json:"id"`
	}
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

// Students cannot post announcements.
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
	var q struct {
		ID int `json:"id"`
	}
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

// A TA who doesn't own the queue cannot post announcements to it.
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
	var q struct {
		ID int `json:"id"`
	}
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

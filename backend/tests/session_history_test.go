package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestSessionHistory_Unauthorized(t *testing.T) {
	_, ts := newTestServer(t)
	rr := doJSON(t, ts, http.MethodGet, "/api/session-history", nil, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSessionHistory_TAandStudent_SeeRelevantRows(t *testing.T) {
	_, ts := newTestServer(t)
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_hist_%d", suffix),
		"email":    fmt.Sprintf("ta_hist_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_hist_%d", suffix),
		"email":    fmt.Sprintf("st_hist_%d@example.com", suffix),
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
	rr = doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st2_hist_%d", suffix),
		"email":    fmt.Sprintf("st2_hist_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student2: %d %s", rr.Code, rr.Body.String())
	}
	st2 := parseAuthUser(t, rr)
	_ = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, st2.Token)

	_ = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
	_ = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)

	rr = doJSON(t, ts, http.MethodGet, "/api/session-history", nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("ta session-history: %d %s", rr.Code, rr.Body.String())
	}
	var taBody struct {
		Sessions []struct {
			StudentID       int     `json:"student_id"`
			TAID            int     `json:"ta_id"`
			DurationSeconds float64 `json:"duration"`
		} `json:"sessions"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&taBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(taBody.Sessions) != 1 {
		t.Fatalf("ta want 1 session, got %d", len(taBody.Sessions))
	}
	if taBody.Sessions[0].StudentID != st.User.ID {
		t.Fatalf("ta row student_id: got %d want %d", taBody.Sessions[0].StudentID, st.User.ID)
	}
	if taBody.Sessions[0].TAID != ta.User.ID {
		t.Fatalf("ta row ta_id: got %d want %d", taBody.Sessions[0].TAID, ta.User.ID)
	}

	rr = doJSON(t, ts, http.MethodGet, "/api/session-history", nil, st.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("student session-history: %d %s", rr.Code, rr.Body.String())
	}
	var stBody struct {
		Sessions []struct {
			StudentID int `json:"student_id"`
		} `json:"sessions"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&stBody); err != nil {
		t.Fatalf("decode student: %v", err)
	}
	if len(stBody.Sessions) != 1 {
		t.Fatalf("student want 1 session, got %d", len(stBody.Sessions))
	}
	if stBody.Sessions[0].StudentID != st.User.ID {
		t.Fatalf("student row: student_id %d", stBody.Sessions[0].StudentID)
	}

	rr = doJSON(t, ts, http.MethodGet, "/api/session-history", nil, st2.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("student2 session-history: %d %s", rr.Code, rr.Body.String())
	}
	var st2Body struct {
		Sessions []any `json:"sessions"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&st2Body)
	if len(st2Body.Sessions) != 0 {
		t.Fatalf("student2 not in a completed session yet: want 0 got %d", len(st2Body.Sessions))
	}
}

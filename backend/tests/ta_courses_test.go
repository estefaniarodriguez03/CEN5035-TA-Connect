package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestTACourses_AddAndList(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_course_%d", suffix),
		"email":    fmt.Sprintf("ta_course_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/ta/courses", map[string]any{
		"code": "CEN5035",
		"name": "Software Engineering",
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("add ta course: expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}
	var addResp struct {
		TAID   int `json:"ta_id"`
		Course struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"course"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&addResp); err != nil {
		t.Fatalf("decode add response: %v", err)
	}
	if addResp.TAID != ta.User.ID {
		t.Fatalf("ta_id mismatch: got %d want %d", addResp.TAID, ta.User.ID)
	}
	if addResp.Course.Code != "CEN5035" {
		t.Fatalf("course code mismatch: got %q", addResp.Course.Code)
	}

	// Idempotent link insert.
	rr = doJSON(t, ts, http.MethodPost, "/api/ta/courses", map[string]any{
		"course_id": addResp.Course.ID,
	}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("idempotent add ta course: expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, ts, http.MethodGet, "/api/ta/courses", nil, ta.Token)
	if rr.Code != http.StatusOK {
		t.Fatalf("list ta courses: expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var listResp struct {
		Courses []struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"courses"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Courses) != 1 {
		t.Fatalf("expected 1 ta course, got %d", len(listResp.Courses))
	}
	if listResp.Courses[0].ID != addResp.Course.ID {
		t.Fatalf("course id mismatch: got %d want %d", listResp.Courses[0].ID, addResp.Course.ID)
	}
}

func TestTACourses_StudentForbidden(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("st_course_%d", suffix),
		"email":    fmt.Sprintf("st_course_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	st := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/ta/courses", map[string]any{
		"code": "CEN5035",
	}, st.Token)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("student add ta course: expected 403 got %d body=%s", rr.Code, rr.Body.String())
	}
}

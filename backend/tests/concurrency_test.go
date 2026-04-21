package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

// Concurrent joins: N students joining in parallel all receive unique, contiguous positions 1..N.
// This is the payoff for wrapping Join in a transaction + SELECT ... FOR UPDATE on the queue row.
func TestConcurrentJoins_PositionsUniqueAndContiguous(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// TA creates a queue.
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_conc_%d", suffix),
		"email":    fmt.Sprintf("ta_conc_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

	rr = doJSON(t, ts, http.MethodPost, "/api/queues", map[string]any{"course_id": 777}, ta.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create queue: %d %s", rr.Code, rr.Body.String())
	}
	var q struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("decode queue: %v", err)
	}

	// Register N students up front.
	const N = 10
	tokens := make([]string, 0, N)
	for i := 0; i < N; i++ {
		rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
			"username": fmt.Sprintf("stc_%d_%d", suffix, i),
			"email":    fmt.Sprintf("stc_%d_%d@example.com", suffix, i),
			"password": "pw",
			"role":     "student",
		}, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("register student %d: %d %s", i, rr.Code, rr.Body.String())
		}
		tokens = append(tokens, parseAuthUser(t, rr).Token)
	}

	// Fire all N joins concurrently.
	type joinResult struct {
		status   int
		position int
		body     string
	}
	results := make([]joinResult, N)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			rr := doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, tokens[i])
			results[i].status = rr.Code
			results[i].body = rr.Body.String()
			if rr.Code == http.StatusCreated {
				var resp struct {
					Position int `json:"position"`
				}
				_ = json.Unmarshal([]byte(results[i].body), &resp)
				results[i].position = resp.Position
			}
		}(i)
	}
	wg.Wait()

	// All should succeed; positions should be 1..N with no duplicates.
	seen := make(map[int]bool, N)
	for i, r := range results {
		if r.status != http.StatusCreated {
			t.Fatalf("student %d join: expected 201 got %d body=%s", i, r.status, r.body)
		}
		if r.position < 1 || r.position > N {
			t.Fatalf("student %d position out of range: %d", i, r.position)
		}
		if seen[r.position] {
			t.Fatalf("duplicate position %d across concurrent joins", r.position)
		}
		seen[r.position] = true
	}
	if len(seen) != N {
		t.Fatalf("expected %d unique positions, got %d", N, len(seen))
	}
}

// A second join by the same student hits the UNIQUE(queue_id, student_id) constraint and returns 409.
func TestDuplicateJoin_UniqueConstraintReturns409(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_dup_%d", suffix),
		"email":    fmt.Sprintf("ta_dup_%d@example.com", suffix),
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
		"username": fmt.Sprintf("st_dup_%d", suffix),
		"email":    fmt.Sprintf("st_dup_%d@example.com", suffix),
		"password": "pw",
		"role":     "student",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register student: %d %s", rr.Code, rr.Body.String())
	}
	student := parseAuthUser(t, rr)

	// First join succeeds
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first join: expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	// Second join hits the UNIQUE(queue_id, student_id) constraint
	rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, student.Token)
	if rr.Code != http.StatusConflict {
		t.Fatalf("second join: expected 409 got %d body=%s", rr.Code, rr.Body.String())
	}
	var body errorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 409: %v", err)
	}
	if body.Error != "already in queue" {
		t.Fatalf("expected 'already in queue', got %q", body.Error)
	}
}

// Concurrent /next calls can never double-serve the same student, courtesy of
// the transaction + SKIP LOCKED in the ServeNext handler.
func TestConcurrentServeNext_NeverDoubleServesSameStudent(t *testing.T) {
	database, ts := newTestServer(t)
	defer database.Close()
	suffix := uniqueSuffix()

	// Register TA and create queue
	rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
		"username": fmt.Sprintf("ta_next_%d", suffix),
		"email":    fmt.Sprintf("ta_next_%d@example.com", suffix),
		"password": "pw",
		"role":     "ta",
	}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("register ta: %d %s", rr.Code, rr.Body.String())
	}
	ta := parseAuthUser(t, rr)

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

	// Put N students in the queue.
	const N = 5
	expectedStudentIDs := make(map[int]bool, N)
	for i := 0; i < N; i++ {
		rr := doJSON(t, ts, http.MethodPost, "/api/register", map[string]any{
			"username": fmt.Sprintf("stn_%d_%d", suffix, i),
			"email":    fmt.Sprintf("stn_%d_%d@example.com", suffix, i),
			"password": "pw",
			"role":     "student",
		}, "")
		if rr.Code != http.StatusOK {
			t.Fatalf("register student %d: %d %s", i, rr.Code, rr.Body.String())
		}
		s := parseAuthUser(t, rr)
		expectedStudentIDs[s.User.ID] = true

		rr = doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/join", q.ID), map[string]any{}, s.Token)
		if rr.Code != http.StatusCreated {
			t.Fatalf("join %d: %d %s", i, rr.Code, rr.Body.String())
		}
	}

	// Fire concurrent /next calls greater than the queue size; extras should get 404 "queue is empty".
	const Calls = N + 3
	type nextRes struct {
		status    int
		studentID int
		body      string
	}
	results := make([]nextRes, Calls)
	var wg sync.WaitGroup
	wg.Add(Calls)
	for i := 0; i < Calls; i++ {
		go func(i int) {
			defer wg.Done()
			rr := doJSON(t, ts, http.MethodPost, fmt.Sprintf("/api/queues/%d/next", q.ID), nil, ta.Token)
			results[i].status = rr.Code
			results[i].body = rr.Body.String()
			if rr.Code == http.StatusOK {
				var resp struct {
					Student struct {
						StudentID int `json:"student_id"`
					} `json:"student"`
				}
				_ = json.Unmarshal([]byte(results[i].body), &resp)
				results[i].studentID = resp.Student.StudentID
			}
		}(i)
	}
	wg.Wait()

	served := make(map[int]bool)
	successes := 0
	notFounds := 0
	for i, r := range results {
		switch r.status {
		case http.StatusOK:
			successes++
			if r.studentID == 0 {
				t.Fatalf("next[%d]: 200 but missing student_id. body=%s", i, r.body)
			}
			if served[r.studentID] {
				t.Fatalf("student %d was served more than once!", r.studentID)
			}
			if !expectedStudentIDs[r.studentID] {
				t.Fatalf("served unknown student id %d", r.studentID)
			}
			served[r.studentID] = true
		case http.StatusNotFound:
			notFounds++
		default:
			t.Fatalf("next[%d]: unexpected status %d body=%s", i, r.status, r.body)
		}
	}
	if successes != N {
		t.Fatalf("expected exactly %d successes, got %d (not_founds=%d)", N, successes, notFounds)
	}
	if notFounds != Calls-N {
		t.Fatalf("expected %d 404s, got %d", Calls-N, notFounds)
	}
}

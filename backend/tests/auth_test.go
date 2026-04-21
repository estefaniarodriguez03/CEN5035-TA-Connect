package tests

import (
	"fmt"
	"net/http"
	"testing"
)

// Register + login happy path for both roles.
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

// UNIQUE(email) and UNIQUE(username) constraints must translate to 409.
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

// Registration with a role outside {student,ta} must be rejected with 400.
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

// Wrong password on login returns 401.
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

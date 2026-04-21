package tests

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/db"
	"backend/internal/routes"

	"github.com/joho/godotenv"
)

// authUser mirrors the /api/register and /api/login response bodies.
type authUser struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// forbiddenBody mirrors auth.ForbiddenResponse for decoding test responses.
type forbiddenBody struct {
	Error        string `json:"error"`
	Code         string `json:"code"`
	RequiredRole string `json:"required_role,omitempty"`
}

// errorBody is the generic {"error","code"} JSON shape used for 401/409/etc.
type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// setupTestDB connects to the database, runs migrations, and returns *sql.DB.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if err := godotenv.Load(".env"); err != nil {
		_ = godotenv.Load("../.env")
	}

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

// newTestServer wires the router around a fresh migrated DB.
func newTestServer(t *testing.T) (*sql.DB, http.Handler) {
	database := setupTestDB(t)
	router := routes.SetupRoutes(database)
	return database, router
}

// doJSON sends a JSON request to the in-memory handler and returns the recorder.
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

// parseAuthUser decodes a /register or /login response body.
func parseAuthUser(t *testing.T, rr *httptest.ResponseRecorder) authUser {
	t.Helper()
	var au authUser
	if err := json.NewDecoder(rr.Body).Decode(&au); err != nil {
		t.Fatalf("decode auth response: %v (status %d, body: %s)", err, rr.Code, rr.Body.String())
	}
	return au
}

// uniqueSuffix returns a monotonically increasing value for generating
// unique emails/usernames within a test run.
func uniqueSuffix() int64 {
	return time.Now().UnixNano()
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

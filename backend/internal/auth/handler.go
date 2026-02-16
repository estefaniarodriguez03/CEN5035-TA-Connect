package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// LoginRequest is the JSON body for POST /api/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the JSON body for POST /api/register.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "student" or "ta"
}

// AuthResponse is returned on successful login or register.
type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// Login handles POST /api/login. Expects JSON: { "email", "password" }.
func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
			return
		}

		var id int
		var username, email, passwordHash, role string
		err := db.QueryRowContext(r.Context(),
			`SELECT id, username, email, password, role FROM users WHERE email = $1`, req.Email).Scan(&id, &username, &email, &passwordHash, &role)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}
		if !CheckPassword(passwordHash, req.Password) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}

		token, err := NewToken(id, email, role)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create token"})
			return
		}
		writeAuthResponse(w, token, id, username, email, role)
	}
}

// Register handles POST /api/register. Expects JSON: { "username", "email", "password", "role" }.
func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if req.Username == "" || req.Email == "" || req.Password == "" || req.Role == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username, email, password and role required"})
			return
		}
		if req.Role != "student" && req.Role != "ta" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be 'student' or 'ta'"})
			return
		}

		hash, err := HashPassword(req.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not hash password"})
			return
		}

		var id int
		err = db.QueryRowContext(r.Context(),
			`INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Username, req.Email, hash, req.Role).Scan(&id)
		if err != nil {
			if isUniqueViolation(err) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "username or email already in use"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
			return
		}

		token, err := NewToken(id, req.Email, req.Role)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create token"})
			return
		}
		writeAuthResponse(w, token, id, req.Username, req.Email, req.Role)
	}
}

func writeAuthResponse(w http.ResponseWriter, token string, id int, username, email, role string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := AuthResponse{
		Token: token,
	}
	resp.User.ID = id
	resp.User.Username = username
	resp.User.Email = email
	resp.User.Role = role
	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// isUniqueViolation checks for PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate key")
}

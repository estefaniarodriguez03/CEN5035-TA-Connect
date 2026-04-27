package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/internal/httperr"
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
			httperr.Write(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_json", "invalid JSON", nil)
			return
		}
		if req.Email == "" || req.Password == "" {
			httperr.Write(w, http.StatusBadRequest, "validation_error", "email and password required", nil)
			return
		}

		var id int
		var username, email, passwordHash, role string
		err := db.QueryRowContext(r.Context(),
			`SELECT id, username, email, password, role FROM users WHERE email = $1`, req.Email).Scan(&id, &username, &email, &passwordHash, &role)
		if err == sql.ErrNoRows {
			httperr.Write(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
			return
		}
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}
		if !CheckPassword(passwordHash, req.Password) {
			httperr.Write(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
			return
		}

		token, err := NewToken(id, email, role)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "token_unavailable", "could not create token", nil)
			return
		}
		writeAuthResponse(w, token, id, username, email, role)
	}
}

// Register handles POST /api/register. Expects JSON: { "username", "email", "password", "role" }.
func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httperr.Write(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_json", "invalid JSON", nil)
			return
		}
		if req.Username == "" || req.Email == "" || req.Password == "" || req.Role == "" {
			httperr.Write(w, http.StatusBadRequest, "validation_error", "username, email, password and role required", nil)
			return
		}
		if req.Role != "student" && req.Role != "ta" {
			httperr.Write(w, http.StatusBadRequest, "invalid_role", "role must be 'student' or 'ta'", nil)
			return
		}

		hash, err := HashPassword(req.Password)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "password_hash_failed", "could not hash password", nil)
			return
		}

		var id int
		err = db.QueryRowContext(r.Context(),
			`INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Username, req.Email, hash, req.Role).Scan(&id)
		if err != nil {
			if IsUniqueViolation(err) {
				httperr.Write(w, http.StatusConflict, "conflict_unique", "username or email already in use", nil)
				return
			}
			httperr.Write(w, http.StatusInternalServerError, "internal_error", "database error", nil)
			return
		}

		token, err := NewToken(id, req.Email, req.Role)
		if err != nil {
			httperr.Write(w, http.StatusInternalServerError, "token_unavailable", "could not create token", nil)
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

// IsUniqueViolation checks for PostgreSQL unique constraint violation.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate key")
}

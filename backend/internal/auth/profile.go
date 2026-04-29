package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/httperr"

	"github.com/go-chi/chi/v5"
)

// UpdateProfileRequest is the JSON body for PUT /api/users/{id}/profile.
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Courses  []struct {
		ID   int    `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"courses"`
}

// UpdateProfileResponse is returned on successful profile update.
type UpdateProfileResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// UpdateProfile handles PUT /api/users/{id}/profile.
func UpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			httperr.Write(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}

		// Get user ID from URL parameter
		userIDStr := chi.URLParam(r, "id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
			return
		}

		// Get authenticated user ID from claims
		claims := ClaimsFromContext(r.Context())
		if claims == nil || claims.UserID == 0 {
			httperr.Write(w, http.StatusUnauthorized, "unauthorized", "not authenticated", nil)
			return
		}

		// Only allow users to update their own profile
		if claims.UserID != userID {
			httperr.Write(w, http.StatusForbidden, "forbidden", "cannot update other user profiles", nil)
			return
		}

		var req UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httperr.Write(w, http.StatusBadRequest, "invalid_json", "invalid JSON", nil)
			return
		}

		// Validate username if provided
		if req.Username == "" {
			httperr.Write(w, http.StatusBadRequest, "validation_error", "username is required", nil)
			return
		}

		// Update username in database
		query := "UPDATE users SET username = $1 WHERE id = $2"
		if err := db.QueryRow(query, req.Username, userID).Err(); err != nil {
			httperr.Write(w, http.StatusInternalServerError, "db_error", "failed to update profile", nil)
			return
		}

		// Get updated user info
		var user struct {
			ID       int
			Username string
			Email    string
			Role     string
		}

		const getUserQuery = "SELECT id, username, email, role FROM users WHERE id = $1"
		if err := db.QueryRow(getUserQuery, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Role); err != nil {
			if err == sql.ErrNoRows {
				httperr.Write(w, http.StatusNotFound, "not_found", "user not found", nil)
				return
			}
			httperr.Write(w, http.StatusInternalServerError, "db_error", "failed to fetch user", nil)
			return
		}

		// Return updated user info
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := UpdateProfileResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

package auth

import (
	"context"
	"net/http"

	"backend/internal/httperr"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

// ClaimsFromContext retrieves the JWT claims stored by RequireAuth middleware.
// Returns nil if no claims are present (should not happen behind RequireAuth).
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}

// RequireAuth is Chi middleware that validates the JWT Bearer token and
// injects *Claims into the request context. Returns 401 on missing/invalid token.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaimsFromRequest(r)
		if err != nil || claims == nil {
			httperr.Write(w, http.StatusUnauthorized, "auth_required", "authorization required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns Chi middleware that checks the authenticated user has one
// of the allowed roles. Must be placed after RequireAuth.
// Returns a structured 403 if the role does not match.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				httperr.Write(w, http.StatusUnauthorized, "auth_required", "authorization required", nil)
				return
			}
			for _, role := range allowed {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			var details any
			if len(allowed) == 1 {
				details = map[string]string{"required_role": allowed[0]}
			} else {
				details = map[string]any{"allowed_roles": allowed}
			}
			httperr.Write(w, http.StatusForbidden, "role_forbidden", "you do not have permission to access this resource", details)
		})
	}
}

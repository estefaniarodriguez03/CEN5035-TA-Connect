// Package httperr provides a single JSON error shape for all API error responses.
package httperr

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API error body: { "code", "message", "details" }.
type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

// Write sends a JSON error response. details may be nil (encoded as JSON null).
func Write(w http.ResponseWriter, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Details: details,
	})
}

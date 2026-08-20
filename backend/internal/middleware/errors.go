package middleware

import (
	"encoding/json"
	"net/http"

	"pgcs/backend/internal/apierr"
)

// HandlerFunc is a net/http-compatible handler that may return an error.
// Returning a non-nil error lets handlers stay thin: they delegate to services
// and simply propagate typed errors, which Error then renders into the
// consistent APIError body with the correct status code.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Error adapts a HandlerFunc into a standard http.Handler. When the handler
// returns a non-nil error, Error translates it to the appropriate HTTP status
// and structured APIError body (Req 18.1, 18.5). Unexpected/untyped errors are
// rendered as a generic 500 that never leaks internal detail (Req 18.4).
//
// If the handler returns nil it is assumed to have written its own response.
func Error(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, err)
		}
	})
}

// WriteError serializes err into the structured APIError JSON body and writes
// it with the mapped HTTP status code. It is safe to call for any error: typed
// service errors are mapped per the design table, and any other error becomes a
// generic 500 (Req 18.4).
func WriteError(w http.ResponseWriter, err error) {
	status, body := apierr.Translate(err)
	WriteJSON(w, status, body)
}

// WriteJSON writes v as a JSON response with the given status code and the
// application/json content type.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Best-effort encode. If encoding fails the status/headers are already
	// committed; there is nothing safe left to do but stop.
	_ = json.NewEncoder(w).Encode(v)
}

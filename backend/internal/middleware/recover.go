package middleware

import (
	"log"
	"net/http"

	"pgcs/backend/internal/apierr"
)

// Recover wraps next with a deferred panic handler. If any downstream handler
// panics, Recover traps it, logs the panic value server-side for diagnostics,
// and writes a generic 500 response with the structured APIError body. The
// stack trace and panic value are never written to the response, so internal
// detail and secrets cannot leak to the client (Req 18.4).
//
// Because the panic is contained per request, a panic in one handler cannot
// crash the server or affect other in-flight requests.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Log server-side only. The response stays generic.
				log.Printf("recovered from panic handling %s %s: %v", r.Method, r.URL.Path, rec)
				WriteJSON(w, http.StatusInternalServerError, apierr.Internal())
			}
		}()
		next.ServeHTTP(w, r)
	})
}

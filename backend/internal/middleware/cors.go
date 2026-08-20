package middleware

import "net/http"

// CORS wraps next with permissive cross-origin headers so browser-based
// clients (e.g. the React Native Web build served from a different origin such
// as http://localhost:8081) can call the API.
//
// The API authenticates with a Bearer token in the Authorization header rather
// than cookies, so it does not need credentialed CORS. That lets us reflect a
// wildcard origin ("*") without the browser restrictions that apply when
// Access-Control-Allow-Credentials is true.
//
// Preflight OPTIONS requests are answered immediately with 204 and the allow
// headers, so they never reach the route table.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		h.Set("Access-Control-Max-Age", "86400")

		// Short-circuit the CORS preflight request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

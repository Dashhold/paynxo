// Package api wires the HTTP router, request decoding/encoding, and handlers.
//
// NewRouter builds the application handler: a Go 1.22+ method-aware ServeMux
// wrapped with the global middleware chain (Recover and structured route-error
// rendering), with per-route Auth/TenantScope/Error middleware applied as each
// route is registered. Handlers are thin error-returning functions
// (middleware.HandlerFunc) that delegate to services and propagate typed errors
// for the Error middleware to render.
//
// Implemented in task 6.1:
//   - the router and its middleware wiring (router.go);
//   - the authentication endpoints POST /api/auth/login, POST /api/auth/logout,
//     and GET /api/me (auth.go);
//   - the role-authorization helpers RequireRole and RequireRoles (authz.go);
//   - structured 404 (unknown route) and 405 (unsupported method) responses
//     (Req 1.7).
//
// The protected(...) chain (Auth + TenantScope) and the RequireRoles decorator
// are the extension points later tasks use to mount the tenant-scoped,
// role-guarded business, ledger, report, and lease route groups.
package api

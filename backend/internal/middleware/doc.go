// Package middleware provides cross-cutting HTTP middleware.
//
// Implemented in task 4.1:
//   - Error: adapts error-returning handlers (HandlerFunc) into net/http
//     handlers, rendering typed errors from internal/apierr as the structured
//     APIError body with the correct status code.
//   - Recover: traps panics per request and returns a safe generic 500 without
//     leaking stack traces or secret values.
//
// Implemented in task 5.2:
//   - Auth: validates the Session_Token via the Auth_Service, places the
//     authenticated service.Principal into the request context, and rejects
//     missing/invalid/expired/revoked tokens with 401 and non-active leases
//     with 403.
//   - TenantScope: guards that an authenticated principal is present (set by
//     Auth) so the tenant-scoped repository layer always has a scope to apply.
//   - WithPrincipal / PrincipalFromContext: context helpers used to carry the
//     principal from Auth through to handlers and repositories.
//
// The router wiring that mounts these middleware is added in task 6.1.
package middleware

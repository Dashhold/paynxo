package middleware

import (
	"net/http"

	"pgcs/backend/internal/apierr"
)

// TenantScope is the explicit boundary that guarantees a tenant-scoped
// principal is present before any business handler runs. It must be mounted
// after Auth, which performs token validation and stores the principal in the
// request context.
//
// The principal already carries everything the repository layer needs to scope
// data access — TenantID plus the owner fields for Company/Affiliate/Merchant
// portals (Req 4, 7.5). The actual GORM scope (repo.ScopeTenant) consumes that
// principal directly, so this middleware does not transform it; it confirms the
// principal is in context and hands it through unchanged. Handlers recover it
// with PrincipalFromContext and pass it to the repositories they call.
//
// If no principal is present (i.e. Auth did not run, or did not run first),
// TenantScope rejects the request as unauthenticated (401, Req 7.6) rather than
// letting a handler execute without a tenant scope and risk leaking data.
func TenantScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFromContext(r.Context()); !ok {
			WriteError(w, apierr.ErrUnauthenticated)
			return
		}
		next.ServeHTTP(w, r)
	})
}

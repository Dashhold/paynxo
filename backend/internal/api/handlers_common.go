package api

import (
	"net/http"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service"
)

// requirePrincipal recovers the authenticated principal that the Auth
// middleware placed in the request context. The business handlers are mounted
// behind protected(...) (Auth + TenantScope), so a principal is expected; if
// none is present the chain was misconfigured, and the handler fails closed
// with apierr.ErrUnauthenticated (401) rather than acting without a tenant
// scope and risking a data leak.
func requirePrincipal(r *http.Request) (service.Principal, error) {
	p, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return service.Principal{}, apierr.ErrUnauthenticated
	}
	return p, nil
}

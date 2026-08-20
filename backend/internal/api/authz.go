package api

import (
	"net/http"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service"
)

// RequireRole reports whether the principal p is permitted to act in one of the
// allowed roles. It returns nil when p.Role is in allowed, and
// apierr.ErrForbidden (403, Req 7.3) otherwise.
//
// This is the single, reusable authorization primitive for the HTTP layer. The
// entity, ledger, report, and lease handlers added in later tasks call it (or
// the RequireRoles decorator below) to gate access — for example lease
// endpoints pass only service.RoleSuperAdmin (Req 7.4, 15.7), while most
// business endpoints pass service.RoleAdmin and service.RoleSuperAdmin.
//
// Passing no allowed roles denies everyone, which keeps "forgot to specify
// roles" failing closed rather than open.
func RequireRole(p service.Principal, allowed ...string) error {
	for _, role := range allowed {
		if p.Role == role {
			return nil
		}
	}
	return apierr.ErrForbidden
}

// RequireRoles returns a decorator that guards a middleware.HandlerFunc so it
// runs only for principals whose role is in allowed. It is the handler-level
// counterpart to RequireRole and is intended to wrap the business handlers
// mounted on protected route groups.
//
// The guard recovers the authenticated principal that the Auth middleware
// placed in the request context. If none is present (the route was not mounted
// behind Auth) it fails closed with apierr.ErrUnauthenticated (401, Req 7.2);
// if the principal's role is not permitted it returns apierr.ErrForbidden (403,
// Req 7.3). The wrapped handler runs only once the role check passes.
func RequireRoles(allowed ...string) func(middleware.HandlerFunc) middleware.HandlerFunc {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			p, ok := middleware.PrincipalFromContext(r.Context())
			if !ok {
				return apierr.ErrUnauthenticated
			}
			if err := RequireRole(p, allowed...); err != nil {
				return err
			}
			return next(w, r)
		}
	}
}

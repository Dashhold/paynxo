package middleware

import (
	"context"

	"pgcs/backend/internal/service"
)

// contextKey is an unexported type for the keys this package stores in a
// request context. Using a private type (rather than a bare string) guarantees
// the key cannot collide with keys set by other packages (Go context best
// practice).
type contextKey int

const (
	// principalKey identifies the authenticated service.Principal carried in a
	// request context.
	principalKey contextKey = iota
)

// WithPrincipal returns a copy of ctx that carries the authenticated principal.
// The Auth middleware calls this after a Session_Token validates so that
// downstream handlers and the repository layer can recover the principal (and
// its tenant/owner scope) via PrincipalFromContext.
func WithPrincipal(ctx context.Context, p service.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFromContext returns the authenticated principal stored in ctx and
// reports whether one was present. The repository layer derives its tenant and
// owner scope from this principal (see repo.ScopeTenant), so handlers pass the
// recovered principal into the repositories they call. The boolean is false for
// any request that did not pass through the Auth middleware.
func PrincipalFromContext(ctx context.Context) (service.Principal, bool) {
	p, ok := ctx.Value(principalKey).(service.Principal)
	return p, ok
}

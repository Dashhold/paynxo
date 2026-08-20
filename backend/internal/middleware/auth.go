package middleware

import (
	"net/http"
	"strings"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service/auth"
)

// authScheme is the (case-insensitive) Authorization scheme that carries the
// Session_Token.
const authScheme = "Bearer"

// Auth returns middleware that authenticates every request against the
// Auth_Service before the wrapped handler runs.
//
// It extracts the Session_Token from the "Authorization: Bearer <token>"
// header, validates it via svc.Authenticate, and on success places the
// resulting service.Principal into the request context (see WithPrincipal) so
// downstream handlers and the tenant-scoped repository layer can consume it.
//
// Failures are rendered with the shared WriteError helper as the structured
// APIError body (Req 18.1):
//   - missing or empty token        -> 401 unauthenticated (Req 7.2, 7.6)
//   - invalid / expired / revoked    -> 401 unauthenticated (Req 6.6)
//   - expired/suspended/revoked lease -> 403 lease_inactive (Req 14.3, 15.4, 15.6)
//
// The 401/403 distinction is produced by the Auth_Service: a bad token yields
// apierr.ErrUnauthenticated, while a non-active lease yields
// apierr.ErrLeaseInactive. Auth simply propagates whichever typed error it
// receives, so the status code always matches the cause.
func Auth(svc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				WriteError(w, err)
				return
			}

			p, err := svc.Authenticate(token)
			if err != nil {
				WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		})
	}
}

// BearerToken extracts the Session_Token from the request's Authorization
// header using the same parsing rules as the Auth middleware. Handlers that
// must act on the raw token itself — notably logout, which revokes the
// presented token — use this to recover it after Auth has already validated the
// request. A missing or malformed header yields apierr.ErrUnauthenticated.
func BearerToken(r *http.Request) (string, error) {
	return bearerToken(r)
}

// bearerToken extracts the Session_Token from the request's Authorization
// header. A missing header, a non-Bearer scheme, or an empty token value are
// all treated as unauthenticated (401, Req 7.2) without revealing which check
// failed. The scheme match is case-insensitive per RFC 7235.
func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", apierr.ErrUnauthenticated
	}

	scheme, rest, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, authScheme) {
		return "", apierr.ErrUnauthenticated
	}

	token := strings.TrimSpace(rest)
	if token == "" {
		return "", apierr.ErrUnauthenticated
	}
	return token, nil
}

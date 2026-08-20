package api

import (
	"net/http"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/auth"
)

// authHandlers serves the authentication endpoints. It depends only on the
// Auth_Service behavior, keeping the transport layer thin: each handler decodes
// the request, delegates to the service, and encodes the result, propagating
// typed errors for the Error middleware to render.
type authHandlers struct {
	auth auth.AuthService
}

// loginRequest is the POST /api/auth/login body.
type loginRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}

// principalView is the JSON projection of the authenticated principal returned
// by /api/me and embedded in the login response. It deliberately omits nothing
// sensitive: a principal carries only identity and scope, no secrets.
type principalView struct {
	AccountID string `json:"accountId"`
	Role      string `json:"role"`
	TenantID  string `json:"tenantId"`
	OwnerType string `json:"ownerType,omitempty"`
	OwnerID   string `json:"ownerId,omitempty"`
}

// loginResponse is the POST /api/auth/login success body: the issued
// Session_Token plus the principal it identifies, so the client can render the
// current user without a follow-up /api/me call.
type loginResponse struct {
	Token     string        `json:"token"`
	Principal principalView `json:"principal"`
}

// viewPrincipal projects a service.Principal into its JSON view.
func viewPrincipal(p service.Principal) principalView {
	return principalView{
		AccountID: p.AccountID,
		Role:      p.Role,
		TenantID:  p.TenantID,
		OwnerType: p.OwnerType,
		OwnerID:   p.OwnerID,
	}
}

// login handles POST /api/auth/login (public). It verifies the submitted
// credentials and, on success, issues a Session_Token (Req 6.1).
//
// A bad user id or password yields apierr.ErrInvalidCredentials from the
// service, which the Error middleware renders as a generic 401 that does not
// reveal which field was wrong (Req 6.2). A leased Admin with a non-active
// lease yields a 403 lease-inactive error (Req 14.2, 15.4, 15.6). Both are
// returned unchanged for the error model to map.
func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	token, p, err := h.auth.Login(req.UserID, req.Password)
	if err != nil {
		return err
	}

	middleware.WriteJSON(w, http.StatusOK, loginResponse{Token: token, Principal: viewPrincipal(p)})
	return nil
}

// logout handles POST /api/auth/logout (authenticated). It recovers the bearer
// token presented on the request and invalidates it via the Auth_Service so it
// can no longer authenticate subsequent requests (Req 6.7), then responds 204
// No Content.
//
// The route is mounted behind the Auth middleware, so a valid token is already
// guaranteed; recovering it again here lets the service revoke that exact
// token's id.
func (h *authHandlers) logout(w http.ResponseWriter, r *http.Request) error {
	token, err := middleware.BearerToken(r)
	if err != nil {
		return err
	}
	if err := h.auth.Logout(token); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// me handles GET /api/me (authenticated). It returns the current principal —
// role, tenant, and owner scope — taken from the request context where the Auth
// middleware placed it after validating the Session_Token.
func (h *authHandlers) me(w http.ResponseWriter, r *http.Request) error {
	p, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		// Defensive: this handler is only mounted behind Auth, so a principal
		// is expected. Fail closed if the chain was misconfigured.
		return apierr.ErrUnauthenticated
	}
	middleware.WriteJSON(w, http.StatusOK, viewPrincipal(p))
	return nil
}

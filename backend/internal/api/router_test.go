package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// fakeAuth is a stub auth.AuthService whose behavior is configured per test.
type fakeAuth struct {
	loginToken string
	loginPrinc service.Principal
	loginErr   error

	authPrinc service.Principal
	authErr   error

	logoutErr  error
	loggedOut  string
}

func (f *fakeAuth) Login(userID, password string) (string, service.Principal, error) {
	return f.loginToken, f.loginPrinc, f.loginErr
}

func (f *fakeAuth) Authenticate(token string) (service.Principal, error) {
	return f.authPrinc, f.authErr
}

func (f *fakeAuth) Logout(token string) error {
	f.loggedOut = token
	return f.logoutErr
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) apierr.APIError {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var body apierr.APIError
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body
}

func TestUnknownRouteReturnsStructured404(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if body := decodeError(t, rec); body.Code != apierr.CodeNotFound {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeNotFound)
	}
}

func TestUnsupportedMethodReturnsStructured405(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{}})

	// /api/auth/login exists for POST; GET must yield a structured 405.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/login", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if body := decodeError(t, rec); body.Code != codeMethodNotAllowed {
		t.Errorf("code = %q, want %q", body.Code, codeMethodNotAllowed)
	}
	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodPost) {
		t.Errorf("Allow header = %q, want it to list POST", allow)
	}
}

func TestLoginSuccessReturnsTokenAndPrincipal(t *testing.T) {
	want := service.Principal{AccountID: "acc1", Role: service.RoleAdmin, TenantID: "t1"}
	svc := &fakeAuth{loginToken: "tok-123", loginPrinc: want}
	h := NewRouter(Deps{Auth: svc})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"userId":"admin","password":"admin123"}`))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	var resp loginResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Token != "tok-123" {
		t.Errorf("token = %q, want %q", resp.Token, "tok-123")
	}
	if resp.Principal.AccountID != want.AccountID || resp.Principal.Role != want.Role {
		t.Errorf("principal = %+v, want %+v", resp.Principal, want)
	}
}

func TestLoginInvalidCredentialsReturns401(t *testing.T) {
	svc := &fakeAuth{loginErr: apierr.ErrInvalidCredentials}
	h := NewRouter(Deps{Auth: svc})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"userId":"x","password":"y"}`))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if body := decodeError(t, rec); body.Code != apierr.CodeInvalidCredentials {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeInvalidCredentials)
	}
}

func TestMeRequiresAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMeReturnsPrincipalWhenAuthenticated(t *testing.T) {
	want := service.Principal{AccountID: "acc9", Role: service.RoleMerchant, TenantID: "t2", OwnerType: service.RoleMerchant, OwnerID: "m1"}
	svc := &fakeAuth{authPrinc: want}
	h := NewRouter(Deps{Auth: svc})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer good.token")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	var got principalView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.AccountID != want.AccountID || got.OwnerID != want.OwnerID {
		t.Errorf("principal = %+v, want account %q owner %q", got, want.AccountID, want.OwnerID)
	}
}

func TestLogoutRevokesTokenAndReturns204(t *testing.T) {
	svc := &fakeAuth{authPrinc: service.Principal{AccountID: "a", TenantID: "t"}}
	h := NewRouter(Deps{Auth: svc})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer revoke.me")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body %s)", rec.Code, rec.Body.String())
	}
	if svc.loggedOut != "revoke.me" {
		t.Errorf("Logout got token %q, want %q", svc.loggedOut, "revoke.me")
	}
}

// TestRequireRole verifies the reusable role guard used by later handlers.
func TestRequireRole(t *testing.T) {
	admin := service.Principal{Role: service.RoleAdmin}
	if err := RequireRole(admin, service.RoleAdmin, service.RoleSuperAdmin); err != nil {
		t.Errorf("admin should be allowed, got %v", err)
	}
	merchant := service.Principal{Role: service.RoleMerchant}
	if err := RequireRole(merchant, service.RoleAdmin, service.RoleSuperAdmin); err == nil {
		t.Error("merchant should be forbidden, got nil")
	}
	if err := RequireRole(admin); err == nil {
		t.Error("no allowed roles should deny everyone, got nil")
	}
}

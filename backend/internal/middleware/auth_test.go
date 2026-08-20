package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/auth"
)

// fakeAuth is a stub auth.AuthService whose Authenticate returns the configured
// principal/error. Only Authenticate is exercised by the Auth middleware.
type fakeAuth struct {
	principal service.Principal
	err       error
	gotToken  string
}

var _ auth.AuthService = (*fakeAuth)(nil)

func (f *fakeAuth) Login(userID, password string) (string, service.Principal, error) {
	return "", service.Principal{}, nil
}

func (f *fakeAuth) Authenticate(token string) (service.Principal, error) {
	f.gotToken = token
	return f.principal, f.err
}

func (f *fakeAuth) Logout(token string) error { return nil }

// captureHandler records whether it ran and the principal it observed in
// context, so tests can assert the Auth middleware both gates and propagates.
type captureHandler struct {
	called    bool
	principal service.Principal
	hadPrinc  bool
}

func (c *captureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.called = true
	c.principal, c.hadPrinc = PrincipalFromContext(r.Context())
	w.WriteHeader(http.StatusOK)
}

func TestAuthRejectsMissingHeader(t *testing.T) {
	svc := &fakeAuth{}
	next := &captureHandler{}
	h := Auth(svc)(next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if next.called {
		t.Error("next handler ran despite missing token")
	}
	if decodeBody(t, rec).Code != apierr.CodeUnauthenticated {
		t.Errorf("code = %q, want %q", decodeBody(t, rec).Code, apierr.CodeUnauthenticated)
	}
}

func TestAuthRejectsNonBearerScheme(t *testing.T) {
	svc := &fakeAuth{}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if next.called {
		t.Error("next handler ran for non-Bearer scheme")
	}
}

func TestAuthRejectsEmptyBearerToken(t *testing.T) {
	svc := &fakeAuth{}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer    ")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if next.called {
		t.Error("next handler ran for empty bearer token")
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	svc := &fakeAuth{err: apierr.ErrUnauthenticated}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer bad.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if next.called {
		t.Error("next handler ran for invalid token")
	}
}

func TestAuthInactiveLeaseReturns403(t *testing.T) {
	svc := &fakeAuth{err: apierr.LeaseInactive("lease has expired")}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer valid.but.expired.lease")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if next.called {
		t.Error("next handler ran for inactive lease")
	}
	if body := decodeBody(t, rec); body.Code != apierr.CodeLeaseInactive {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeLeaseInactive)
	}
}

func TestAuthSuccessPlacesPrincipalInContext(t *testing.T) {
	want := service.Principal{
		AccountID: "acc1",
		Role:      service.RoleMerchant,
		TenantID:  "tenant-demo",
		OwnerType: service.RoleMerchant,
		OwnerID:   "m1",
	}
	svc := &fakeAuth{principal: want}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/merchants", nil)
	req.Header.Set("Authorization", "Bearer good.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !next.called {
		t.Fatal("next handler did not run on success")
	}
	if svc.gotToken != "good.token" {
		t.Errorf("Authenticate got token %q, want %q", svc.gotToken, "good.token")
	}
	if !next.hadPrinc {
		t.Fatal("principal not present in downstream context")
	}
	if next.principal != want {
		t.Errorf("principal = %+v, want %+v", next.principal, want)
	}
}

// TestAuthBearerSchemeCaseInsensitive verifies the scheme match tolerates
// alternate casing per RFC 7235.
func TestAuthBearerSchemeCaseInsensitive(t *testing.T) {
	svc := &fakeAuth{principal: service.Principal{AccountID: "a", TenantID: "t"}}
	next := &captureHandler{}
	h := Auth(svc)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "bearer good.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.gotToken != "good.token" {
		t.Errorf("Authenticate got token %q, want %q", svc.gotToken, "good.token")
	}
}

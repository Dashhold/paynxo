package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

func TestTenantScopeRejectsWhenNoPrincipal(t *testing.T) {
	next := &captureHandler{}
	h := TenantScope(next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/merchants", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if next.called {
		t.Error("next handler ran without a principal in context")
	}
	if body := decodeBody(t, rec); body.Code != apierr.CodeUnauthenticated {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeUnauthenticated)
	}
}

func TestTenantScopePassesThroughWithPrincipal(t *testing.T) {
	want := service.Principal{AccountID: "a1", Role: service.RoleAdmin, TenantID: "t1"}
	next := &captureHandler{}
	h := TenantScope(next)

	req := httptest.NewRequest(http.MethodGet, "/api/merchants", nil)
	req = req.WithContext(WithPrincipal(req.Context(), want))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !next.called {
		t.Fatal("next handler did not run with a principal present")
	}
	if !next.hadPrinc || next.principal != want {
		t.Errorf("downstream principal = %+v (present=%v), want %+v", next.principal, next.hadPrinc, want)
	}
}

// TestAuthThenTenantScopeChain exercises the intended mounting order: Auth sets
// the principal, TenantScope confirms it, and the handler observes it.
func TestAuthThenTenantScopeChain(t *testing.T) {
	want := service.Principal{AccountID: "a1", Role: service.RoleCompany, TenantID: "t1", OwnerType: service.RoleCompany, OwnerID: "co1"}
	svc := &fakeAuth{principal: want}
	next := &captureHandler{}
	h := Auth(svc)(TenantScope(next))

	req := httptest.NewRequest(http.MethodGet, "/api/companies", nil)
	req.Header.Set("Authorization", "Bearer good.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !next.hadPrinc || next.principal != want {
		t.Errorf("downstream principal = %+v (present=%v), want %+v", next.principal, next.hadPrinc, want)
	}
}

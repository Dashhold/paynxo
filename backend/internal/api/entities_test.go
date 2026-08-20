package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// The core-entity routes are mounted behind protected(...) + RequireRoles(
// Admin, SuperAdmin). These tests verify the guard rejects unauthenticated and
// unauthorized callers before any handler (and thus any database access) runs,
// so they need no database.

// coreEntityRoutes is the set of (method, path) pairs registered by task 7.1.
var coreEntityRoutes = []struct {
	method, path string
}{
	{http.MethodGet, "/api/gateways"},
	{http.MethodPost, "/api/gateways"},
	{http.MethodGet, "/api/gateways/gw1"},
	{http.MethodPut, "/api/gateways/gw1"},
	{http.MethodDelete, "/api/gateways/gw1"},
	{http.MethodGet, "/api/companies"},
	{http.MethodPost, "/api/companies"},
	{http.MethodGet, "/api/companies/co1"},
	{http.MethodPut, "/api/companies/co1"},
	{http.MethodDelete, "/api/companies/co1"},
	{http.MethodGet, "/api/affiliates"},
	{http.MethodPost, "/api/affiliates"},
	{http.MethodGet, "/api/affiliates/af1"},
	{http.MethodPut, "/api/affiliates/af1"},
	{http.MethodDelete, "/api/affiliates/af1"},
	{http.MethodGet, "/api/merchants"},
	{http.MethodPost, "/api/merchants"},
	{http.MethodGet, "/api/merchants/m1"},
	{http.MethodPut, "/api/merchants/m1"},
	{http.MethodDelete, "/api/merchants/m1"},
}

func TestCoreEntityRoutesRequireAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{authErr: apierr.ErrUnauthenticated}})

	for _, rt := range coreEntityRoutes {
		rec := httptest.NewRecorder()
		// No Authorization header: the Auth middleware rejects with 401 before
		// the handler runs.
		h.ServeHTTP(rec, httptest.NewRequest(rt.method, rt.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401", rt.method, rt.path, rec.Code)
		}
	}
}

func TestCoreEntityRoutesForbidNonAdminRoles(t *testing.T) {
	// A Merchant principal is authenticated but not permitted on these
	// Admin/SuperAdmin endpoints, so RequireRoles rejects with 403 before the
	// handler touches the database.
	svc := &fakeAuth{authPrinc: service.Principal{
		AccountID: "acc1", Role: service.RoleMerchant, TenantID: "t1",
		OwnerType: service.RoleMerchant, OwnerID: "m1",
	}}
	h := NewRouter(Deps{Auth: svc})

	for _, rt := range coreEntityRoutes {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(rt.method, rt.path, nil)
		req.Header.Set("Authorization", "Bearer good.token")
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s: status = %d, want 403", rt.method, rt.path, rec.Code)
		}
		if body := decodeError(t, rec); body.Code != apierr.CodeForbidden {
			t.Errorf("%s %s: code = %q, want %q", rt.method, rt.path, body.Code, apierr.CodeForbidden)
		}
	}
}

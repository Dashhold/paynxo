package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// transactionAndPaymentRoutes is the set of (method, path) pairs registered by
// task 7.3 (settlements, affiliate/merchant payments) plus the transaction
// CRUD/breakdown wiring. Like the core-entity routes they are mounted behind
// protected(...) + RequireRoles(Admin, SuperAdmin), so the guard tests need no
// database: the chain rejects unauthenticated and unauthorized callers before
// any handler runs.
var transactionAndPaymentRoutes = []struct {
	method, path string
}{
	{http.MethodGet, "/api/transactions"},
	{http.MethodPost, "/api/transactions"},
	{http.MethodGet, "/api/transactions/tx1"},
	{http.MethodPut, "/api/transactions/tx1"},
	{http.MethodDelete, "/api/transactions/tx1"},
	{http.MethodGet, "/api/settlements"},
	{http.MethodPost, "/api/settlements"},
	{http.MethodGet, "/api/settlements/s1"},
	{http.MethodPut, "/api/settlements/s1"},
	{http.MethodDelete, "/api/settlements/s1"},
	{http.MethodGet, "/api/affiliate-payments"},
	{http.MethodPost, "/api/affiliate-payments"},
	{http.MethodGet, "/api/affiliate-payments/ap1"},
	{http.MethodPut, "/api/affiliate-payments/ap1"},
	{http.MethodDelete, "/api/affiliate-payments/ap1"},
	{http.MethodGet, "/api/merchant-payments"},
	{http.MethodPost, "/api/merchant-payments"},
	{http.MethodGet, "/api/merchant-payments/mp1"},
	{http.MethodPut, "/api/merchant-payments/mp1"},
	{http.MethodDelete, "/api/merchant-payments/mp1"},
}

func TestTransactionAndPaymentRoutesRequireAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{authErr: apierr.ErrUnauthenticated}})

	for _, rt := range transactionAndPaymentRoutes {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(rt.method, rt.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401", rt.method, rt.path, rec.Code)
		}
	}
}

func TestTransactionAndPaymentRoutesForbidNonAdminRoles(t *testing.T) {
	svc := &fakeAuth{authPrinc: service.Principal{
		AccountID: "acc1", Role: service.RoleMerchant, TenantID: "t1",
		OwnerType: service.RoleMerchant, OwnerID: "m1",
	}}
	h := NewRouter(Deps{Auth: svc})

	for _, rt := range transactionAndPaymentRoutes {
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

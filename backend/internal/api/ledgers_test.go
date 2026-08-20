package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// The ledger routes are mounted behind protected(...) + RequireRoles(Admin,
// SuperAdmin, <matching portal role>). These tests verify the guard rejects
// unauthenticated callers and portal principals reaching a ledger type that is
// not theirs, before any handler (and thus any database access) runs — so they
// need no database, matching the other api route-guard tests.

// ledgerRoutes is the set of (method, path) pairs registered by task 10.1.
var ledgerRoutes = []struct {
	method, path string
}{
	{http.MethodGet, "/api/ledgers/company/co1"},
	{http.MethodGet, "/api/ledgers/affiliate/af1"},
	{http.MethodGet, "/api/ledgers/merchant/m1"},
}

func TestLedgerRoutesRequireAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{authErr: apierr.ErrUnauthenticated}})

	for _, rt := range ledgerRoutes {
		rec := httptest.NewRecorder()
		// No Authorization header: the Auth middleware rejects with 401 before
		// the handler runs.
		h.ServeHTTP(rec, httptest.NewRequest(rt.method, rt.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401", rt.method, rt.path, rec.Code)
		}
	}
}

// TestLedgerRoutesRestrictPortalRolesToTheirOwnLedger verifies a portal
// principal can only reach the ledger type that applies to its role: a Company
// principal is forbidden on the affiliate and merchant ledgers, an Affiliate on
// the company and merchant ledgers, and a Merchant on the company and affiliate
// ledgers (Req 7.3, 7.5). Each rejection is a 403 raised by RequireRoles before
// any database access.
func TestLedgerRoutesRestrictPortalRolesToTheirOwnLedger(t *testing.T) {
	cases := []struct {
		name, role, path string
	}{
		{"company-on-affiliate", service.RoleCompany, "/api/ledgers/affiliate/af1"},
		{"company-on-merchant", service.RoleCompany, "/api/ledgers/merchant/m1"},
		{"affiliate-on-company", service.RoleAffiliate, "/api/ledgers/company/co1"},
		{"affiliate-on-merchant", service.RoleAffiliate, "/api/ledgers/merchant/m1"},
		{"merchant-on-company", service.RoleMerchant, "/api/ledgers/company/co1"},
		{"merchant-on-affiliate", service.RoleMerchant, "/api/ledgers/affiliate/af1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeAuth{authPrinc: service.Principal{
				AccountID: "acc1", Role: tc.role, TenantID: "t1",
				OwnerType: tc.role, OwnerID: "own1",
			}}
			h := NewRouter(Deps{Auth: svc})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer good.token")
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", rec.Code)
			}
			if body := decodeError(t, rec); body.Code != apierr.CodeForbidden {
				t.Errorf("code = %q, want %q", body.Code, apierr.CodeForbidden)
			}
		})
	}
}

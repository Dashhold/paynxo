package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// The report route is mounted behind protected(...) + RequireRoles(Admin,
// SuperAdmin). These guard tests verify the route rejects unauthenticated
// callers and non-admin portal principals before any handler (and thus any
// database access) runs, so they need no database — matching the other api
// route-guard tests.

func TestReportRouteRequiresAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{authErr: apierr.ErrUnauthenticated}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/reports/company", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestReportRouteForbidsPortalRoles(t *testing.T) {
	for _, role := range []string{service.RoleCompany, service.RoleAffiliate, service.RoleMerchant} {
		t.Run(role, func(t *testing.T) {
			svc := &fakeAuth{authPrinc: service.Principal{
				AccountID: "acc1", Role: role, TenantID: "t1",
				OwnerType: role, OwnerID: "own1",
			}}
			h := NewRouter(Deps{Auth: svc})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/reports/company", nil)
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

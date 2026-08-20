package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// leaseRoutes is the set of SuperAdmin-only lease endpoints. The guard runs
// before the handler, so these route-guard tests never reach the (nil-DB)
// Lease_Manager.
var leaseRoutes = []struct {
	method string
	path   string
}{
	{http.MethodGet, "/api/leases"},
	{http.MethodPost, "/api/leases"},
	{http.MethodPost, "/api/leases/lease1/extend"},
	{http.MethodPost, "/api/leases/lease1/suspend"},
	{http.MethodPost, "/api/leases/lease1/reactivate"},
	{http.MethodPost, "/api/leases/lease1/revoke"},
}

// TestLeaseRoutesRequireAuth verifies that without a valid Session_Token every
// lease endpoint is rejected with 401 by the Auth middleware (Req 7.2).
func TestLeaseRoutesRequireAuth(t *testing.T) {
	h := NewRouter(Deps{Auth: &fakeAuth{}})

	for _, rt := range leaseRoutes {
		rec := httptest.NewRecorder()
		// No Authorization header -> Auth middleware fails closed with 401.
		h.ServeHTTP(rec, httptest.NewRequest(rt.method, rt.path, nil))

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401 (body %s)", rt.method, rt.path, rec.Code, rec.Body.String())
		}
	}
}

// TestLeaseRoutesForbiddenForNonSuperAdmin verifies that an authenticated
// non-SuperAdmin principal (here an Admin) is denied every lease endpoint with
// 403 (Req 7.4, 15.7).
func TestLeaseRoutesForbiddenForNonSuperAdmin(t *testing.T) {
	admin := service.Principal{AccountID: "a1", Role: service.RoleAdmin, TenantID: "t1"}
	h := NewRouter(Deps{Auth: &fakeAuth{authPrinc: admin}})

	for _, rt := range leaseRoutes {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(rt.method, rt.path, strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer good.token")
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s: status = %d, want 403 (body %s)", rt.method, rt.path, rec.Code, rec.Body.String())
		}
		if body := decodeError(t, rec); body.Code != apierr.CodeForbidden {
			t.Errorf("%s %s: code = %q, want %q", rt.method, rt.path, body.Code, apierr.CodeForbidden)
		}
	}
}

// TestParseLeaseDate covers the accepted date formats and the rejection of a
// malformed value, independent of the HTTP layer.
func TestParseLeaseDate(t *testing.T) {
	// Empty resolves to the zero time so manager validation reports it.
	if got, err := parseLeaseDate("", "startDate"); err != nil || !got.IsZero() {
		t.Errorf("empty: got (%v, %v), want (zero, nil)", got, err)
	}

	// yyyy-mm-dd is accepted.
	if got, err := parseLeaseDate("2025-01-31", "startDate"); err != nil || got.IsZero() {
		t.Errorf("yyyy-mm-dd: got (%v, %v), want a parsed time", got, err)
	}

	// RFC3339 is accepted.
	if got, err := parseLeaseDate("2025-01-31T00:00:00Z", "startDate"); err != nil || got.IsZero() {
		t.Errorf("RFC3339: got (%v, %v), want a parsed time", got, err)
	}

	// A malformed value is a 400 validation error.
	if _, err := parseLeaseDate("31-01-2025", "startDate"); err == nil {
		t.Error("malformed date: got nil error, want a validation error")
	}
}

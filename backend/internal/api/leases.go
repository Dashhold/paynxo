package api

import (
	"net/http"
	"strings"
	"time"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/lease"
)

// leaseDateLayout is the ISO yyyy-mm-dd layout accepted for the startDate and
// expiryDate fields, matching the date style used elsewhere in the API (see
// reports.go). For convenience the lease handlers also accept a full RFC3339
// timestamp, so a client may send either "2025-01-31" or
// "2025-01-31T00:00:00Z"; both resolve to the same instant.
const leaseDateLayout = "2006-01-02"

// leaseHandlers serves the SuperAdmin-only lease administration endpoints
// (Req 13, 15). Each method is thin: it decodes the request, delegates to the
// Lease_Manager, and encodes the result, propagating the manager's typed errors
// for the Error middleware to render (400 validation, 409 duplicate, 404 not
// found). Authorization to SuperAdmin is enforced by the route guard, not here
// (Req 7.4, 15.7).
type leaseHandlers struct {
	mgr lease.LeaseManager
}

// newLeaseHandlers builds the lease handlers over a database handle, mirroring
// how the report and CRUD handlers construct their service from deps.DB.
func newLeaseHandlers(deps Deps) *leaseHandlers {
	return &leaseHandlers{mgr: lease.NewManager(deps.DB)}
}

// list handles GET /api/leases: every lease with its identity, tenure, and
// effective Lease_Status (Req 15.1). The SuperAdmin sees all leases, including
// those whose effective status is Expired (Req 14.5).
func (h *leaseHandlers) list(w http.ResponseWriter, r *http.Request) error {
	views, err := h.mgr.List()
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, views)
	return nil
}

// create handles POST /api/leases. It decodes the new Admin's user id, password,
// and tenure bounds, then provisions a new Tenant, leased Admin Account, and
// Active Lease (Req 13.1, 13.3). A field validation failure (including an
// expiry not after the start) yields 400 (Req 13.4, 13.6); a duplicate user id
// yields 409 (Req 13.5). On success it returns 201 with the created lease view.
func (h *leaseHandlers) create(w http.ResponseWriter, r *http.Request) error {
	var req createLeaseRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	start, err := parseLeaseDate(req.StartDate, "startDate")
	if err != nil {
		return err
	}
	expiry, err := parseLeaseDate(req.ExpiryDate, "expiryDate")
	if err != nil {
		return err
	}

	// Zero-valued dates are passed through so the manager's field validation
	// reports the missing field (Req 13.4) with the rest of the input.
	created, err := h.mgr.Create(lease.CreateLeaseInput{
		AdminUserID: req.AdminUserID,
		AdminName:   req.AdminName,
		Password:    req.Password,
		StartDate:   start,
		ExpiryDate:  expiry,
	})
	if err != nil {
		return err
	}

	middleware.WriteJSON(w, http.StatusCreated, toLeaseView(created))
	return nil
}

// extend handles POST /api/leases/{id}/extend. It decodes the new expiry date
// and moves the lease's tenure forward, re-activating it (Req 15.2). A missing
// or malformed expiry yields 400; an unknown id yields 404.
func (h *leaseHandlers) extend(w http.ResponseWriter, r *http.Request) error {
	var req extendLeaseRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	newExpiry, err := parseLeaseDate(req.ExpiryDate, "expiryDate")
	if err != nil {
		return err
	}
	if newExpiry.IsZero() {
		return apierr.ValidationField("expiryDate", "expiry date is required")
	}

	updated, err := h.mgr.Extend(r.PathValue("id"), newExpiry)
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, toLeaseView(updated))
	return nil
}

// suspend handles POST /api/leases/{id}/suspend, pausing access for the leased
// Admin while preserving the tenure (Req 15.3). An unknown id yields 404.
func (h *leaseHandlers) suspend(w http.ResponseWriter, r *http.Request) error {
	updated, err := h.mgr.Suspend(r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, toLeaseView(updated))
	return nil
}

// reactivate handles POST /api/leases/{id}/reactivate, clearing a suspension
// (Req 15.5). If the tenure has since lapsed the lease resolves back to Expired
// via the effective-status derivation. An unknown id yields 404.
func (h *leaseHandlers) reactivate(w http.ResponseWriter, r *http.Request) error {
	updated, err := h.mgr.Reactivate(r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, toLeaseView(updated))
	return nil
}

// revoke handles POST /api/leases/{id}/revoke, permanently ending the lease and
// denying all access for the associated leased Admin (Req 15.6). An unknown id
// yields 404.
func (h *leaseHandlers) revoke(w http.ResponseWriter, r *http.Request) error {
	updated, err := h.mgr.Revoke(r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, toLeaseView(updated))
	return nil
}

// createLeaseRequest is the POST /api/leases body: the new Admin's credentials
// and tenure bounds. Dates are yyyy-mm-dd (or RFC3339); see leaseDateLayout.
type createLeaseRequest struct {
	AdminUserID string `json:"adminUserId"`
	AdminName   string `json:"adminName"`
	Password    string `json:"password"`
	StartDate   string `json:"startDate"`
	ExpiryDate  string `json:"expiryDate"`
}

// extendLeaseRequest is the POST /api/leases/{id}/extend body: the new expiry
// date the tenure is moved to.
type extendLeaseRequest struct {
	ExpiryDate string `json:"expiryDate"`
}

// toLeaseView projects a stored model.Lease into the read view returned by the
// mutation endpoints, resolving the effective Lease_Status at the current time
// so create/extend/suspend/reactivate/revoke responses match the shape of the
// list endpoint (Req 15.1).
func toLeaseView(l model.Lease) lease.LeaseView {
	return lease.LeaseView{
		ID:          l.ID,
		AdminUserID: l.AdminUserID,
		AdminName:   l.AdminName,
		TenantID:    l.TenantID,
		AccountID:   l.AccountID,
		StartDate:   l.StartDate,
		ExpiryDate:  l.ExpiryDate,
		Status:      lease.EffectiveStatus(l, time.Now()),
	}
}

// parseLeaseDate parses a date field accepted as either yyyy-mm-dd or RFC3339.
// An empty value resolves to the zero time so the Lease_Manager's field
// validation can report it as a missing required field with the rest of the
// input (Req 13.4); a non-empty but malformed value is a 400 validation error
// identifying the offending field (Req 18.3).
func parseLeaseDate(raw, field string) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(leaseDateLayout, raw); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Time{}, apierr.ValidationField(field, "must be a date in yyyy-mm-dd or RFC3339 format")
}

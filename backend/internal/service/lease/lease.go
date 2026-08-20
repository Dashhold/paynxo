// Package lease implements the Lease_Manager: the multi-tenant leasing
// capability by which a SuperAdmin grants an Admin an isolated Tenant for a
// defined tenure.
//
// This file defines the lease status state machine — the canonical derivation
// of a lease's *effective* status. The database stores only the administrative
// intent (Active | Suspended | Revoked) on model.Lease.Status; the Expired
// status is never stored. Instead it is derived at read time by comparing the
// current time against the lease's ExpiryDate, so tenure enforcement needs no
// scheduled job (Req 14.1).
//
// The lease create/list/extend/suspend/reactivate/revoke operations are
// implemented in a later task and build on the status model defined here.
//
// State diagram (design "Lease_Manager"):
//
//	[*]       --> Active:    Create (expiry > start)
//	Active    --> Expired:   now > expiry (derived)
//	Active    --> Suspended: Suspend
//	Suspended --> Active:    Reactivate (expiry in future)
//	Active    --> Revoked:   Revoke
//	Suspended --> Revoked:   Revoke
//	Expired   --> Active:    Extend (newExpiry > now)
//	Expired   --> Revoked:   Revoke
//	Revoked   --> [*]:       terminal
package lease

import (
	"time"

	"pgcs/backend/internal/model"
)

// LeaseStatus is the effective state of a lease as seen by access control and
// lease administration (Lease_Status in the requirements).
type LeaseStatus string

// The four lease statuses. Active, Suspended, and Revoked are also the values
// stored as administrative intent on model.Lease.Status; Expired is only ever
// derived (see EffectiveStatus) and is never persisted.
const (
	// Active means the lease is in force: within tenure and neither suspended
	// nor revoked. The leased Admin may authenticate and operate.
	Active LeaseStatus = "Active"
	// Expired is derived when the current time is past the lease's ExpiryDate
	// (Req 14.1). It is never stored.
	Expired LeaseStatus = "Expired"
	// Suspended means a SuperAdmin has paused the lease (Req 15.3). Suspension
	// overrides expiry for access purposes (Req 15.4).
	Suspended LeaseStatus = "Suspended"
	// Revoked is the terminal state: a SuperAdmin has permanently ended the
	// lease (Req 15.6). It overrides every other status.
	Revoked LeaseStatus = "Revoked"
)

// EffectiveStatus resolves the visible status of a lease at time now from its
// stored administrative intent (l.Status) and its ExpiryDate.
//
// Resolution precedence (design "Lease_Manager", status model):
//  1. Revoked is terminal — a revoked lease always resolves to Revoked
//     regardless of expiry (Req 15.6).
//  2. Suspended resolves to Suspended; suspension overrides expiry for access
//     purposes (Req 15.3, 15.4).
//  3. Otherwise, if now is after ExpiryDate the lease is Expired (Req 14.1).
//  4. Otherwise the lease is Active.
//
// An unrecognized stored Status is treated as administrative intent to be
// active, so it still falls through to the Expired/Active expiry check rather
// than silently granting or denying access on a bad value.
func EffectiveStatus(l model.Lease, now time.Time) LeaseStatus {
	switch LeaseStatus(l.Status) {
	case Revoked:
		return Revoked
	case Suspended:
		return Suspended
	}
	if now.After(l.ExpiryDate) {
		return Expired
	}
	return Active
}

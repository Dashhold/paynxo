package lease

import (
	"testing"
	"time"

	"pgcs/backend/internal/model"
)

// fixed reference time and a lease expiring one hour later, used to exercise
// the before/after-expiry branches deterministically.
var (
	now    = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	future = now.Add(time.Hour)
	past   = now.Add(-time.Hour)
)

func leaseWith(status string, expiry time.Time) model.Lease {
	return model.Lease{Status: status, ExpiryDate: expiry}
}

func TestEffectiveStatusPrecedence(t *testing.T) {
	tests := []struct {
		name  string
		lease model.Lease
		want  LeaseStatus
	}{
		// Revoked is terminal: it wins even when expiry is still in the future.
		{"revoked overrides active window", leaseWith("Revoked", future), Revoked},
		// ...and even when the lease is also past expiry.
		{"revoked overrides expiry", leaseWith("Revoked", past), Revoked},
		// Suspended overrides expiry for access purposes.
		{"suspended within tenure", leaseWith("Suspended", future), Suspended},
		{"suspended overrides expiry", leaseWith("Suspended", past), Suspended},
		// Active intent: derived from expiry.
		{"active within tenure", leaseWith("Active", future), Active},
		{"active past expiry derives expired", leaseWith("Active", past), Expired},
		// Unknown stored intent falls through to the expiry check.
		{"unknown intent within tenure", leaseWith("", future), Active},
		{"unknown intent past expiry", leaseWith("", past), Expired},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EffectiveStatus(tc.lease, now); got != tc.want {
				t.Errorf("EffectiveStatus(%+v, now) = %q, want %q", tc.lease, got, tc.want)
			}
		})
	}
}

// Expiry is strict: a lease whose ExpiryDate equals now is not yet expired
// (now must be *after* expiry per Req 14.1).
func TestEffectiveStatusExpiryBoundary(t *testing.T) {
	if got := EffectiveStatus(leaseWith("Active", now), now); got != Active {
		t.Errorf("at exact expiry instant = %q, want %q", got, Active)
	}
}

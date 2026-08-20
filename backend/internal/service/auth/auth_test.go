package auth

import (
	"testing"
	"time"

	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
)

// newTestService builds a Service without a database for tests that exercise
// only the stateless token/lease-derivation helpers.
func newTestService(secret string, ttl time.Duration, now time.Time) *Service {
	return &Service{
		secret: []byte(secret),
		ttl:    ttl,
		now:    func() time.Time { return now },
	}
}

func TestIssueAndParseRoundTrip(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := newTestService("test-secret", time.Hour, now)

	p := service.Principal{
		AccountID: "acc1",
		Role:      service.RoleMerchant,
		TenantID:  "ten1",
		OwnerType: "Merchant",
		OwnerID:   "m1",
	}
	token, err := s.issueToken(p)
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}

	claims, err := s.parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.AccountID != p.AccountID || claims.Role != p.Role ||
		claims.TenantID != p.TenantID || claims.OwnerType != p.OwnerType ||
		claims.OwnerID != p.OwnerID {
		t.Fatalf("claims mismatch: %+v vs %+v", claims, p)
	}
	if claims.ID == "" {
		t.Fatalf("expected a non-empty jti")
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected expiry: %v", claims.ExpiresAt)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	issuer := newTestService("secret-a", time.Hour, now)
	token, err := issuer.issueToken(service.Principal{AccountID: "a", Role: service.RoleAdmin, TenantID: "t"})
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}

	verifier := newTestService("secret-b", time.Hour, now)
	if _, err := verifier.parse(token); err == nil {
		t.Fatalf("expected parse to reject a token signed with a different secret")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	issued := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := newTestService("test-secret", time.Minute, issued)
	token, err := s.issueToken(service.Principal{AccountID: "a", Role: service.RoleAdmin, TenantID: "t"})
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}

	// Advance the clock well past the token's lifetime.
	s.now = func() time.Time { return issued.Add(time.Hour) }
	if _, err := s.parse(token); err == nil {
		t.Fatalf("expected parse to reject an expired token")
	}
}

func TestEffectiveLeaseStatusPrecedence(t *testing.T) {
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name   string
		status string
		expiry time.Time
		want   string
	}{
		{"revoked overrides everything", leaseRevoked, future, leaseRevoked},
		{"suspended overrides expiry", leaseSuspended, past, leaseSuspended},
		{"active but past expiry is expired", leaseActive, past, leaseExpired},
		{"active and within tenure", leaseActive, future, leaseActive},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := effectiveLeaseStatus(model.Lease{Status: c.status, ExpiryDate: c.expiry}, now)
			if got != c.want {
				t.Fatalf("effectiveLeaseStatus = %q, want %q", got, c.want)
			}
		})
	}
}

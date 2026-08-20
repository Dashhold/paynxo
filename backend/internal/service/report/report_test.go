package report

import (
	"errors"
	"testing"
	"time"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

func mustDate(t *testing.T, s string) *time.Time {
	t.Helper()
	d, err := time.Parse(dateLayout, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return &d
}

// TestInRange covers the inclusive [start, end] date filtering (Req 11.2),
// including open-ended bounds and the inclusive endpoints.
func TestInRange(t *testing.T) {
	start := mustDate(t, "2024-01-10")
	end := mustDate(t, "2024-01-20")

	cases := []struct {
		name        string
		date        string
		start, end  *time.Time
		wantInRange bool
	}{
		{"before start", "2024-01-09", start, end, false},
		{"on start (inclusive)", "2024-01-10", start, end, true},
		{"within", "2024-01-15", start, end, true},
		{"on end (inclusive)", "2024-01-20", start, end, true},
		{"after end", "2024-01-21", start, end, false},
		{"no bounds", "2024-01-15", nil, nil, true},
		{"only start, after", "2024-02-01", start, nil, true},
		{"only start, before", "2024-01-01", start, nil, false},
		{"only end, before", "2024-01-01", nil, end, true},
		{"only end, after", "2024-02-01", nil, end, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inRange(tc.date, tc.start, tc.end); got != tc.wantInRange {
				t.Errorf("inRange(%q) = %v, want %v", tc.date, got, tc.wantInRange)
			}
		})
	}
}

// TestGenerateUnknownTypeIsValidationError verifies an unrecognized report type
// is reported as a 400 validation error naming the offending field, without
// touching the database (the default branch runs before any load).
func TestGenerateUnknownTypeIsValidationError(t *testing.T) {
	s := NewService(nil)
	_, err := s.Generate(service.Principal{Role: service.RoleAdmin, TenantID: "t1"},
		ReportRequest{Type: ReportType("bogus")})
	if err == nil {
		t.Fatal("expected an error for an unknown report type")
	}
	var se *apierr.ServiceError
	if !errors.As(err, &se) || se.Kind != apierr.KindValidation {
		t.Fatalf("error = %v, want a validation ServiceError", err)
	}
	if _, ok := se.Fields["type"]; !ok {
		t.Errorf("validation fields = %v, want a \"type\" entry", se.Fields)
	}
}

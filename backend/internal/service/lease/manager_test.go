package lease

import (
	"errors"
	"testing"
	"time"

	"pgcs/backend/internal/apierr"
)

// validInput is a well-formed create request used as a baseline; tests mutate
// individual fields to exercise the validation rules.
func validInput() CreateLeaseInput {
	return CreateLeaseInput{
		AdminUserID: "leased-admin",
		Password:    "s3cret-pass",
		StartDate:   now,
		ExpiryDate:  future,
	}
}

func isValidationErr(t *testing.T, err error) *apierr.ServiceError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a validation error, got nil")
	}
	if !errors.Is(err, apierr.ErrValidation) {
		t.Fatalf("expected a validation (400) error, got %v", err)
	}
	var se *apierr.ServiceError
	if !errors.As(err, &se) {
		t.Fatalf("expected *apierr.ServiceError, got %T", err)
	}
	return se
}

func TestValidateCreateInputAcceptsWellFormed(t *testing.T) {
	if err := validateCreateInput(validInput()); err != nil {
		t.Fatalf("validateCreateInput(valid) = %v, want nil", err)
	}
}

func TestValidateCreateInputRequiredFields(t *testing.T) {
	tests := []struct {
		name  string
		mutate func(*CreateLeaseInput)
		field string
	}{
		{"missing user id", func(in *CreateLeaseInput) { in.AdminUserID = "  " }, "adminUserId"},
		{"missing password", func(in *CreateLeaseInput) { in.Password = "" }, "password"},
		{"missing start date", func(in *CreateLeaseInput) { in.StartDate = time.Time{} }, "startDate"},
		{"missing expiry date", func(in *CreateLeaseInput) { in.ExpiryDate = time.Time{} }, "expiryDate"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mutate(&in)
			se := isValidationErr(t, validateCreateInput(in))
			if _, ok := se.Fields[tc.field]; !ok {
				t.Errorf("expected offending field %q, got fields %v", tc.field, se.Fields)
			}
		})
	}
}

// Req 13.4: an expiry that is not strictly after the start date is a 400.
func TestValidateCreateInputExpiryNotAfterStart(t *testing.T) {
	cases := map[string]time.Time{
		"expiry equals start":  now,
		"expiry before start":  past,
	}
	for name, expiry := range cases {
		t.Run(name, func(t *testing.T) {
			in := validInput()
			in.ExpiryDate = expiry
			se := isValidationErr(t, validateCreateInput(in))
			if _, ok := se.Fields["expiryDate"]; !ok {
				t.Errorf("expected expiryDate field error, got %v", se.Fields)
			}
		})
	}
}

// Req 13.6: validation precedence. Create runs field validation before the
// duplicate-user-id lookup, so a malformed request returns a 400 without ever
// touching the database. A nil *gorm.DB proves the DB path was not reached:
// if validation did not short-circuit, the duplicate Count query would panic.
func TestCreateValidationPrecedesDuplicateCheck(t *testing.T) {
	m := &Manager{db: nil, now: func() time.Time { return now }}

	in := validInput()
	in.ExpiryDate = past // invalid tenure -> validation failure

	_, err := m.Create(in)
	se := isValidationErr(t, err)
	if _, ok := se.Fields["expiryDate"]; !ok {
		t.Errorf("expected expiryDate validation error, got %v", se.Fields)
	}
}

// Req 15.2: an extension must move the expiry strictly forward.
func TestValidateExtend(t *testing.T) {
	if err := validateExtend(now, future); err != nil {
		t.Errorf("validateExtend(now, future) = %v, want nil", err)
	}
	for name, newExpiry := range map[string]time.Time{
		"same expiry":    now,
		"earlier expiry": past,
	} {
		t.Run(name, func(t *testing.T) {
			err := validateExtend(now, newExpiry)
			se := isValidationErr(t, err)
			if _, ok := se.Fields["expiryDate"]; !ok {
				t.Errorf("expected expiryDate validation error, got %v", se.Fields)
			}
		})
	}
}

// The status-transition helpers must map to the documented stored intent so
// the effective-status derivation in EffectiveStatus resolves correctly.
func TestTransitionStatusConstants(t *testing.T) {
	if Suspended != "Suspended" || Active != "Active" || Revoked != "Revoked" {
		t.Fatalf("unexpected status constants: %q %q %q", Suspended, Active, Revoked)
	}
}

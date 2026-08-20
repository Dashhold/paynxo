package apierr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestTranslateMapsKindsToStatusAndCode(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"validation", ErrValidation, http.StatusBadRequest, CodeValidation},
		{"unauthenticated", ErrUnauthenticated, http.StatusUnauthorized, CodeUnauthenticated},
		{"invalid credentials", ErrInvalidCredentials, http.StatusUnauthorized, CodeInvalidCredentials},
		{"forbidden", ErrForbidden, http.StatusForbidden, CodeForbidden},
		{"lease inactive", ErrLeaseInactive, http.StatusForbidden, CodeLeaseInactive},
		{"not found", ErrNotFound, http.StatusNotFound, CodeNotFound},
		{"conflict", ErrConflict, http.StatusConflict, CodeConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := Translate(tc.err)
			if status != tc.wantStatus {
				t.Errorf("status = %d, want %d", status, tc.wantStatus)
			}
			if body.Code != tc.wantCode {
				t.Errorf("code = %q, want %q", body.Code, tc.wantCode)
			}
			if body.Message == "" {
				t.Errorf("message should not be empty")
			}
		})
	}
}

func TestTranslateValidationCarriesFields(t *testing.T) {
	err := ValidationField("expiryDate", "must be after start date")
	status, body := Translate(err)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if body.Fields["expiryDate"] != "must be after start date" {
		t.Errorf("fields = %v, want offending field present", body.Fields)
	}
}

func TestTranslateUnexpectedErrorIsGenericSafe500(t *testing.T) {
	secret := "super-secret-token-value"
	err := fmt.Errorf("db exploded with %s and stack trace", secret)
	status, body := Translate(err)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if body.Code != CodeInternal {
		t.Errorf("code = %q, want %q", body.Code, CodeInternal)
	}
	if body.Message == "" || body.Message == err.Error() {
		t.Errorf("500 message must be generic, got %q", body.Message)
	}
	if body.Fields != nil {
		t.Errorf("500 body must not carry fields, got %v", body.Fields)
	}
}

func TestServiceErrorWrappedAsInternalDoesNotLeak(t *testing.T) {
	// A ServiceError explicitly built as internal must still be generic.
	err := &ServiceError{Kind: KindInternal, Message: "secret leak: password=hunter2"}
	_, body := Translate(err)
	if body.Message == err.Message {
		t.Errorf("internal ServiceError leaked its message: %q", body.Message)
	}
}

func TestErrorsIsMatchesByKind(t *testing.T) {
	wrapped := fmt.Errorf("layer: %w", NotFound("merchant missing"))
	if !errors.Is(wrapped, ErrNotFound) {
		t.Errorf("errors.Is should match ErrNotFound through wrapping")
	}
	if errors.Is(wrapped, ErrConflict) {
		t.Errorf("errors.Is should not match a different kind")
	}
}

func TestServiceErrorMessageIncludesFields(t *testing.T) {
	err := Validation("invalid request", map[string]string{"name": "required", "email": "invalid"})
	got := err.Error()
	if got == "invalid request" {
		t.Errorf("expected field detail in error string, got %q", got)
	}
}

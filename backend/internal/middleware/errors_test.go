package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pgcs/backend/internal/apierr"
)

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) apierr.APIError {
	t.Helper()
	var body apierr.APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid APIError JSON: %v (%s)", err, rec.Body.String())
	}
	return body
}

func TestErrorMiddlewareRendersTypedError(t *testing.T) {
	h := Error(func(w http.ResponseWriter, r *http.Request) error {
		return apierr.NotFound("merchant not found")
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/merchants/x", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	body := decodeBody(t, rec)
	if body.Code != apierr.CodeNotFound {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeNotFound)
	}
}

func TestErrorMiddlewareValidationFields(t *testing.T) {
	h := Error(func(w http.ResponseWriter, r *http.Request) error {
		return apierr.ValidationField("name", "required")
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/companies", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	body := decodeBody(t, rec)
	if body.Fields["name"] != "required" {
		t.Errorf("expected offending field in body, got %v", body.Fields)
	}
}

func TestErrorMiddlewareNilDoesNotOverwrite(t *testing.T) {
	h := Error(func(w http.ResponseWriter, r *http.Request) error {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "yes"})
		return nil
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestErrorMiddlewareUnexpectedErrorIsSafe500(t *testing.T) {
	secret := "TOKEN_SECRET=should-not-appear"
	h := Error(func(w http.ResponseWriter, r *http.Request) error {
		return errorsNew(secret)
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), secret) {
		t.Errorf("500 response leaked secret content: %s", rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body.Code != apierr.CodeInternal {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeInternal)
	}
}

// errorsNew returns a plain untyped error carrying the given message.
func errorsNew(msg string) error { return &plainError{msg} }

type plainError struct{ s string }

func (e *plainError) Error() string { return e.s }

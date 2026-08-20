package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pgcs/backend/internal/apierr"
)

func TestRecoverTrapsPanicAndReturnsSafe500(t *testing.T) {
	secret := "panic with password=hunter2"
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(secret)
	}))

	rec := httptest.NewRecorder()
	// Should not propagate the panic to the caller.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "hunter2") {
		t.Errorf("recover response leaked panic content: %s", rec.Body.String())
	}
	var body apierr.APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not valid APIError JSON: %v", err)
	}
	if body.Code != apierr.CodeInternal {
		t.Errorf("code = %q, want %q", body.Code, apierr.CodeInternal)
	}
}

func TestRecoverPassesThroughWhenNoPanic(t *testing.T) {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "yes"})
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ok", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

// TestServerSurvivesHandlerPanic confirms a panic in one handler does not crash
// the server: a subsequent request on the same wrapped handler still succeeds.
func TestServerSurvivesHandlerPanic(t *testing.T) {
	calls := 0
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			panic("boom")
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "yes"})
	}))

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/api/x", nil))
	if rec1.Code != http.StatusInternalServerError {
		t.Fatalf("first call status = %d, want 500", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/api/x", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("second call status = %d, want 200", rec2.Code)
	}
}

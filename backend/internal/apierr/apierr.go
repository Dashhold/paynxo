// Package apierr defines the structured API error model and the typed
// (sentinel) service errors used throughout the backend.
//
// Business logic in the service and repository layers returns these typed
// errors; the HTTP layer (see internal/middleware) converts them into the
// consistent APIError JSON body and the appropriate HTTP status code. Keeping
// the error vocabulary in one small package lets every layer agree on error
// semantics without importing net/http.
//
// The 500 (internal) path deliberately produces a generic message so that
// unexpected errors never leak internal details, stack traces, or secret
// values into a response body (Req 18.4).
package apierr

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Machine-readable error codes carried in APIError.Code. These are stable
// identifiers the frontend can branch on (Req 18.1).
const (
	CodeValidation         = "validation_error"
	CodeUnauthenticated    = "unauthenticated"
	CodeInvalidCredentials = "invalid_credentials"
	CodeForbidden          = "forbidden"
	CodeLeaseInactive      = "lease_inactive"
	CodeNotFound           = "not_found"
	CodeConflict           = "conflict"
	CodeInternal           = "internal_error"
)

// Kind classifies a ServiceError independently of its message or field detail.
// Sentinel errors and errors.Is matching are driven by Kind.
type Kind int

const (
	// KindInternal is the zero value and represents an unexpected error that
	// maps to HTTP 500 with a generic, non-leaking message.
	KindInternal Kind = iota
	KindValidation
	KindUnauthenticated
	KindInvalidCredentials
	KindForbidden
	KindLeaseInactive
	KindNotFound
	KindConflict
)

// APIError is the structured error body returned to clients. It always carries
// a machine-readable Code and a human-readable Message; Fields is populated for
// validation errors to identify the offending field(s) (Req 18.1, 18.3).
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ServiceError is the typed error returned by the service and repository
// layers. Its Kind selects the HTTP status and error code; Message provides a
// human-readable description; Fields carries per-field validation detail.
type ServiceError struct {
	Kind    Kind
	Message string
	Fields  map[string]string
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if len(e.Fields) > 0 {
		// Sort keys so the string form is deterministic.
		keys := make([]string, 0, len(e.Fields))
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, e.Fields[k]))
		}
		return fmt.Sprintf("%s (%s)", e.Message, strings.Join(parts, "; "))
	}
	return e.Message
}

// Is reports whether target is a ServiceError of the same Kind, enabling
// errors.Is(err, ErrNotFound) style checks regardless of the concrete Message
// or Fields carried by err.
func (e *ServiceError) Is(target error) bool {
	var t *ServiceError
	if !errors.As(target, &t) {
		return false
	}
	return e.Kind == t.Kind
}

// withFields returns a copy of e with the given fields attached. It never
// mutates the receiver so package-level sentinels stay immutable.
func (e *ServiceError) withFields(fields map[string]string) *ServiceError {
	cp := &ServiceError{Kind: e.Kind, Message: e.Message, Fields: fields}
	return cp
}

// Sentinel errors. Use these directly, with errors.Is, or via the constructors
// below to attach a specific message / field detail.
var (
	// ErrNotFound indicates a referenced entity does not exist within the
	// principal's tenant scope; maps to HTTP 404 (Req 18.2).
	ErrNotFound = &ServiceError{Kind: KindNotFound, Message: "resource not found"}
	// ErrValidation indicates request validation failed; maps to HTTP 400
	// and carries the offending field(s) (Req 18.3).
	ErrValidation = &ServiceError{Kind: KindValidation, Message: "validation failed"}
	// ErrConflict indicates a uniqueness/state conflict; maps to HTTP 409.
	ErrConflict = &ServiceError{Kind: KindConflict, Message: "conflict"}
	// ErrForbidden indicates the principal lacks permission; maps to HTTP 403.
	ErrForbidden = &ServiceError{Kind: KindForbidden, Message: "forbidden"}
	// ErrUnauthenticated indicates a missing/invalid/expired token; maps to
	// HTTP 401.
	ErrUnauthenticated = &ServiceError{Kind: KindUnauthenticated, Message: "authentication required"}
	// ErrInvalidCredentials indicates a failed login; maps to HTTP 401 with a
	// generic message that does not reveal which field was wrong (Req 6.2).
	ErrInvalidCredentials = &ServiceError{Kind: KindInvalidCredentials, Message: "invalid credentials"}
	// ErrLeaseInactive indicates the principal's lease is expired, suspended,
	// or revoked; maps to HTTP 403 (Req 14.3, 15.4, 15.6).
	ErrLeaseInactive = &ServiceError{Kind: KindLeaseInactive, Message: "lease is not active"}
)

// Validation builds a 400 validation error with the given message and
// field-level detail (field name -> reason).
func Validation(message string, fields map[string]string) *ServiceError {
	if message == "" {
		message = ErrValidation.Message
	}
	return &ServiceError{Kind: KindValidation, Message: message, Fields: fields}
}

// ValidationField is a convenience for a single offending field.
func ValidationField(field, reason string) *ServiceError {
	msg := reason
	if msg == "" {
		msg = ErrValidation.Message
	}
	return &ServiceError{
		Kind:    KindValidation,
		Message: msg,
		Fields:  map[string]string{field: reason},
	}
}

// NotFound builds a 404 error with an optional custom message.
func NotFound(message string) *ServiceError {
	if message == "" {
		return ErrNotFound
	}
	return &ServiceError{Kind: KindNotFound, Message: message}
}

// Conflict builds a 409 error with an optional custom message.
func Conflict(message string) *ServiceError {
	if message == "" {
		return ErrConflict
	}
	return &ServiceError{Kind: KindConflict, Message: message}
}

// Forbidden builds a 403 error with an optional custom message.
func Forbidden(message string) *ServiceError {
	if message == "" {
		return ErrForbidden
	}
	return &ServiceError{Kind: KindForbidden, Message: message}
}

// Unauthenticated builds a 401 error with an optional custom message.
func Unauthenticated(message string) *ServiceError {
	if message == "" {
		return ErrUnauthenticated
	}
	return &ServiceError{Kind: KindUnauthenticated, Message: message}
}

// LeaseInactive builds a 403 lease-inactive error with an optional custom
// message (e.g. "lease has expired", "lease is suspended").
func LeaseInactive(message string) *ServiceError {
	if message == "" {
		return ErrLeaseInactive
	}
	return &ServiceError{Kind: KindLeaseInactive, Message: message}
}

// statusByKind maps each Kind to its HTTP status code and error code per the
// design's error-handling table.
func statusByKind(kind Kind) (int, string) {
	switch kind {
	case KindValidation:
		return http.StatusBadRequest, CodeValidation // 400
	case KindUnauthenticated:
		return http.StatusUnauthorized, CodeUnauthenticated // 401
	case KindInvalidCredentials:
		return http.StatusUnauthorized, CodeInvalidCredentials // 401
	case KindForbidden:
		return http.StatusForbidden, CodeForbidden // 403
	case KindLeaseInactive:
		return http.StatusForbidden, CodeLeaseInactive // 403
	case KindNotFound:
		return http.StatusNotFound, CodeNotFound // 404
	case KindConflict:
		return http.StatusConflict, CodeConflict // 409
	default:
		return http.StatusInternalServerError, CodeInternal // 500
	}
}

// HTTPStatus returns the HTTP status code for an arbitrary error. Unknown /
// untyped errors map to 500.
func HTTPStatus(err error) int {
	status, _ := Translate(err)
	return status
}

// Translate converts any error into the HTTP status code and structured
// APIError body to return to the client.
//
// Recognized ServiceError values are mapped per the design table. Any other
// (untyped/unexpected) error is treated as an internal error: it maps to 500
// with a fixed generic message so that internal details, stack traces, or
// secret values are never exposed in the response body (Req 18.4, 18.5).
func Translate(err error) (int, APIError) {
	if err == nil {
		return http.StatusOK, APIError{}
	}

	var se *ServiceError
	if errors.As(err, &se) {
		status, code := statusByKind(se.Kind)
		if status == http.StatusInternalServerError {
			return status, internalAPIError()
		}
		msg := se.Message
		if msg == "" {
			msg = http.StatusText(status)
		}
		return status, APIError{Code: code, Message: msg, Fields: se.Fields}
	}

	// Untyped error: never surface its content to the client.
	return http.StatusInternalServerError, internalAPIError()
}

// internalAPIError is the fixed body for any 500 response. It is intentionally
// generic and contains no dynamic content (Req 18.4).
func internalAPIError() APIError {
	return APIError{Code: CodeInternal, Message: "an internal error occurred"}
}

// Internal returns the standard structured body for an unexpected server
// error. Callers (e.g. the Recover middleware) use this to render a safe 500.
func Internal() APIError {
	return internalAPIError()
}

// IsUniqueViolation reports whether err looks like a database unique-constraint
// violation (e.g. a duplicate login user id racing past the application-level
// check). Service layers use it to convert such a race into a clean 409
// Conflict instead of a generic 500. The check is driver-agnostic: it matches
// the SQLSTATE 23505 marker and the common textual forms emitted by Postgres.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "unique constraint")
}

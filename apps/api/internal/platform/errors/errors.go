// Package errors provides standardized error types and HTTP response helpers
// for the Zenvikar API. Domain errors are mapped to consistent JSON responses
// with appropriate HTTP status codes.
package errors

import (
	"encoding/json"
	"net/http"
)

// Error codes used across the API.
const (
	CodeTenantNotFound = "tenant_not_found"
	CodeTenantDisabled = "tenant_disabled"
	CodeSlotUnavailable = "slot_unavailable"
	CodeSlugReserved   = "slug_reserved"
	CodeForbidden      = "forbidden"
	CodeInfraFailure   = "service_unavailable"
	CodeInternalError  = "internal_error"
)

// AppError represents a structured API error with an error code,
// human-readable message, and corresponding HTTP status.
type AppError struct {
	Code       string `json:"error"`
	Message    string `json:"message,omitempty"`
	HTTPStatus int    `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// codeStatusMap maps known error codes to their HTTP status codes.
var codeStatusMap = map[string]int{
	CodeTenantNotFound:  http.StatusNotFound,            // 404
	CodeTenantDisabled:  http.StatusForbidden,            // 403
	CodeSlotUnavailable: http.StatusConflict,             // 409
	CodeSlugReserved:    http.StatusUnprocessableEntity,   // 422
	CodeForbidden:       http.StatusForbidden,            // 403
	CodeInfraFailure:    http.StatusServiceUnavailable,   // 503
	CodeInternalError:   http.StatusInternalServerError,  // 500
}

// New creates an AppError with the given code and message.
// The HTTP status is looked up from the known code map; unknown codes default to 500.
func New(code, message string) *AppError {
	status, ok := codeStatusMap[code]
	if !ok {
		status = http.StatusInternalServerError
	}
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// NewTenantNotFound returns a tenant_not_found error.
func NewTenantNotFound() *AppError {
	return New(CodeTenantNotFound, "No tenant found for this address")
}

// NewTenantDisabled returns a tenant_disabled error.
func NewTenantDisabled() *AppError {
	return New(CodeTenantDisabled, "This booking page is currently unavailable")
}

// NewSlotUnavailable returns a slot_unavailable error.
func NewSlotUnavailable() *AppError {
	return New(CodeSlotUnavailable, "The requested time slot is no longer available")
}

// NewSlugReserved returns a slug_reserved error.
func NewSlugReserved() *AppError {
	return New(CodeSlugReserved, "This name is not available")
}

// NewForbidden returns a forbidden error for membership/authorization failures.
func NewForbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

// NewInfraFailure returns a service_unavailable error for database/Redis failures.
func NewInfraFailure() *AppError {
	return New(CodeInfraFailure, "Service temporarily unavailable")
}

// WriteError writes a JSON error response to the http.ResponseWriter.
// If err is an *AppError, its code, message, and status are used directly.
// Otherwise a generic 500 internal error is returned.
func WriteError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		appErr = &AppError{
			Code:       CodeInternalError,
			Message:    "An unexpected error occurred",
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	_ = json.NewEncoder(w).Encode(appErr)
}

// ErrorHandler returns middleware that recovers *AppError values from panics
// and writes them as standardized JSON responses. Non-AppError panics are
// re-raised so the upstream recovery middleware can handle them.
func ErrorHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if appErr, ok := rec.(*AppError); ok {
						WriteError(w, appErr)
						return
					}
					// Re-panic for non-AppError values so recovery middleware handles them.
					panic(rec)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

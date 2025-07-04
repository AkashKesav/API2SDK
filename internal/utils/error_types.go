package utils

import (
	"fmt"
	"net/http"
)

// ErrorType represents different types of errors in the system
type ErrorType string

const (
	// ValidationError represents validation failures
	ValidationError ErrorType = "validation_error"
	// AuthenticationError represents authentication failures
	AuthenticationError ErrorType = "authentication_error"
	// AuthorizationError represents authorization failures
	AuthorizationError ErrorType = "authorization_error"
	// NotFoundError represents resource not found errors
	NotFoundError ErrorType = "not_found_error"
	// ConflictError represents resource conflict errors
	ConflictError ErrorType = "conflict_error"
	// ExternalServiceError represents errors from external services
	ExternalServiceError ErrorType = "external_service_error"
	// DatabaseError represents database-related errors
	DatabaseError ErrorType = "database_error"
	// InternalError represents internal system errors
	InternalError ErrorType = "internal_error"
	// RateLimitError represents rate limiting errors
	RateLimitError ErrorType = "rate_limit_error"
	// TimeoutError represents timeout errors
	TimeoutError ErrorType = "timeout_error"
	// CircuitBreakerError represents circuit breaker errors
	CircuitBreakerError ErrorType = "circuit_breaker_error"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Code       string                 `json:"code,omitempty"`
	StatusCode int                    `json:"status_code"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause of the error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(message, details string) *AppError {
	return &AppError{
		Type:       ValidationError,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_FAILED",
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Type:       AuthenticationError,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Code:       "AUTHENTICATION_REQUIRED",
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:       AuthorizationError,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Code:       "INSUFFICIENT_PERMISSIONS",
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *AppError {
	return &AppError{
		Type:       NotFoundError,
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    fmt.Sprintf("No %s found with ID: %s", resource, id),
		StatusCode: http.StatusNotFound,
		Code:       "RESOURCE_NOT_FOUND",
		Metadata: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message, details string) *AppError {
	return &AppError{
		Type:       ConflictError,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusConflict,
		Code:       "RESOURCE_CONFLICT",
	}
}

// NewExternalServiceError creates a new external service error
func NewExternalServiceError(service, message string, cause error) *AppError {
	return &AppError{
		Type:       ExternalServiceError,
		Message:    fmt.Sprintf("External service error: %s", service),
		Details:    message,
		StatusCode: http.StatusBadGateway,
		Code:       "EXTERNAL_SERVICE_ERROR",
		Cause:      cause,
		Metadata: map[string]interface{}{
			"service": service,
		},
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation, message string, cause error) *AppError {
	return &AppError{
		Type:       DatabaseError,
		Message:    "Database operation failed",
		Details:    message,
		StatusCode: http.StatusInternalServerError,
		Code:       "DATABASE_ERROR",
		Cause:      cause,
		Metadata: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:       InternalError,
		Message:    "Internal server error",
		Details:    message,
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Cause:      cause,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(limit int, window string) *AppError {
	return &AppError{
		Type:       RateLimitError,
		Message:    "Rate limit exceeded",
		Details:    fmt.Sprintf("Maximum %d requests per %s allowed", limit, window),
		StatusCode: http.StatusTooManyRequests,
		Code:       "RATE_LIMIT_EXCEEDED",
		Metadata: map[string]interface{}{
			"limit":  limit,
			"window": window,
		},
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, timeout string) *AppError {
	return &AppError{
		Type:       TimeoutError,
		Message:    "Operation timed out",
		Details:    fmt.Sprintf("Operation '%s' exceeded timeout of %s", operation, timeout),
		StatusCode: http.StatusRequestTimeout,
		Code:       "OPERATION_TIMEOUT",
		Metadata: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		},
	}
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(service string) *AppError {
	return &AppError{
		Type:       CircuitBreakerError,
		Message:    "Service temporarily unavailable",
		Details:    fmt.Sprintf("Circuit breaker is open for service: %s", service),
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Metadata: map[string]interface{}{
			"service": service,
		},
	}
}

// IsRetryable determines if an error should be retried
func (e *AppError) IsRetryable() bool {
	switch e.Type {
	case ExternalServiceError, DatabaseError, TimeoutError:
		return true
	case CircuitBreakerError:
		return false // Circuit breaker should handle retries
	default:
		return false
	}
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithCode sets the error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// ToHTTPResponse converts the error to an HTTP response format
func (e *AppError) ToHTTPResponse() map[string]interface{} {
	response := map[string]interface{}{
		"error":   true,
		"type":    string(e.Type),
		"message": e.Message,
		"code":    e.Code,
	}

	if e.Details != "" {
		response["details"] = e.Details
	}

	if e.Metadata != nil {
		response["metadata"] = e.Metadata
	}

	return response
}
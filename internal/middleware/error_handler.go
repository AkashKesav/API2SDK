package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorHandlerConfig defines the config for error handler middleware
type ErrorHandlerConfig struct {
	Logger           *zap.Logger
	EnableStackTrace bool
	EnableDebugMode  bool
}

// DefaultErrorHandlerConfig is the default error handler configuration
var DefaultErrorHandlerConfig = ErrorHandlerConfig{
	EnableStackTrace: false,
	EnableDebugMode:  false,
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error      bool        `json:"error"`
	Message    string      `json:"message"`
	Code       int         `json:"code"`
	Details    interface{} `json:"details,omitempty"`
	StackTrace string      `json:"stack_trace,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
}

// ErrorHandlerMiddleware creates a centralized error handling middleware
func ErrorHandlerMiddleware(config ...ErrorHandlerConfig) fiber.Handler {
	// Set default config
	cfg := DefaultErrorHandlerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Continue to next handler
		err := c.Next()

		if err != nil {
			return handleError(c, err, cfg)
		}

		return nil
	}
}

// handleError processes and responds to errors
func handleError(c fiber.Ctx, err error, cfg ErrorHandlerConfig) error {
	// Get request ID for tracing
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = c.Locals("request_id").(string)
	}

	// Initialize error response
	errorResp := ErrorResponse{
		Error:     true,
		RequestID: requestID,
	}

	// Handle different error types
	var fiberErr *fiber.Error
	var appErr *utils.AppError

	switch {
	case errors.As(err, &appErr):
		// Handle AppError type
		errorResp.Code = appErr.StatusCode
		errorResp.Message = appErr.Message
		if appErr.Details != "" {
			errorResp.Details = appErr.Details
		}
		if appErr.Metadata != nil {
			// If we already have details as a string, convert to map and merge
			if details, ok := errorResp.Details.(string); ok {
				metadataWithDetails := make(map[string]interface{})
				metadataWithDetails["message"] = details
				for k, v := range appErr.Metadata {
					metadataWithDetails[k] = v
				}
				errorResp.Details = metadataWithDetails
			} else {
				errorResp.Details = appErr.Metadata
			}
		}

	case errors.As(err, &fiberErr):
		// Fiber framework errors
		errorResp.Code = fiberErr.Code
		errorResp.Message = fiberErr.Message

	default:
		// Unknown errors
		errorResp.Code = fiber.StatusInternalServerError
		errorResp.Message = "Internal server error"

		// Log unknown errors
		if cfg.Logger != nil {
			cfg.Logger.Error("Unknown error occurred",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
		}

		// Include error message in debug mode
		if cfg.EnableDebugMode {
			errorResp.Message = err.Error()
		}
	}

	// Add stack trace if enabled
	if cfg.EnableStackTrace {
		errorResp.StackTrace = string(debug.Stack())
	}

	// Log error details
	if cfg.Logger != nil {
		cfg.Logger.Info("Error response sent",
			zap.String("request_id", requestID),
			zap.Int("status_code", errorResp.Code),
			zap.String("message", errorResp.Message),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
		)
	}

	return c.Status(errorResp.Code).JSON(errorResp)
}

// RecoveryMiddleware handles panics and converts them to errors
func RecoveryMiddleware(config ...ErrorHandlerConfig) fiber.Handler {
	// Set default config
	cfg := DefaultErrorHandlerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Get request ID for tracing
				requestID := c.Get("X-Request-ID")
				if requestID == "" {
					if rid := c.Locals("request_id"); rid != nil {
						requestID = rid.(string)
					}
				}

				// Log panic
				if cfg.Logger != nil {
					cfg.Logger.Error("Panic recovered",
						zap.String("request_id", requestID),
						zap.Any("panic", r),
						zap.String("stack", string(debug.Stack())),
						zap.String("method", c.Method()),
						zap.String("path", c.Path()),
						zap.String("ip", c.IP()),
					)
				}

				// Create error response
				errorResp := ErrorResponse{
					Error:     true,
					Code:      fiber.StatusInternalServerError,
					Message:   "Internal server error",
					RequestID: requestID,
				}

				// Include panic details in debug mode
				if cfg.EnableDebugMode {
					errorResp.Details = map[string]interface{}{
						"panic": fmt.Sprintf("%v", r),
					}
				}

				// Include stack trace if enabled
				if cfg.EnableStackTrace {
					errorResp.StackTrace = string(debug.Stack())
				}

				c.Status(fiber.StatusInternalServerError).JSON(errorResp)
			}
		}()

		return c.Next()
	}
}

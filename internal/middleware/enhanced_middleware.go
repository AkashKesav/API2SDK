package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// TracingMiddleware adds request tracing to all requests
func TracingMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Create trace context
		tc := utils.NewTraceContext(logger)
		tc.SetTag("method", c.Method())
		tc.SetTag("path", c.Path())
		tc.SetTag("user_agent", c.Get("User-Agent"))
		tc.SetTag("remote_ip", c.IP())

		// Add trace ID to response headers
		c.Set("X-Trace-ID", tc.TraceID)

		// Store trace context in fiber context
		c.Locals("trace_context", tc)

		tc.LogInfo("Request started")

		// Process request
		err := c.Next()

		// Add response status to trace
		tc.SetTag("status_code", strconv.Itoa(c.Response().StatusCode()))

		if err != nil {
			tc.LogError("Request failed", err)
		} else {
			tc.LogInfo("Request completed")
		}

		tc.Finish()
		return err
	}
}

// MetricsMiddleware collects metrics for all requests
func MetricsMiddleware(logger *zap.Logger) fiber.Handler {
	metrics := utils.GetGlobalMetricsCollector(logger)
	requestMetrics := utils.NewRequestMetrics(metrics)

	return func(c fiber.Ctx) error {
		startTime := time.Now()
		path := c.Path()
		method := c.Method()

		// Process request
		err := c.Next()

		// Track metrics
		statusCode := c.Response().StatusCode()
		requestMetrics.TrackRequest(path, method, statusCode, startTime)

		return err
	}
}

// EnhancedRateLimitMiddleware provides enhanced rate limiting with metrics
func EnhancedRateLimitMiddleware(limiter *RateLimiter, logger *zap.Logger) fiber.Handler {
	metrics := utils.GetGlobalMetricsCollector(logger)

	return func(c fiber.Ctx) error {
		ip := c.IP()

		// Check rate limit
		if !limiter.Allow(ip) {
			// Track rate limit violations
			metrics.Counter("rate_limit_violations_total", map[string]string{
				"ip":   ip,
				"path": c.Path(),
			}).Inc()

			remaining := limiter.GetRemaining(ip)

			// Get trace context if available
			if tc, ok := c.Locals("trace_context").(*utils.TraceContext); ok {
				tc.LogInfo("Rate limit exceeded",
					zap.String("ip", ip),
					zap.Int("remaining", remaining))
			}

			return c.Status(fiber.StatusTooManyRequests).JSON(utils.NewRateLimitError(
				limiter.limit,
				limiter.window.String(),
			).ToHTTPResponse())
		}

		// Track successful requests
		metrics.Counter("rate_limit_allowed_total", map[string]string{
			"ip":   ip,
			"path": c.Path(),
		}).Inc()

		// Add rate limit headers
		remaining := limiter.GetRemaining(ip)
		c.Set("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.window).Unix(), 10))

		return c.Next()
	}
}

// CircuitBreakerMiddleware provides circuit breaker protection for routes
func CircuitBreakerMiddleware(serviceName string, logger *zap.Logger) fiber.Handler {
	registry := utils.NewCircuitBreakerRegistry(logger)
	cb := registry.Get(serviceName, utils.CircuitBreakerConfig{
		Name:             serviceName,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		Logger:           logger,
	})

	return func(c fiber.Ctx) error {
		// Check if circuit breaker allows the request
		if !cb.AllowRequest() {
			// Get trace context if available
			if tc, ok := c.Locals("trace_context").(*utils.TraceContext); ok {
				tc.LogInfo("Circuit breaker is open", zap.String("service", serviceName))
			}

			return c.Status(fiber.StatusServiceUnavailable).JSON(
				utils.NewCircuitBreakerError(serviceName).ToHTTPResponse(),
			)
		}

		// Execute the request
		err := c.Next()

		// Record the result
		if err != nil {
			cb.RecordFailure()
		} else {
			// Consider 5xx status codes as failures
			if c.Response().StatusCode() >= 500 {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
			}
		}

		return err
	}
}

// ErrorHandlingMiddleware provides enhanced error handling
func ErrorHandlingMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			// Get trace context if available
			var tc *utils.TraceContext
			if traceCtx, ok := c.Locals("trace_context").(*utils.TraceContext); ok {
				tc = traceCtx
			}

			// Handle different error types
			switch e := err.(type) {
			case *utils.AppError:
				if tc != nil {
					tc.LogError("Application error", err)
				}
				return c.Status(e.StatusCode).JSON(e.ToHTTPResponse())
			case *fiber.Error:
				if tc != nil {
					tc.LogError("Fiber error", err)
				}
				return c.Status(e.Code).JSON(map[string]interface{}{
					"error":   true,
					"message": e.Message,
					"code":    e.Code,
				})
			default:
				if tc != nil {
					tc.LogError("Unknown error", err)
				}
				logger.Error("Unhandled error", zap.Error(err))
				return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
					"error":   true,
					"message": "Internal server error",
					"code":    fiber.StatusInternalServerError,
				})
			}
		}

		return nil
	}
}

// HealthCheckMiddleware provides health check functionality
func HealthCheckMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Skip health check paths
		if c.Path() == "/health" || c.Path() == "/health/ready" || c.Path() == "/health/live" {
			return c.Next()
		}

		// For other paths, continue normally
		return c.Next()
	}
}

// CompressionMiddleware adds response compression
func CompressionMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check if client accepts compression
		acceptEncoding := c.Get("Accept-Encoding")
		if acceptEncoding == "" {
			return c.Next()
		}

		// Set compression header if supported
		if strings.Contains(acceptEncoding, "gzip") {
			c.Set("Content-Encoding", "gzip")
		}

		return c.Next()
	}
}

// AuthSkipMiddleware skips authentication for certain paths
func AuthSkipMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		path := c.Path()

		// Skip auth for MCP routes, health checks, and public endpoints
		skipPaths := []string{
			"/mcp/",
			"/api/v1/health",
			"/api/v1/auth/",
			"/api/v1/public-apis",
			"/",
			"/login",
			"/register",
		}

		for _, skipPath := range skipPaths {
			if strings.HasPrefix(path, skipPath) {
				// Set mock user context for these routes if needed
				c.Locals("user_id", "mock-user-123")
				c.Locals("user_role", "admin")
				return c.Next()
			}
		}

		return c.Next()
	}
}

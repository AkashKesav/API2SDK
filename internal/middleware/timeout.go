package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// TimeoutConfig defines the config for timeout middleware
type TimeoutConfig struct {
	// Timeout defines the timeout duration for requests
	Timeout time.Duration

	// TimeoutHandler is called when a request times out
	TimeoutHandler fiber.Handler
}

// DefaultTimeoutConfig is the default timeout config
var DefaultTimeoutConfig = TimeoutConfig{
	Timeout: 150 * time.Second, // Default 2.5 minutes
	TimeoutHandler: func(c fiber.Ctx) error {
		return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
			"error":   true,
			"message": "Request timeout - operation took too long to complete. Please try again or use a smaller collection.",
			"code":    fiber.StatusRequestTimeout,
		})
	},
}

// TimeoutMiddleware creates a new timeout middleware with config
func TimeoutMiddleware(config ...TimeoutConfig) fiber.Handler {
	cfg := DefaultTimeoutConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Context(), cfg.Timeout)
		defer cancel()

		// Set the context with timeout on the fiber context
		c.SetContext(ctx)

		// Create a channel to signal completion
		done := make(chan error, 1)

		// Run the next handler in a goroutine
		go func() {
			done <- c.Next()
		}()

		// Wait for either completion or timeout
		select {
		case err := <-done:
			// Request completed within timeout
			return err
		case <-ctx.Done():
			// Request timed out
			if ctx.Err() == context.DeadlineExceeded {
				return cfg.TimeoutHandler(c)
			}
			return ctx.Err()
		}
	}
}

// LongOperationTimeoutMiddleware creates timeout middleware specifically for long operations
func LongOperationTimeoutMiddleware() fiber.Handler {
	return TimeoutMiddleware(TimeoutConfig{
		Timeout: 170 * time.Second, // Slightly less than server timeout
		TimeoutHandler: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"error":   true,
				"message": "Request timeout - operation took too long to complete. Please try again or use a smaller collection.",
				"code":    fiber.StatusRequestTimeout,
			})
		},
	})
}

// ShortOperationTimeoutMiddleware creates timeout middleware for quick operations
func ShortOperationTimeoutMiddleware() fiber.Handler {
	return TimeoutMiddleware(TimeoutConfig{
		Timeout: 30 * time.Second, // 30 seconds for quick operations
		TimeoutHandler: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"error":   true,
				"message": "Request timeout - operation took too long to complete. Please try again.",
				"code":    fiber.StatusRequestTimeout,
			})
		},
	})
}

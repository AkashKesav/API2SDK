package utils

import (
	"context"
	"errors"
	"math"
	"time"

	"go.uber.org/zap"
)

// RetryConfig defines configuration for retry operations
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64
	RetryableErrors []error
}

// DefaultRetryConfig provides sensible defaults for retry operations
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialInterval: 100 * time.Millisecond,
	MaxInterval:     10 * time.Second,
	Multiplier:      2.0,
	RandomFactor:    0.1,
}

// IsRetryable checks if an error should be retried based on the configuration
func (c *RetryConfig) IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// If no specific errors are defined, retry all errors
	if len(c.RetryableErrors) == 0 {
		return true
	}

	// Check if the error is in the list of retryable errors
	for _, retryableErr := range c.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// RetryWithBackoff executes the given function with exponential backoff
func RetryWithBackoff(ctx context.Context, logger *zap.Logger, operation string, fn func() error, config RetryConfig) error {
	var err error
	currentInterval := config.InitialInterval

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err = fn()
		
		// If no error or error is not retryable, return immediately
		if err == nil || !config.IsRetryable(err) {
			return err
		}

		// Check if we've reached max retries
		if attempt == config.MaxRetries {
			logger.Warn("Max retries reached",
				zap.String("operation", operation),
				zap.Int("attempts", attempt+1),
				zap.Error(err))
			return err
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			logger.Warn("Context cancelled during retry",
				zap.String("operation", operation),
				zap.Int("attempts", attempt+1),
				zap.Error(ctx.Err()))
			return ctx.Err()
		default:
			// Calculate next backoff interval with jitter
			jitter := 1.0 + config.RandomFactor*(2.0*math.Float64frombits(uint64(time.Now().UnixNano()))/math.MaxUint64 - 1.0)
			nextInterval := time.Duration(float64(currentInterval) * jitter)
			
			logger.Debug("Retrying operation",
				zap.String("operation", operation),
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", nextInterval),
				zap.Error(err))
			
			// Wait for backoff period
			timer := time.NewTimer(nextInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// Continue with next attempt
			}
			
			// Increase interval for next attempt
			currentInterval = time.Duration(float64(currentInterval) * config.Multiplier)
			if currentInterval > config.MaxInterval {
				currentInterval = config.MaxInterval
			}
		}
	}

	return err
}
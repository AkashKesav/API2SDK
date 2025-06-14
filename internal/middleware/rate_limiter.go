package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

// RateLimiter represents a rate limiting configuration
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, timestamps := range rl.requests {
			var validTimestamps []time.Time
			for _, timestamp := range timestamps {
				if timestamp.After(cutoff) {
					validTimestamps = append(validTimestamps, timestamp)
				}
			}
			if len(validTimestamps) > 0 {
				rl.requests[ip] = validTimestamps
			} else {
				delete(rl.requests, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing timestamps for this IP
	timestamps := rl.requests[ip]

	// Filter out old timestamps
	var validTimestamps []time.Time
	for _, timestamp := range timestamps {
		if timestamp.After(cutoff) {
			validTimestamps = append(validTimestamps, timestamp)
		}
	}

	// Check if we're under the limit
	if len(validTimestamps) >= rl.limit {
		return false
	}

	// Add current timestamp
	validTimestamps = append(validTimestamps, now)
	rl.requests[ip] = validTimestamps

	return true
}

// GetRemaining returns the number of remaining requests for an IP
func (rl *RateLimiter) GetRemaining(ip string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	cutoff := time.Now().Add(-rl.window)
	timestamps := rl.requests[ip]

	var validCount int
	for _, timestamp := range timestamps {
		if timestamp.After(cutoff) {
			validCount++
		}
	}

	remaining := rl.limit - validCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// Global rate limiters for different endpoints
var (
	// General API rate limiter: 100 requests per minute
	apiLimiter = NewRateLimiter(100, time.Minute)

	// SDK generation rate limiter: 10 requests per minute (more restrictive)
	sdkLimiter = NewRateLimiter(10, time.Minute)

	// Public API rate limiter: 50 requests per minute
	publicAPILimiter = NewRateLimiter(50, time.Minute)
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get client IP
		ip := c.IP()

		// Check if request is allowed
		if !limiter.Allow(ip) {
			remaining := limiter.GetRemaining(ip)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       true,
				"message":     "Rate limit exceeded",
				"remaining":   remaining,
				"retry_after": int(limiter.window.Seconds()),
			})
		}

		// Add rate limit headers
		remaining := limiter.GetRemaining(ip)
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.window).Unix()))

		return c.Next()
	}
}

// APIRateLimitMiddleware applies general API rate limiting
func APIRateLimitMiddleware() fiber.Handler {
	return RateLimitMiddleware(apiLimiter)
}

// SDKRateLimitMiddleware applies SDK generation rate limiting
func SDKRateLimitMiddleware() fiber.Handler {
	return RateLimitMiddleware(sdkLimiter)
}

// PublicAPIRateLimitMiddleware applies public API rate limiting
func PublicAPIRateLimitMiddleware() fiber.Handler {
	return RateLimitMiddleware(publicAPILimiter)
}

// CustomRateLimitMiddleware creates a middleware with custom limits
func CustomRateLimitMiddleware(limit int, window time.Duration) fiber.Handler {
	limiter := NewRateLimiter(limit, window)
	return RateLimitMiddleware(limiter)
}

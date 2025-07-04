package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// RedisRateLimiterConfig defines the config for Redis-based rate limiter
type RedisRateLimiterConfig struct {
	RedisClient    *redis.Client
	KeyPrefix      string
	RequestsPerMin int
	WindowSize     time.Duration
	Logger         *zap.Logger
}

// DefaultRedisRateLimiterConfig is the default Redis rate limiter configuration
var DefaultRedisRateLimiterConfig = RedisRateLimiterConfig{
	KeyPrefix:      "rate_limit:",
	RequestsPerMin: 60,
	WindowSize:     time.Minute,
}

// RedisRateLimiterMiddleware creates a Redis-based rate limiter middleware
func RedisRateLimiterMiddleware(config RedisRateLimiterConfig) fiber.Handler {
	// Set defaults if not provided
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultRedisRateLimiterConfig.KeyPrefix
	}
	if config.RequestsPerMin == 0 {
		config.RequestsPerMin = DefaultRedisRateLimiterConfig.RequestsPerMin
	}
	if config.WindowSize == 0 {
		config.WindowSize = DefaultRedisRateLimiterConfig.WindowSize
	}

	return func(c fiber.Ctx) error {
		// Get client IP
		clientIP := c.IP()
		
		// Create rate limit key
		key := fmt.Sprintf("%s%s", config.KeyPrefix, clientIP)
		
		// Get current time window
		now := time.Now()
		window := now.Truncate(config.WindowSize).Unix()
		windowKey := fmt.Sprintf("%s:%d", key, window)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Use Redis pipeline for atomic operations
		pipe := config.RedisClient.Pipeline()
		
		// Increment counter for current window
		incrCmd := pipe.Incr(ctx, windowKey)
		
		// Set expiration for the key (2 * window size to handle edge cases)
		pipe.Expire(ctx, windowKey, config.WindowSize*2)
		
		// Execute pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Error("Redis rate limiter error", zap.Error(err))
			}
			// If Redis is down, allow the request (fail open)
			return c.Next()
		}
		
		// Get the current count
		count := incrCmd.Val()
		
		// Check if limit exceeded
		if count > int64(config.RequestsPerMin) {
			// Calculate reset time
			resetTime := time.Unix(window, 0).Add(config.WindowSize)
			
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMin))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %d per %v", config.RequestsPerMin, config.WindowSize),
				"retry_after": resetTime.Sub(now).Seconds(),
			})
		}
		
		// Set rate limit headers
		remaining := config.RequestsPerMin - int(count)
		if remaining < 0 {
			remaining = 0
		}
		
		resetTime := time.Unix(window, 0).Add(config.WindowSize)
		c.Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMin))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		
		return c.Next()
	}
}

// UserBasedRedisRateLimiterMiddleware creates a user-based Redis rate limiter
func UserBasedRedisRateLimiterMiddleware(config RedisRateLimiterConfig) fiber.Handler {
	// Set defaults if not provided
	if config.KeyPrefix == "" {
		config.KeyPrefix = "user_rate_limit:"
	}
	if config.RequestsPerMin == 0 {
		config.RequestsPerMin = 100 // Higher limit for authenticated users
	}
	if config.WindowSize == 0 {
		config.WindowSize = DefaultRedisRateLimiterConfig.WindowSize
	}

	return func(c fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			// Fall back to IP-based rate limiting
			return RedisRateLimiterMiddleware(config)(c)
		}
		
		// Create rate limit key for user
		key := fmt.Sprintf("%s%v", config.KeyPrefix, userID)
		
		// Get current time window
		now := time.Now()
		window := now.Truncate(config.WindowSize).Unix()
		windowKey := fmt.Sprintf("%s:%d", key, window)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Use Redis pipeline for atomic operations
		pipe := config.RedisClient.Pipeline()
		
		// Increment counter for current window
		incrCmd := pipe.Incr(ctx, windowKey)
		
		// Set expiration for the key
		pipe.Expire(ctx, windowKey, config.WindowSize*2)
		
		// Execute pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Error("Redis user rate limiter error", zap.Error(err))
			}
			// If Redis is down, allow the request (fail open)
			return c.Next()
		}
		
		// Get the current count
		count := incrCmd.Val()
		
		// Check if limit exceeded
		if count > int64(config.RequestsPerMin) {
			// Calculate reset time
			resetTime := time.Unix(window, 0).Add(config.WindowSize)
			
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMin))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %d per %v", config.RequestsPerMin, config.WindowSize),
				"retry_after": resetTime.Sub(now).Seconds(),
			})
		}
		
		// Set rate limit headers
		remaining := config.RequestsPerMin - int(count)
		if remaining < 0 {
			remaining = 0
		}
		
		resetTime := time.Unix(window, 0).Add(config.WindowSize)
		c.Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMin))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		
		return c.Next()
	}
}

// AdaptiveRedisRateLimiterMiddleware creates an adaptive rate limiter that adjusts based on system load
func AdaptiveRedisRateLimiterMiddleware(config RedisRateLimiterConfig, getSystemLoad func() float64) fiber.Handler {
	baseLimit := config.RequestsPerMin
	
	return func(c fiber.Ctx) error {
		// Adjust rate limit based on system load
		systemLoad := getSystemLoad()
		adjustedLimit := baseLimit
		
		if systemLoad > 0.8 {
			// High load: reduce limit by 50%
			adjustedLimit = baseLimit / 2
		} else if systemLoad > 0.6 {
			// Medium load: reduce limit by 25%
			adjustedLimit = (baseLimit * 3) / 4
		}
		
		// Update config with adjusted limit
		adaptiveConfig := config
		adaptiveConfig.RequestsPerMin = adjustedLimit
		
		return RedisRateLimiterMiddleware(adaptiveConfig)(c)
	}
}
package middleware

import (
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// CORSConfig defines the config for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig is the default CORS configuration
var DefaultCORSConfig = CORSConfig{
	AllowOrigins:     []string{"https://api2sdk.com"},
	AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
	AllowCredentials: false,
	ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
	MaxAge:           86400, // 24 hours
}

// DevelopmentCORSConfig is a more permissive CORS configuration for development
var DevelopmentCORSConfig = CORSConfig{
	AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://127.0.0.1:3000", "http://127.0.0.1:8080"},
	AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
	AllowCredentials: true,
	ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-Total-Count"},
	MaxAge:           3600, // 1 hour
}

// ProductionCORSConfig is a more restrictive CORS configuration for production
// Uses allowed origins from environment variables
func GetProductionCORSConfig() CORSConfig {
	// Get allowed origins from environment variables
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	if allowedOriginsStr != "" {
		// Split by comma and trim spaces
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	// Fallback to default if no origins specified
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"https://api2sdk.com"}
	}

	return CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-Total-Count"},
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config ...CORSConfig) fiber.Handler {
	// Set default config
	cfg := DefaultCORSConfig

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Get origin from request
		origin := c.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		if len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*" {
			allowedOrigin = "*"
		} else {
			for _, allowedOriginPattern := range cfg.AllowOrigins {
				if allowedOriginPattern == "*" || allowedOriginPattern == origin {
					allowedOrigin = origin
					break
				}
				// Support wildcard subdomains
				if strings.HasPrefix(allowedOriginPattern, "*.") {
					domain := strings.TrimPrefix(allowedOriginPattern, "*.")
					if strings.HasSuffix(origin, domain) {
						allowedOrigin = origin
						break
					}
				}
			}
		}

		// Set CORS headers
		if allowedOrigin != "" {
			c.Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		// Set credentials header
		if cfg.AllowCredentials {
			c.Set("Access-Control-Allow-Credentials", "true")
		}

		// Set allowed methods
		if len(cfg.AllowMethods) > 0 {
			c.Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		}

		// Set allowed headers
		if len(cfg.AllowHeaders) > 0 {
			c.Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		}

		// Set exposed headers
		if len(cfg.ExposeHeaders) > 0 {
			c.Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		// Set max age
		if cfg.MaxAge > 0 {
			c.Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
		}

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// DefaultCORSMiddleware returns CORS middleware with default configuration
func DefaultCORSMiddleware() fiber.Handler {
	return CORSMiddleware(DefaultCORSConfig)
}

// DevelopmentCORSMiddleware returns CORS middleware for development
func DevelopmentCORSMiddleware() fiber.Handler {
	return CORSMiddleware(DevelopmentCORSConfig)
}

// ProductionCORSMiddleware returns CORS middleware for production
func ProductionCORSMiddleware() fiber.Handler {
	return CORSMiddleware(GetProductionCORSConfig())
}

// APICORSMiddleware returns CORS middleware specifically for API endpoints
func APICORSMiddleware() fiber.Handler {
	config := CORSConfig{
		AllowOrigins:     []string{"https://my-trusted-api-consumer.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Tye", "Accept", "Authorization", "X-API-Key"},
		AllowCredentials: false,
		ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-Total-Count"},
		MaxAge:           3600,
	}
	return CORSMiddleware(config)
}

package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// SecurityHeadersConfig defines the config for security headers middleware
type SecurityHeadersConfig struct {
	ContentSecurityPolicy   string
	StrictTransportSecurity string
	XFrameOptions           string
	XContentTypeOptions     string
	ReferrerPolicy          string
	PermissionsPolicy       string
	CrossOriginOpenerPolicy string
}

// DefaultSecurityHeadersConfig is the default security headers configuration
var DefaultSecurityHeadersConfig = SecurityHeadersConfig{
	ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';",
	StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
	XFrameOptions:           "DENY",
	XContentTypeOptions:     "nosniff",
	ReferrerPolicy:          "strict-origin-when-cross-origin",
	PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
	CrossOriginOpenerPolicy: "same-origin",
}

// ProductionSecurityHeadersConfig is a more restrictive configuration for production
var ProductionSecurityHeadersConfig = SecurityHeadersConfig{
	ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';",
	StrictTransportSecurity: "max-age=63072000; includeSubDomains; preload",
	XFrameOptions:           "DENY",
	XContentTypeOptions:     "nosniff",
	ReferrerPolicy:          "strict-origin-when-cross-origin",
	PermissionsPolicy:       "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
	CrossOriginOpenerPolicy: "same-origin",
}

// SecurityHeadersMiddleware creates a security headers middleware with the default configuration
func SecurityHeadersMiddleware() fiber.Handler {
	// Use default config
	cfg := DefaultSecurityHeadersConfig

	return func(c fiber.Ctx) error {
		// Set Content Security Policy
		if cfg.ContentSecurityPolicy != "" {
			c.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		// Set Strict Transport Security (HSTS)
		if cfg.StrictTransportSecurity != "" {
			c.Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
		}

		// Set X-Frame-Options
		if cfg.XFrameOptions != "" {
			c.Set("X-Frame-Options", cfg.XFrameOptions)
		}

		// Set X-Content-Type-Options
		if cfg.XContentTypeOptions != "" {
			c.Set("X-Content-Type-Options", cfg.XContentTypeOptions)
		}

		// Set Referrer Policy
		if cfg.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", cfg.ReferrerPolicy)
		}

		// Set Permissions Policy
		if cfg.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", cfg.PermissionsPolicy)
		}

		// Set Cross-Origin-Opener-Policy
		if cfg.CrossOriginOpenerPolicy != "" {
			c.Set("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
		}

		// Additional security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-DNS-Prefetch-Control", "off")
		c.Set("X-Download-Options", "noopen")
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		return c.Next()
	}
}

// DefaultSecurityHeadersMiddleware returns security headers middleware with default configuration
func DefaultSecurityHeadersMiddleware() fiber.Handler {
	return SecurityHeadersMiddleware()
}

// ProductionSecurityHeadersMiddleware returns security headers middleware for production
func ProductionSecurityHeadersMiddleware() fiber.Handler {
	cfg := ProductionSecurityHeadersConfig
	return func(c fiber.Ctx) error {
		// Set Content Security Policy
		if cfg.ContentSecurityPolicy != "" {
			c.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		// Set Strict Transport Security (HSTS)
		if cfg.StrictTransportSecurity != "" {
			c.Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
		}

		// Set X-Frame-Options
		if cfg.XFrameOptions != "" {
			c.Set("X-Frame-Options", cfg.XFrameOptions)
		}

		// Set X-Content-Type-Options
		if cfg.XContentTypeOptions != "" {
			c.Set("X-Content-Type-Options", cfg.XContentTypeOptions)
		}

		// Set Referrer Policy
		if cfg.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", cfg.ReferrerPolicy)
		}

		// Set Permissions Policy
		if cfg.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", cfg.PermissionsPolicy)
		}

		// Set Cross-Origin-Opener-Policy
		if cfg.CrossOriginOpenerPolicy != "" {
			c.Set("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
		}

		// Additional security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-DNS-Prefetch-Control", "off")
		c.Set("X-Download-Options", "noopen")
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		return c.Next()
	}
}

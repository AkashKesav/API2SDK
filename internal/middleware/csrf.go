package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// CSRFConfig defines the config for CSRF middleware
type CSRFConfig struct {
	TokenLength    int
	TokenLookup    string // "header:X-CSRF-Token" or "form:csrf_token" or "query:csrf_token"
	CookieName     string
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
	Expiration     time.Duration
	KeyGenerator   func() (string, error)
	Logger         *zap.Logger
}

// DefaultCSRFConfig is the default CSRF configuration
var DefaultCSRFConfig = CSRFConfig{
	TokenLength:    32,
	TokenLookup:    "header:X-CSRF-Token",
	CookieName:     "csrf_token",
	CookiePath:     "/",
	CookieSecure:   true,
	CookieHTTPOnly: true,
	CookieSameSite: "Strict",
	Expiration:     24 * time.Hour,
	KeyGenerator:   generateCSRFToken,
}

// CSRFMiddleware creates a CSRF protection middleware
func CSRFMiddleware(config ...CSRFConfig) fiber.Handler {
	// Set default config
	cfg := DefaultCSRFConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default key generator if not provided
	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = generateCSRFToken
	}

	return func(c fiber.Ctx) error {
		// Skip CSRF for safe methods
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Get token from cookie
		cookieToken := c.Cookies(cfg.CookieName)
		
		// Get token from request (header, form, or query)
		requestToken := extractTokenFromRequest(c, cfg.TokenLookup)
		
		// If no cookie token exists, this might be the first request
		if cookieToken == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token missing",
			})
		}
		
		// If no request token provided
		if requestToken == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token required",
			})
		}
		
		// Validate tokens match
		if !validateCSRFToken(cookieToken, requestToken) {
			if cfg.Logger != nil {
				cfg.Logger.Warn("CSRF token validation failed",
					zap.String("ip", c.IP()),
					zap.String("user_agent", c.Get("User-Agent")),
					zap.String("path", c.Path()),
				)
			}
			
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token invalid",
			})
		}
		
		return c.Next()
	}
}

// CSRFTokenGeneratorMiddleware generates and sets CSRF tokens
func CSRFTokenGeneratorMiddleware(config ...CSRFConfig) fiber.Handler {
	// Set default config
	cfg := DefaultCSRFConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Check if token already exists
		existingToken := c.Cookies(cfg.CookieName)
		
		var token string
		var err error
		
		if existingToken == "" {
			// Generate new token
			token, err = cfg.KeyGenerator()
			if err != nil {
				if cfg.Logger != nil {
					cfg.Logger.Error("Failed to generate CSRF token", zap.Error(err))
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate CSRF token",
				})
			}
			
			// Set cookie
			c.Cookie(&fiber.Cookie{
				Name:     cfg.CookieName,
				Value:    token,
				Path:     cfg.CookiePath,
				Domain:   cfg.CookieDomain,
				MaxAge:   int(cfg.Expiration.Seconds()),
				Secure:   cfg.CookieSecure,
				HTTPOnly: cfg.CookieHTTPOnly,
				SameSite: cfg.CookieSameSite,
			})
		} else {
			token = existingToken
		}
		
		// Set token in response header for client access
		c.Set("X-CSRF-Token", token)
		
		// Store token in context for use in templates
		c.Locals("csrf_token", token)
		
		return c.Next()
	}
}

// extractTokenFromRequest extracts CSRF token from request based on lookup configuration
func extractTokenFromRequest(c fiber.Ctx, lookup string) string {
	parts := strings.Split(lookup, ":")
	if len(parts) != 2 {
		return ""
	}
	
	switch parts[0] {
	case "header":
		return c.Get(parts[1])
	case "form":
		return c.FormValue(parts[1])
	case "query":
		return c.Query(parts[1])
	default:
		return ""
	}
}

// validateCSRFToken validates CSRF tokens using constant-time comparison
func validateCSRFToken(cookieToken, requestToken string) bool {
	// Decode tokens
	cookieBytes, err1 := base64.StdEncoding.DecodeString(cookieToken)
	requestBytes, err2 := base64.StdEncoding.DecodeString(requestToken)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(cookieBytes, requestBytes) == 1
}

// generateCSRFToken generates a cryptographically secure CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetCSRFToken returns the CSRF token for the current request
func GetCSRFToken(c fiber.Ctx) string {
	if token := c.Locals("csrf_token"); token != nil {
		return token.(string)
	}
	return ""
}

// DefaultCSRFMiddleware returns CSRF middleware with default configuration
func DefaultCSRFMiddleware() fiber.Handler {
	return CSRFMiddleware(DefaultCSRFConfig)
}

// StrictCSRFMiddleware returns CSRF middleware with strict configuration
func StrictCSRFMiddleware() fiber.Handler {
	strictConfig := DefaultCSRFConfig
	strictConfig.CookieSecure = true
	strictConfig.CookieHTTPOnly = true
	strictConfig.CookieSameSite = "Strict"
	strictConfig.Expiration = 1 * time.Hour // Shorter expiration
	
	return CSRFMiddleware(strictConfig)
}
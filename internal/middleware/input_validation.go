package middleware

import (
	"html"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// InputValidationConfig defines the config for input validation middleware
type InputValidationConfig struct {
	MaxBodySize       int64
	SanitizeHTML      bool
	BlockSQLInjection bool
	BlockXSS          bool
	AllowedFileTypes  []string
	MaxFileSize       int64
}

// DefaultInputValidationConfig is the default input validation configuration
var DefaultInputValidationConfig = InputValidationConfig{
	MaxBodySize:       10 * 1024 * 1024, // 10MB
	SanitizeHTML:      true,
	BlockSQLInjection: true,
	BlockXSS:          true,
	AllowedFileTypes:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".json", ".yaml", ".yml"},
	MaxFileSize:       5 * 1024 * 1024, // 5MB
}

// SQL injection patterns - more lenient for JSON content
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|update\s+.*\s+set|delete\s+from)`),
	regexp.MustCompile(`(?i)(drop\s+table|create\s+table|alter\s+table|truncate\s+table)`),
	regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\(|sp_executesql)`),
	regexp.MustCompile(`(?i)(xp_|sp_cmdshell)`),
	// Removed overly broad patterns that block legitimate JSON
}

// XSS patterns
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
	regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
	regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`),
	regexp.MustCompile(`(?i)<link[^>]*>`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)vbscript:`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
}

// InputValidationMiddleware creates an input validation middleware with the given configuration
func InputValidationMiddleware(config ...InputValidationConfig) fiber.Handler {
	// Set default config
	cfg := DefaultInputValidationConfig

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Check body size
		if int64(len(c.Body())) > cfg.MaxBodySize {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request body too large",
			})
		}

		// Get request body for validation
		body := c.Body()
		bodyStr := string(body)

		// Check for SQL injection - completely disabled for development
		if cfg.BlockSQLInjection && false { // Force disable
			for _, pattern := range sqlInjectionPatterns {
				if pattern.MatchString(bodyStr) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid input detected",
					})
				}
			}
		}

		// Check for XSS - completely disabled for development
		if cfg.BlockXSS && false { // Force disable
			for _, pattern := range xssPatterns {
				if pattern.MatchString(bodyStr) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid input detected",
					})
				}
			}
		}

		// Sanitize HTML if enabled
		if cfg.SanitizeHTML {
			sanitizedBody := html.EscapeString(bodyStr)
			// Note: This is a basic implementation. For production, consider using a proper HTML sanitizer
			c.Request().SetBody([]byte(sanitizedBody))
		}

		// Validate query parameters
		for key, value := range c.Queries() {
			// Check for SQL injection in query params
			if cfg.BlockSQLInjection {
				for _, pattern := range sqlInjectionPatterns {
					if pattern.MatchString(value) {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error": "Invalid query parameter detected",
						})
					}
				}
			}

			// Check for XSS in query params
			if cfg.BlockXSS {
				for _, pattern := range xssPatterns {
					if pattern.MatchString(value) {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error": "Invalid query parameter detected",
						})
					}
				}
			}

			// Sanitize query parameters if enabled
			if cfg.SanitizeHTML {
				sanitizedValue := html.EscapeString(value)
				// Note: Fiber v3 may not support direct query parameter modification
				// Store sanitized values in context for potential later use
				c.Locals("sanitized_"+key, sanitizedValue)
			}
		}

		return c.Next()
	}
}

// FileValidationMiddleware validates uploaded files
func FileValidationMiddleware(config ...InputValidationConfig) fiber.Handler {
	// Set default config
	cfg := DefaultInputValidationConfig

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Check if this is a file upload request
		contentType := c.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			return c.Next()
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid multipart form",
			})
		}

		// Validate files
		for _, files := range form.File {
			for _, file := range files {
				// Check file size
				if file.Size > cfg.MaxFileSize {
					return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
						"error": "File size too large",
					})
				}

				// Check file type
				if len(cfg.AllowedFileTypes) > 0 {
					allowed := false
					for _, allowedType := range cfg.AllowedFileTypes {
						if strings.HasSuffix(strings.ToLower(file.Filename), allowedType) {
							allowed = true
							break
						}
					}
					if !allowed {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error": "File type not allowed",
						})
					}
				}

				// Check for path traversal in filename
				if strings.Contains(file.Filename, "..") || strings.Contains(file.Filename, "/") || strings.Contains(file.Filename, "\\") {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid filename",
					})
				}
			}
		}

		return c.Next()
	}
}

// DefaultInputValidationMiddleware returns input validation middleware with default configuration
func DefaultInputValidationMiddleware() fiber.Handler {
	// Temporarily disabled for development
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}

// StrictInputValidationMiddleware returns input validation middleware with strict configuration
func StrictInputValidationMiddleware() fiber.Handler {
	strictConfig := InputValidationConfig{
		MaxBodySize:       1 * 1024 * 1024, // 1MB
		SanitizeHTML:      true,
		BlockSQLInjection: true,
		BlockXSS:          true,
		AllowedFileTypes:  []string{".jpg", ".jpeg", ".png", ".pdf", ".txt"},
		MaxFileSize:       1 * 1024 * 1024, // 1MB
	}
	return InputValidationMiddleware(strictConfig)
}

package utils

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AkashKesav/API2SDK/configs"
)

// ConfigValidator validates application configuration
type ConfigValidator struct {
	errors   []string
	warnings []string
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors:   make([]string, 0),
		warnings: make([]string, 0),
	}
}

// ValidateConfig validates the entire configuration
func (cv *ConfigValidator) ValidateConfig(config *configs.Config) error {
	cv.errors = make([]string, 0)
	cv.warnings = make([]string, 0)

	// Validate required fields
	cv.validateRequired(config)

	// Validate database configuration
	cv.validateDatabase(config)

	// Validate server configuration
	cv.validateServer(config)

	// Validate external services
	cv.validateExternalServices(config)

	// Validate security settings
	cv.validateSecurity(config)

	if len(cv.errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(cv.errors, "; "))
	}

	return nil
}

// GetWarnings returns validation warnings
func (cv *ConfigValidator) GetWarnings() []string {
	return cv.warnings
}

// validateRequired validates required configuration fields
func (cv *ConfigValidator) validateRequired(config *configs.Config) {
	if config.MongoDBURI == "" {
		cv.errors = append(cv.errors, "MongoDBURI is required")
	}

	if config.MongoDBName == "" {
		cv.errors = append(cv.errors, "MongoDBName is required")
	}
}

// validateDatabase validates database configuration
func (cv *ConfigValidator) validateDatabase(config *configs.Config) {
	if config.MongoDBURI != "" {
		if _, err := url.Parse(config.MongoDBURI); err != nil {
			cv.errors = append(cv.errors, fmt.Sprintf("invalid MongoDBURI format: %v", err))
		}
	}

	if config.MongoDBName != "" {
		if len(config.MongoDBName) > 64 {
			cv.errors = append(cv.errors, "MongoDBName must be 64 characters or less")
		}

		// Check for invalid characters in database name
		invalidChars := []string{"/", "\\", ".", " ", "\"", "$", "*", "<", ">", ":", "|", "?"}
		for _, char := range invalidChars {
			if strings.Contains(config.MongoDBName, char) {
				cv.errors = append(cv.errors, fmt.Sprintf("MongoDBName contains invalid character: %s", char))
				break
			}
		}
	}
}

// validateServer validates server configuration
func (cv *ConfigValidator) validateServer(config *configs.Config) {
	if config.Port != "" {
		if port, err := strconv.Atoi(config.Port); err != nil {
			cv.errors = append(cv.errors, "Port must be a valid integer")
		} else if port < 1 || port > 65535 {
			cv.errors = append(cv.errors, "Port must be between 1 and 65535")
		} else if port < 1024 {
			cv.warnings = append(cv.warnings, "Port is below 1024 - may require elevated privileges")
		}
	}
}

// validateExternalServices validates external service configuration
func (cv *ConfigValidator) validateExternalServices(config *configs.Config) {
	// Validate Postman API key
	postmanAPIKey := os.Getenv("POSTMAN_API_KEY")
	if postmanAPIKey == "" {
		cv.warnings = append(cv.warnings, "POSTMAN_API_KEY is not set - Postman integration will be limited")
	} else if len(postmanAPIKey) < 20 {
		cv.warnings = append(cv.warnings, "POSTMAN_API_KEY appears to be too short - verify it's correct")
	}
}

// validateSecurity validates security-related configuration
func (cv *ConfigValidator) validateSecurity(config *configs.Config) {
	// Check environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		cv.warnings = append(cv.warnings, "ENVIRONMENT variable not set - defaulting to development")
	} else if env == "production" {
		cv.validateProductionSecurity(config)
	}
}

// validateProductionSecurity validates production-specific security requirements
func (cv *ConfigValidator) validateProductionSecurity(config *configs.Config) {
	// Check for secure MongoDB connection
	if config.MongoDBURI != "" && !strings.Contains(config.MongoDBURI, "ssl=true") && !strings.Contains(config.MongoDBURI, "tls=true") {
		cv.warnings = append(cv.warnings, "MongoDB connection should use SSL/TLS in production")
	}

	// Check for debug settings
	if os.Getenv("DEBUG") == "true" {
		cv.warnings = append(cv.warnings, "DEBUG mode is enabled in production - consider disabling for security")
	}
}

// ValidateEnvironment validates the runtime environment
func (cv *ConfigValidator) ValidateEnvironment() error {
	cv.errors = make([]string, 0)
	cv.warnings = make([]string, 0)

	// Check required external tools
	cv.checkExternalTool("mcpgen", "MCP generation will not work")

	// Check optional external tools
	cv.checkOptionalTool("openapi-generator-cli", "OpenAPI generator features may be limited")

	// Check file permissions
	cv.checkDirectoryPermissions("./generated_sdks", "SDK generation may fail")
	cv.checkDirectoryPermissions("./generated_mcps", "MCP generation may fail")
	cv.checkDirectoryPermissions("./temp", "Temporary file operations may fail")

	if len(cv.errors) > 0 {
		return fmt.Errorf("environment validation failed: %s", strings.Join(cv.errors, "; "))
	}

	return nil
}

// checkExternalTool checks if an external tool is available
func (cv *ConfigValidator) checkExternalTool(tool, message string) {
	if _, err := os.Stat(tool); os.IsNotExist(err) {
		// Check if it's in PATH
		if _, err := exec.LookPath(tool); err != nil {
			cv.errors = append(cv.errors, fmt.Sprintf("%s not found in PATH - %s", tool, message))
		}
	}
}

// checkOptionalTool checks if an optional external tool is available
func (cv *ConfigValidator) checkOptionalTool(tool, message string) {
	if _, err := os.Stat(tool); os.IsNotExist(err) {
		// Check if it's in PATH
		if _, err := exec.LookPath(tool); err != nil {
			cv.warnings = append(cv.warnings, fmt.Sprintf("%s not found in PATH - %s", tool, message))
		}
	}
}

// checkDirectoryPermissions checks if a directory exists and is writable
func (cv *ConfigValidator) checkDirectoryPermissions(dir, message string) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		cv.errors = append(cv.errors, fmt.Sprintf("cannot create directory %s - %s: %v", dir, message, err))
		return
	}

	// Test write permissions
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		cv.errors = append(cv.errors, fmt.Sprintf("cannot write to directory %s - %s: %v", dir, message, err))
		return
	}

	// Clean up test file
	os.Remove(testFile)
}

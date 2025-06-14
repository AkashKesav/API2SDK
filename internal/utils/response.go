package utils

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v3"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse returns a success response
func SuccessResponse(c fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c fiber.Ctx, statusCode int, message string, errorDetail string) error {
	response := APIResponse{
		Success: false,
		Message: message,
	}

	if errorDetail != "" {
		response.Error = errorDetail
	}

	return c.Status(statusCode).JSON(response)
}

// CreatedResponse returns a 201 created response
func CreatedResponse(c fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContentResponse returns a 204 no content response
func NoContentResponse(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// BadRequestResponse returns a 400 bad request response
func BadRequestResponse(c fiber.Ctx, message string, errorDetail string) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message, errorDetail)
}

// UnauthorizedResponse returns a 401 unauthorized response
func UnauthorizedResponse(c fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, message, "")
}

// ForbiddenResponse returns a 403 forbidden response
func ForbiddenResponse(c fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusForbidden, message, "")
}

// NotFoundResponse returns a 404 not found response
func NotFoundResponse(c fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, message, "")
}

// InternalServerErrorResponse returns a 500 internal server error response
func InternalServerErrorResponse(c fiber.Ctx, message string, errorDetail string) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, message, errorDetail)
}

// ValidationErrorResponse returns a 422 validation error response
func ValidationErrorResponse(c fiber.Ctx, message string, errors interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(APIResponse{
		Success: false,
		Message: message,
		Data:    errors,
	})
}

// EnsureDir creates a directory if it does not already exist.
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}
	return nil
}

// ConvertToPascalCase converts a string to PascalCase.
// Example: "my-package-name" -> "MyPackageName"
// Example: "my_package_name" -> "MyPackageName"
func ConvertToPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// ConvertToKebabCase converts a string to kebab-case.
// Example: "MyPackageName" -> "my-package-name"
// Example: "my_package_name" -> "my-package-name"
// Example: "My Package Name" -> "my-package-name"
func ConvertToKebabCase(s string) string {
	// Replace underscores and spaces with hyphens
	replaced := strings.ReplaceAll(s, "_", "-")
	replaced = strings.ReplaceAll(replaced, " ", "-")

	// Handle CamelCase by inserting hyphens
	var result strings.Builder
	for i, r := range replaced {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(replaced[i-1])) || unicode.IsDigit(rune(replaced[i-1]))) {
			// Add hyphen if current is uppercase and previous is lowercase/digit
			// and also ensure we don't add hyphen if previous was already a hyphen (e.g. from original string)
			if replaced[i-1] != '-' {
				result.WriteRune('-')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	// Remove consecutive hyphens that might have been formed
	final := strings.ReplaceAll(result.String(), "--", "-")
	return final
}

// GetOrgFromPkg extracts the organization part from a package name string.
// If no separator '/' is found, it returns the original string.
// Example: "my-org/my-repo" -> "my-org"
// Example: "my-repo" -> "my-repo"
func GetOrgFromPkg(packageName string) string {
	if idx := strings.Index(packageName, "/"); idx != -1 {
		return packageName[:idx]
	}
	return packageName
}

// GetNameFromPkg extracts the name part from a package name string.
// If no separator '/' is found, it returns the original string.
// Example: "my-org/my-repo" -> "my-repo"
// Example: "my-repo" -> "my-repo"
func GetNameFromPkg(packageName string) string {
	if idx := strings.Index(packageName, "/"); idx != -1 {
		return packageName[idx+1:]
	}
	return packageName
}

// ConvertToAlphanumeric converts a string to contain only alphanumeric characters.
// Non-alphanumeric characters are removed. If the resulting string is empty,
// it returns the provided defaultVal.
func ConvertToAlphanumeric(s string, defaultVal string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	if result.Len() == 0 {
		return defaultVal
	}
	return result.String()
}

// ConvertToSnakeCase converts a string to snake_case.
// Example: "MyPackageName" -> "my_package_name"
// Example: "my-package-name" -> "my_package_name"
// Example: "my_package_name" -> "my_package_name"
// Example: "My Package Name" -> "my_package_name"
// Example: "myOrg/myClient" -> "my_org_my_client"
func ConvertToSnakeCase(s string) string {
	// Replace common separators with underscore
	replaced := strings.ReplaceAll(s, "-", "_")
	replaced = strings.ReplaceAll(replaced, " ", "_")
	replaced = strings.ReplaceAll(replaced, "/", "_")

	var result strings.Builder
	for i, r := range replaced {
		// If it's an uppercase letter
		if unicode.IsUpper(r) {
			// Add an underscore if it's not the first character,
			// and the previous character is not already an underscore,
			// and the previous character is a letter or digit (to avoid underscore after underscore like in "MY_APP" -> "m_y__a_p_p")
			if i > 0 && replaced[i-1] != '_' && (unicode.IsLower(rune(replaced[i-1])) || unicode.IsDigit(rune(replaced[i-1]))) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r) // Keep it as is (already lowercase or underscore or digit)
		}
	}
	// Remove potential consecutive underscores
	final := strings.ReplaceAll(result.String(), "__", "_")
	// Remove leading/trailing underscores that might form
	final = strings.Trim(final, "_")
	return final
}

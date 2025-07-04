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

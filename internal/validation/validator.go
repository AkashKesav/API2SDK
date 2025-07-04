package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Validator interface for input validation
type Validator interface {
	Validate(ctx context.Context, input interface{}) error
	ValidateStruct(input interface{}) error
	RegisterCustomValidation(tag string, fn validator.Func) error
}

// CustomValidator implements the Validator interface
type CustomValidator struct {
	validator *validator.Validate
	logger    *zap.Logger
}

// NewCustomValidator creates a new validator instance
func NewCustomValidator(logger *zap.Logger) *CustomValidator {
	v := validator.New()

	cv := &CustomValidator{
		validator: v,
		logger:    logger,
	}

	// Register custom validations
	cv.registerCustomValidations()

	return cv
}

// Validate validates input with context
func (cv *CustomValidator) Validate(ctx context.Context, input interface{}) error {
	return cv.ValidateStruct(input)
}

// ValidateStruct validates a struct
func (cv *CustomValidator) ValidateStruct(input interface{}) error {
	err := cv.validator.Struct(input)
	if err != nil {
		return cv.formatValidationError(err)
	}
	return nil
}

// RegisterCustomValidation registers a custom validation function
func (cv *CustomValidator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// registerCustomValidations registers all custom validation rules
func (cv *CustomValidator) registerCustomValidations() {
	// Password strength validation
	cv.validator.RegisterValidation("password_strength", cv.validatePasswordStrength)

	// Alpha with spaces validation
	cv.validator.RegisterValidation("alpha_space", cv.validateAlphaSpace)

	// MongoDB ObjectID validation
	cv.validator.RegisterValidation("mongodb_id", cv.validateMongoID)

	// Phone number validation
	cv.validator.RegisterValidation("phone", cv.validatePhoneNumber)

	// URL validation with specific schemes
	cv.validator.RegisterValidation("secure_url", cv.validateSecureURL)

	// Custom email validation (more strict)
	cv.validator.RegisterValidation("business_email", cv.validateBusinessEmail)

	// File extension validation
	cv.validator.RegisterValidation("file_ext", cv.validateFileExtension)

	// JSON validation
	cv.validator.RegisterValidation("json", cv.validateJSON)
}

// validatePasswordStrength validates password strength
func (cv *CustomValidator) validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateAlphaSpace validates alphabetic characters with spaces
func (cv *CustomValidator) validateAlphaSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	alphaSpaceRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
	return alphaSpaceRegex.MatchString(value)
}

// validateMongoID validates MongoDB ObjectID format
func (cv *CustomValidator) validateMongoID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	mongoIDRegex := regexp.MustCompile(`^[a-fA-F0-9]{24}$`)
	return mongoIDRegex.MatchString(value)
}

// validatePhoneNumber validates phone number format
func (cv *CustomValidator) validatePhoneNumber(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(value)
}

// validateSecureURL validates HTTPS URLs
func (cv *CustomValidator) validateSecureURL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.HasPrefix(value, "https://")
}

// validateBusinessEmail validates business email addresses (no common free providers)
func (cv *CustomValidator) validateBusinessEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Basic email validation first
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Check against common free email providers
	freeProviders := []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com",
		"aol.com", "icloud.com", "protonmail.com", "mail.com",
	}

	domain := strings.ToLower(strings.Split(email, "@")[1])
	for _, provider := range freeProviders {
		if domain == provider {
			return false
		}
	}

	return true
}

// validateFileExtension validates file extensions
func (cv *CustomValidator) validateFileExtension(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	allowedExts := strings.Split(fl.Param(), "|")

	for _, ext := range allowedExts {
		if strings.HasSuffix(strings.ToLower(filename), "."+ext) {
			return true
		}
	}

	return false
}

// validateJSON validates JSON format
func (cv *CustomValidator) validateJSON(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Simple JSON validation - starts with { or [
	trimmed := strings.TrimSpace(value)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// formatValidationError formats validation errors into a user-friendly format
func (cv *CustomValidator) formatValidationError(err error) error {
	var errors []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, cv.formatFieldError(fieldError))
		}
	}

	return &ValidationError{
		Message: "Validation failed",
		Errors:  errors,
	}
}

// formatFieldError formats a single field validation error
func (cv *CustomValidator) formatFieldError(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "password_strength":
		return fmt.Sprintf("%s must contain at least 8 characters with uppercase, lowercase, number, and special character", field)
	case "alpha_space":
		return fmt.Sprintf("%s must contain only alphabetic characters and spaces", field)
	case "mongodb_id":
		return fmt.Sprintf("%s must be a valid MongoDB ObjectID", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "secure_url":
		return fmt.Sprintf("%s must be a secure HTTPS URL", field)
	case "business_email":
		return fmt.Sprintf("%s must be a business email address", field)
	case "file_ext":
		return fmt.Sprintf("%s must have one of the following extensions: %s", field, param)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Errors, ", "))
}

// Request validation structs with comprehensive validation tags

// CreateUserRequest validates user creation input
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100,alpha_space"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,password_strength"`
}

// UpdateUserRequest validates user update input
type UpdateUserRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100,alpha_space"`
	Email string `json:"email,omitempty" validate:"omitempty,email,max=255"`
}

// CreateCollectionRequest validates collection creation input
type CreateCollectionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	PostmanData string `json:"postmanData" validate:"required,json"`
}

// SDKGenerationRequest validates SDK generation input
type SDKGenerationRequest struct {
	CollectionID string `json:"collectionId" validate:"required,mongodb_id"`
	Language     string `json:"language" validate:"required,oneof=python javascript php go java"`
	PackageName  string `json:"packageName" validate:"required,min=1,max=100"`
	Version      string `json:"version,omitempty" validate:"omitempty,semver"`
}

// MCPGenerationRequest validates MCP generation input
type MCPGenerationRequest struct {
	CollectionID string `json:"collectionId" validate:"required,mongodb_id"`
	Transport    string `json:"transport" validate:"required,oneof=stdio web streamable-http"`
	Port         int    `json:"port,omitempty" validate:"omitempty,min=1024,max=65535"`
}

// LoginRequest validates login input
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// ChangePasswordRequest validates password change input
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,password_strength"`
}

// FileUploadRequest validates file upload input
type FileUploadRequest struct {
	Filename string `json:"filename" validate:"required,file_ext=json|yaml|yml"`
	Content  string `json:"content" validate:"required"`
}

// Example usage in a controller
func ExampleValidation(validator Validator) {
	req := CreateUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	if err := validator.ValidateStruct(req); err != nil {
		// Handle validation error
		fmt.Printf("Validation failed: %v\n", err)
	}
}

package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/google/uuid"
)

var (
	nonAlphanumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
	leadingDigitRegex    = regexp.MustCompile("^[0-9]+")
)

// SanitizeNameForIdentifier converts a string into a more suitable identifier,
// removing non-alphanumeric characters and handling leading digits.
// It also converts to lowercase by default.
func SanitizeNameForIdentifier(name string, toLower bool) string {
	if name == "" {
		return "unnamed"
	}
	if toLower {
		name = strings.ToLower(name)
	}
	sanitized := nonAlphanumericRegex.ReplaceAllString(name, "")
	if sanitized == "" { // If all characters were non-alphanumeric
		return "generated" // fallback
	}
	// Ensure it doesn't start with a digit for many languages
	if leadingDigitRegex.MatchString(sanitized) {
		sanitized = "pkg_" + sanitized
	}
	return sanitized
}

// DerivePackageName creates a suitable package name from a collection and language.
func DerivePackageName(collection *models.Collection, language string) string {
	baseName := "generated_sdk"
	if collection != nil && collection.Name != "" {
		baseName = collection.Name
	} else if collection != nil && !collection.ID.IsZero() { // Check if ObjectID is not zero
		// Try to use collection ID if name is empty but ID exists
		idHexStr := collection.ID.Hex() // Get hex string representation
		idPartForName := idHexStr
		if len(idHexStr) > 8 { // Take a portion if it's a long ID
			idPartForName = idHexStr[:8]
		}
		baseName = fmt.Sprintf("coll_%s", idPartForName)
	}

	// Language-specific conventions can be added here if needed
	// For now, a generic sanitization.
	// Most package managers prefer lowercase, often with underscores or hyphens.
	// Go packages are typically all lowercase.
	// Java packages are reverse domain.
	// Python modules are lowercase_with_underscores.
	// NPM packages are kebab-case.

	switch strings.ToLower(language) {
	case "go":
		return SanitizeNameForIdentifier(baseName, true)
	case "python":
		return ConvertToSnakeCase(SanitizeNameForIdentifier(baseName, true))
	case "typescript", "javascript":
		return ConvertToKebabCase(SanitizeNameForIdentifier(baseName, true))
	case "java":
		// Example: com.example.generatedsdk.mycollectionname
		orgPart := "com.example" // This could be configurable
		namePart := SanitizeNameForIdentifier(baseName, true)
		return fmt.Sprintf("%s.%s", orgPart, namePart)
	case "csharp", "c#":
		return ConvertToPascalCase(SanitizeNameForIdentifier(baseName, false)) // C# namespaces are PascalCase
	case "ruby":
		return ConvertToSnakeCase(SanitizeNameForIdentifier(baseName, true)) // Gem names are snake_case
	case "php":
		return ConvertToPascalCase(SanitizeNameForIdentifier(baseName, false)) // PHP namespaces are often PascalCase
	default:
		return SanitizeNameForIdentifier(baseName, true)
	}
}

// GenerateUniqueID generates a unique identifier for SDK generation tasks
func GenerateUniqueID() string {
	return fmt.Sprintf("%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

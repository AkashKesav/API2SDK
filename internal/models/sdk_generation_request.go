package models

// SDKGenerationRequest defines the structure for requesting SDK generation.
// Note: This struct is now the canonical definition.
// Ensure all fields are comprehensive and validation tags are correct.
type SDKGenerationRequest struct {
	CollectionID string `json:"collectionId" validate:"required,hexadecimal,len=24"` // Made required, assuming SDK is always from an existing collection
	Language     string `json:"language" validate:"required"`                        // Specific language for this SDK generation (not a list)
	PackageName  string `json:"packageName,omitempty"`                               // Optional: Package name for the SDK
}

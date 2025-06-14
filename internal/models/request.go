package models

// CreatePublicAPIRequest represents the request to create a public API
type CreatePublicAPIRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	BaseURL     string   `json:"base_url" validate:"required"`
	AuthType    string   `json:"auth_type"`
	Tags        []string `json:"tags"`
	PostmanURL  string   `json:"postman_url"`
}

// UpdatePublicAPIRequest represents the request to update a public API
type UpdatePublicAPIRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	BaseURL     string   `json:"base_url"`
	AuthType    string   `json:"auth_type"`
	Tags        []string `json:"tags"`
	PostmanURL  string   `json:"postman_url"`
	IsActive    *bool    `json:"is_active"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `json:"page" query:"page"`
	Limit int `json:"limit" query:"limit"`
}

// SearchRequest represents search parameters
type SearchRequest struct {
	Query    string `json:"query" query:"q"`
	Category string `json:"category" query:"category"`
	Language string `json:"language" query:"language"`
	PaginationRequest
}

// SupportedLanguages defines the languages supported for SDK generation.
// TODO: This could be made configurable or dynamic.
var SupportedLanguages = []string{"go", "typescript", "python", "ruby", "php", "java", "csharp", "rust"}

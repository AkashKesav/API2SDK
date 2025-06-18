package models

// Pagination holds pagination metadata.
type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	Limit       int   `json:"limit"`
	TotalItems  int64 `json:"totalItems"`
	TotalPages  int   `json:"totalPages"`
}

// PaginatedSDKsResponse is the response structure for listing SDKs with pagination.
type PaginatedSDKsResponse struct {
	SDKs       []*SDK     `json:"sdks"`
	Pagination Pagination `json:"pagination"`
}

// SDKStatusResponse defines the structure for SDK status responses.
type SDKStatusResponse struct {
	ID          string              `json:"id"`
	Status      SDKGenerationStatus `json:"status"`
	Error       string              `json:"error,omitempty"`
	DownloadURL string              `json:"downloadUrl,omitempty"`
}

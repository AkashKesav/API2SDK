package models

import "time"

// GenerationRequest represents a request to generate an SDK or MCP server
type GenerationRequest struct {
	CollectionID    string `json:"collection_id" validate:"required"`
	GenerationType  string `json:"generation_type" validate:"required,oneof=sdk mcp"` // "sdk" or "mcp"
	Language        string `json:"language,omitempty" validate:"omitempty,required_if=GenerationType sdk,oneof=go typescript python php java csharp rust ruby"`
	PackageName     string `json:"package_name,omitempty" validate:"omitempty,required_if=GenerationType sdk,min=1,max=100"`
	OutputDirectory string `json:"output_directory,omitempty"`
	MCPTransport    string `json:"mcp_transport,omitempty" validate:"omitempty,required_if=GenerationType mcp,oneof=stdio web streamable-http"`
	MCPPort         int    `json:"mcp_port,omitempty" validate:"omitempty,required_if=MCPTransport web,required_if=MCPTransport streamable-http,min=1024,max=65535"`
	// UserID will be extracted from middleware context, not from request body
}

// GenerationResponse represents the response from SDK or MCP server generation
type GenerationResponse struct {
	Message     string    `json:"message"`
	SDKID       string    `json:"sdk_id,omitempty"`
	Status      string    `json:"status"`
	OutputPath  string    `json:"output_path,omitempty"`
	GeneratedAt time.Time `json:"generated_at,omitempty"`
}

// SDKHistoryRequest represents a request for SDK history with pagination
type SDKHistoryRequest struct {
	Page  int `json:"page,omitempty" query:"page"`
	Limit int `json:"limit,omitempty" query:"limit"`
}

// SDKHistoryResponse represents the response containing SDK history
type SDKHistoryResponse struct {
	SDKs       []*SDK `json:"sdks"`
	TotalCount int64  `json:"total_count"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

// DeleteSDKRequest represents a request to delete an SDK
type DeleteSDKRequest struct {
	SDKID string `json:"sdk_id" validate:"required"`
}

// DeleteSDKResponse represents the response from SDK deletion
type DeleteSDKResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SDKGenerationStatus represents the status of an SDK generation task.
type SDKGenerationStatus string

const (
	SDKStatusPending    SDKGenerationStatus = "pending"
	SDKStatusInProgress SDKGenerationStatus = "inprogress" // Added InProgress
	SDKStatusCompleted  SDKGenerationStatus = "completed"
	SDKStatusFailed     SDKGenerationStatus = "failed"
	SDKStatusDeleted    SDKGenerationStatus = "deleted" // Soft delete status
)

// GenerationType defines the type of artifact generated.
type GenerationType string

const (
	GenerationTypeSDK GenerationType = "sdk"
	GenerationTypeMCP GenerationType = "mcp"
)

// SDK represents an SDK or MCP generation record.
type SDK struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         string              `bson:"userId" json:"userId"` // User who generated the SDK
	CollectionID   string              `bson:"collectionId,omitempty" json:"collectionId,omitempty"`
	GenerationType GenerationType      `bson:"generationType" json:"generationType"` // New: "sdk" or "mcp"

	// SDK-specific fields (optional if GenerationType is mcp)
	PackageName string `bson:"packageName,omitempty" json:"packageName,omitempty"`
	Language    string `bson:"language,omitempty" json:"language,omitempty"`

	// MCP-specific fields (optional if GenerationType is sdk)
	MCPTransport string `bson:"mcpTransport,omitempty" json:"mcpTransport,omitempty"`
	MCPPort      int    `bson:"mcpPort,omitempty" json:"mcpPort,omitempty"`

	Status         SDKGenerationStatus `bson:"status" json:"status"`
	FilePath       string              `bson:"filePath,omitempty" json:"filePath,omitempty"`                 // Path to the generated SDK archive/folder
	DownloadURL    string              `bson:"downloadUrl,omitempty" json:"downloadUrl,omitempty"`           // If served via a specific URL
	ErrorMessage   string              `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`         // If status is "failed"
	GeneratedAt    time.Time           `bson:"generatedAt,omitempty" json:"generatedAt,omitempty"`           // Timestamp of actual successful generation
	FinishedAt     time.Time           `bson:"finishedAt,omitempty" json:"finishedAt,omitempty"`             // Added FinishedAt
	GenerationTime int64               `bson:"generationTimeMs,omitempty" json:"generationTimeMs,omitempty"` // Time taken in milliseconds
	CreatedAt      time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time           `bson:"updatedAt" json:"updatedAt"`
	IsDeleted      bool                `bson:"isDeleted,omitempty" json:"isDeleted,omitempty"` // For soft deletes
}

// MCPGenerationRequest is now defined in internal/models/request_types.go

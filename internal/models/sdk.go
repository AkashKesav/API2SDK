package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SDKGenerationStatus represents the status of an SDK generation task.
type SDKGenerationStatus string

const (
	SDKStatusPending   SDKGenerationStatus = "pending"
	SDKStatusCompleted SDKGenerationStatus = "completed"
	SDKStatusFailed    SDKGenerationStatus = "failed"
	SDKStatusDeleted   SDKGenerationStatus = "deleted" // Soft delete status
)

// SDK represents an SDK generation record.
type SDK struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         string              `bson:"userId" json:"userId"` // User who generated the SDK
	CollectionID   string              `bson:"collectionId,omitempty" json:"collectionId,omitempty"`
	PackageName    string              `bson:"packageName" json:"packageName"`
	Language       string              `bson:"language" json:"language"`
	Status         SDKGenerationStatus `bson:"status" json:"status"`
	FilePath       string              `bson:"filePath,omitempty" json:"filePath,omitempty"`                 // Path to the generated SDK archive/folder
	DownloadURL    string              `bson:"downloadUrl,omitempty" json:"downloadUrl,omitempty"`           // If served via a specific URL
	ErrorMessage   string              `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`         // If status is "failed"
	GeneratedAt    time.Time           `bson:"generatedAt,omitempty" json:"generatedAt,omitempty"`           // Timestamp of actual successful generation
	GenerationTime int64               `bson:"generationTimeMs,omitempty" json:"generationTimeMs,omitempty"` // Time taken in milliseconds
	CreatedAt      time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time           `bson:"updatedAt" json:"updatedAt"`
	IsDeleted      bool                `bson:"isDeleted,omitempty" json:"isDeleted,omitempty"` // For soft deletes
}

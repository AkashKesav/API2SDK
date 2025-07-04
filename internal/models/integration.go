package models

import (
	"time"

	"github.com/AkashKesav/API2SDK/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Integration represents a third-party API that can be integrated into the unified MCP.
type Integration struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string                `bson:"name" json:"name"`
	Description string                `bson:"description" json:"description"`
	BaseURL     string                `bson:"baseURL" json:"baseURL"`
	APIKey      types.EncryptedString `bson:"apiKey,omitempty" json:"-"`
	OpenAPISpec string                `bson:"openapiSpec" json:"openapiSpec"`
	CreatedAt   time.Time             `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time             `bson:"updatedAt" json:"updatedAt"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlatformSettings holds the general platform settings as a single document.
type PlatformSettings struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	PostmanAPIKey   string                 `bson:"postmanApiKey,omitempty" json:"postmanApiKey,omitempty"`
	JWTSecretKey    string                 `bson:"jwtSecretKey,omitempty" json:"jwtSecretKey,omitempty"`
	MaintenanceMode bool                   `bson:"maintenanceMode,omitempty" json:"maintenanceMode,omitempty"` // Added Maintenance Mode
	LogConfig       LogSettings            `bson:"logConfig,omitempty" json:"logConfig,omitempty"`             // Added for log management settings
	Settings        map[string]interface{} `bson:"settings" json:"settings"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updatedAt"`
}

// LogSettings defines configuration for system logging.
type LogSettings struct {
	Enabled       bool   `bson:"enabled,omitempty" json:"enabled,omitempty"`             // Whether custom log retrieval is enabled
	SourceType    string `bson:"sourceType,omitempty" json:"sourceType,omitempty"`       // e.g., "file", "database", "external_service"
	SourceDetails string `bson:"sourceDetails,omitempty" json:"sourceDetails,omitempty"` // Path for file, connection string for DB, URL for service
	DefaultLevel  string `bson:"defaultLevel,omitempty" json:"defaultLevel,omitempty"`   // Default log level to query if not specified
	RetentionDays int    `bson:"retentionDays,omitempty" json:"retentionDays,omitempty"` // How long logs are kept (informational)
}

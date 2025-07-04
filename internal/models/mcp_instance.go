package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MCPInstance represents a user's configured instance of an Integration.
type MCPInstance struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        string             `bson:"userId" json:"userId"`
	IntegrationID string             `bson:"integrationId" json:"integrationId"`
	Port          int                `bson:"port" json:"port"`
	Transport     string             `bson:"transport" json:"transport"`
	ServerType    string             `bson:"serverType" json:"serverType"`
	ToolsURL      string             `bson:"toolsUrl,omitempty" json:"toolsUrl,omitempty"`
	Resources     []Resource         `bson:"resources,omitempty" json:"resources"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

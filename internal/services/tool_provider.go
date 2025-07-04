package services

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToolProvider defines the interface for any integration that wants to expose
// its functionality as a set of tools to the MCP.
type ToolProvider interface {
	// GetTools returns a list of all tools provided by the integration.
	GetTools(ctx context.Context, integrationID primitive.ObjectID) ([]models.Tool, error)

	// ExecuteTool runs a specific tool with the given arguments.
	ExecuteTool(ctx context.Context, integrationID primitive.ObjectID, toolName string, arguments map[string]interface{}) (interface{}, error)
}

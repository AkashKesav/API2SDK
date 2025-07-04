package mcpgen

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Generator holds the services required for MCP generation.
type Generator struct {
	MCPInstanceService services.MCPInstanceService
}

// GenerateMCPServer creates a new MCP instance in the database.
func (g *Generator) GenerateMCPServer(ctx context.Context, integrationID, userID string) (*models.MCPInstance, error) {
	instance := &models.MCPInstance{
		ID:            primitive.NewObjectID(),
		UserID:        userID,
		IntegrationID: integrationID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	createdInstance, err := g.MCPInstanceService.CreateMCPInstance(ctx, instance)
	if err != nil {
		return nil, err
	}

	return createdInstance, nil
}

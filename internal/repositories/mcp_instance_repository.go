package repositories

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MCPInstanceRepository defines the interface for interacting with MCPInstance data.
type MCPInstanceRepository interface {
	Create(ctx context.Context, mcpInstance *models.MCPInstance) (*models.MCPInstance, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.MCPInstance, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.MCPInstance, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// mcpInstanceRepository is the concrete implementation of MCPInstanceRepository.
type mcpInstanceRepository struct {
	collection *mongo.Collection
}

// NewMCPInstanceRepository creates a new MCPInstanceRepository.
func NewMCPInstanceRepository(db *mongo.Database) MCPInstanceRepository {
	return &mcpInstanceRepository{
		collection: db.Collection("mcp_instances"),
	}
}

// Create creates a new MCPInstance.
func (r *mcpInstanceRepository) Create(ctx context.Context, mcpInstance *models.MCPInstance) (*models.MCPInstance, error) {
	mcpInstance.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, mcpInstance)
	if err != nil {
		return nil, err
	}
	return mcpInstance, nil
}

// GetByID retrieves an MCPInstance by its ID.
func (r *mcpInstanceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.MCPInstance, error) {
	var mcpInstance models.MCPInstance
	err := r.collection.FindOne(ctx, primitive.M{"_id": id}).Decode(&mcpInstance)
	if err != nil {
		return nil, err
	}
	return &mcpInstance, nil
}

// GetByUserID retrieves all MCPInstances for a given user.
func (r *mcpInstanceRepository) GetByUserID(ctx context.Context, userID string) ([]*models.MCPInstance, error) {
	var mcpInstances []*models.MCPInstance
	cursor, err := r.collection.Find(ctx, primitive.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &mcpInstances); err != nil {
		return nil, err
	}

	return mcpInstances, nil
}

// Delete deletes an MCPInstance by its ID.
func (r *mcpInstanceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, primitive.M{"_id": id})
	return err
}

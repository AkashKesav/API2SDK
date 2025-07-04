package repositories

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IntegrationRepository defines the interface for interacting with Integration data.
type IntegrationRepository interface {
	Create(ctx context.Context, integration *models.Integration) (*models.Integration, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Integration, error)
	GetAll(ctx context.Context) ([]*models.Integration, error)
	Update(ctx context.Context, id primitive.ObjectID, integration *models.Integration) (*models.Integration, error)
}

// integrationRepository is the concrete implementation of IntegrationRepository.
type integrationRepository struct {
	collection *mongo.Collection
}

// NewIntegrationRepository creates a new IntegrationRepository.
func NewIntegrationRepository(db *mongo.Database) IntegrationRepository {
	return &integrationRepository{
		collection: db.Collection("integrations"),
	}
}

// Create creates a new Integration.
func (r *integrationRepository) Create(ctx context.Context, integration *models.Integration) (*models.Integration, error) {
	integration.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, integration)
	if err != nil {
		return nil, err
	}
	return integration, nil
}

// GetByID retrieves an Integration by its ID.
func (r *integrationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Integration, error) {
	var integration models.Integration
	err := r.collection.FindOne(ctx, primitive.M{"_id": id}).Decode(&integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

// GetAll retrieves all Integrations.
func (r *integrationRepository) GetAll(ctx context.Context) ([]*models.Integration, error) {
	var integrations []*models.Integration
	cursor, err := r.collection.Find(ctx, primitive.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &integrations); err != nil {
		return nil, err
	}

	return integrations, nil
}

// Update updates an existing Integration.
func (r *integrationRepository) Update(ctx context.Context, id primitive.ObjectID, integration *models.Integration) (*models.Integration, error) {
	_, err := r.collection.UpdateOne(ctx, primitive.M{"_id": id}, bson.M{"$set": integration})
	if err != nil {
		return nil, err
	}
	return integration, nil
}

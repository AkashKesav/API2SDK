package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PublicAPIRepository struct {
	collection *mongo.Collection
}

// NewPublicAPIRepository creates a new public API repository
func NewPublicAPIRepository(db *mongo.Database) *PublicAPIRepository {
	return &PublicAPIRepository{
		collection: db.Collection("public_apis"),
	}
}

// Create creates a new public API
func (r *PublicAPIRepository) Create(ctx context.Context, publicAPI *models.PublicAPI) (*models.PublicAPI, error) {
	_, err := r.collection.InsertOne(ctx, publicAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to create public API: %w", err)
	}

	return publicAPI, nil
}

// GetByID retrieves a public API by ID
func (r *PublicAPIRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.PublicAPI, error) {
	var publicAPI models.PublicAPI
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&publicAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API: %w", err)
	}

	return &publicAPI, nil
}

// GetAll retrieves all public APIs with optional filtering
func (r *PublicAPIRepository) GetAll(ctx context.Context, filter bson.M, page, limit int) ([]models.PublicAPI, int64, error) {
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count public APIs: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find public APIs: %w", err)
	}
	defer cursor.Close(ctx)

	var publicAPIs []models.PublicAPI
	if err = cursor.All(ctx, &publicAPIs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode public APIs: %w", err)
	}

	return publicAPIs, total, nil
}

// Update updates an existing public API
func (r *PublicAPIRepository) Update(ctx context.Context, publicAPI *models.PublicAPI) (*models.PublicAPI, error) {
	publicAPI.UpdatedAt = time.Now()

	filter := bson.M{"_id": publicAPI.ID}
	update := bson.M{"$set": publicAPI}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update public API: %w", err)
	}

	return publicAPI, nil
}

// Delete deletes a public API by ID
func (r *PublicAPIRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete public API: %w", err)
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Search searches for public APIs based on text query
func (r *PublicAPIRepository) Search(ctx context.Context, query string, category string, page, limit int) ([]models.PublicAPI, int64, error) {
	filter := bson.M{}

	// Add text search if query is provided
	if query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$elemMatch": bson.M{"$regex": query, "$options": "i"}}},
		}
	}

	// Add category filter if provided
	if category != "" {
		filter["category"] = bson.M{"$regex": category, "$options": "i"}
	}

	// Only active APIs
	filter["is_active"] = true

	return r.GetAll(ctx, filter, page, limit)
}

// GetByPostmanID retrieves a public API by Postman collection ID
func (r *PublicAPIRepository) GetByPostmanID(ctx context.Context, postmanID string) (*models.PublicAPI, error) {
	var publicAPI models.PublicAPI
	filter := bson.M{"postman_id": postmanID}
	err := r.collection.FindOne(ctx, filter).Decode(&publicAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API by Postman ID: %w", err)
	}

	return &publicAPI, nil
}

// GetCategories returns distinct categories
func (r *PublicAPIRepository) GetCategories(ctx context.Context) ([]string, error) {
	categories, err := r.collection.Distinct(ctx, "category", bson.M{"is_active": true})
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	result := make([]string, len(categories))
	for i, cat := range categories {
		if str, ok := cat.(string); ok {
			result[i] = str
		}
	}

	return result, nil
}

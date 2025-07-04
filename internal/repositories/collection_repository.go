package repositories

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionRepository struct {
	collection *mongo.Collection
}

// NewCollectionRepository creates a new collection repository
func NewCollectionRepository(db *mongo.Database) *CollectionRepository {
	return &CollectionRepository{
		collection: db.Collection("collections"),
	}
}

// Create inserts a new collection
func (r *CollectionRepository) Create(ctx context.Context, collection *models.Collection) (*models.Collection, error) {
	collection.ID = primitive.NewObjectID()
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, collection)
	if err != nil {
		return nil, err
	}

	collection.ID = result.InsertedID.(primitive.ObjectID)
	return collection, nil
}

// GetAll retrieves all collections
func (r *CollectionRepository) GetAll(ctx context.Context) ([]*models.Collection, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*models.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		return nil, err
	}

	return collections, nil
}

// GetByID retrieves a collection by ID
func (r *CollectionRepository) GetByID(ctx context.Context, id string) (*models.Collection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var collection models.Collection
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&collection)
	if err != nil {
		return nil, err
	}

	return &collection, nil
}

// Update updates a collection
func (r *CollectionRepository) Update(ctx context.Context, id string, updateData *models.UpdateCollectionRequest) (*models.Collection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if updateData.Name != "" {
		update["$set"].(bson.M)["name"] = updateData.Name
	}
	if updateData.Description != "" {
		update["$set"].(bson.M)["description"] = updateData.Description
	}
	if updateData.PostmanData != nil {
		update["$set"].(bson.M)["postman_data"] = updateData.PostmanData
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedCollection models.Collection
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		update,
		opts,
	).Decode(&updatedCollection)

	if err != nil {
		return nil, err
	}

	return &updatedCollection, nil
}

// Delete removes a collection
func (r *CollectionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// GetByUserID retrieves collections by user ID
func (r *CollectionRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Collection, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*models.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		return nil, err
	}

	return collections, nil
}

// CountCreatedAfter returns the number of collections created after the given time
func (r *CollectionRepository) CountCreatedAfter(ctx context.Context, after time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gt": after}}
	return r.collection.CountDocuments(ctx, filter)
}

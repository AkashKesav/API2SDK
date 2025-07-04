package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const sdkCollectionName = "sdks"

// SDKRepository handles database operations for SDKs.
type SDKRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewSDKRepository creates a new SDKRepository.
func NewSDKRepository(db *mongo.Database, logger *zap.Logger) *SDKRepository {
	return &SDKRepository{
		collection: db.Collection(sdkCollectionName),
		logger:     logger,
	}
}

// Create inserts a new SDK record into the database.
func (r *SDKRepository) Create(ctx context.Context, sdk *models.SDK) (*models.SDK, error) {
	sdk.ID = primitive.NewObjectID()
	sdk.CreatedAt = time.Now()
	sdk.UpdatedAt = time.Now()
	sdk.IsDeleted = false

	_, err := r.collection.InsertOne(ctx, sdk)
	if err != nil {
		r.logger.Error("Failed to create SDK record", zap.Error(err), zap.Any("sdk", sdk))
		return nil, err
	}
	r.logger.Info("SDK record created successfully", zap.String("sdkID", sdk.ID.Hex()))
	return sdk, nil
}

// GetByID retrieves an SDK record by its ID.
func (r *SDKRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.SDK, error) {
	var sdk models.SDK
	filter := bson.M{"_id": id, "isDeleted": bson.M{"$ne": true}} // Exclude soft-deleted
	err := r.collection.FindOne(ctx, filter).Decode(&sdk)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Warn("SDK record not found", zap.String("sdkID", id.Hex()))
			return nil, nil // Or return a specific "not found" error
		}
		r.logger.Error("Failed to get SDK record by ID", zap.Error(err), zap.String("sdkID", id.Hex()))
		return nil, err
	}
	return &sdk, nil
}

// GetByUserID retrieves a paginated list of SDK records for a specific user.
// Only returns non-soft-deleted SDKs.
func (r *SDKRepository) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error) {
	var sdks []*models.SDK
	filter := bson.M{"userId": userID, "isDeleted": bson.M{"$ne": true}}

	skip := int64((page - 1) * limit)
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by newest first

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("Failed to find SDK records by userID", zap.Error(err), zap.String("userID", userID))
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &sdks); err != nil {
		r.logger.Error("Failed to decode SDK records for userID", zap.Error(err), zap.String("userID", userID))
		return nil, 0, err
	}

	totalCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to count SDK records for userID", zap.Error(err), zap.String("userID", userID))
		return nil, 0, err
	}

	return sdks, totalCount, nil
}

// GetByCollectionID retrieves all SDK records for a specific collection.
// Only returns non-soft-deleted SDKs.
func (r *SDKRepository) GetByCollectionID(ctx context.Context, collectionID string) ([]*models.SDK, error) {
	var sdks []*models.SDK
	filter := bson.M{"collectionId": collectionID, "isDeleted": bson.M{"$ne": true}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to find SDK records by collectionID", zap.Error(err), zap.String("collectionID", collectionID))
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &sdks); err != nil {
		r.logger.Error("Failed to decode SDK records for collectionID", zap.Error(err), zap.String("collectionID", collectionID))
		return nil, err
	}

	return sdks, nil
}

// Update modifies an existing SDK record.
func (r *SDKRepository) Update(ctx context.Context, sdk *models.SDK) error {
	sdk.UpdatedAt = time.Now()
	filter := bson.M{"_id": sdk.ID, "userId": sdk.UserID} // Ensure user owns the SDK
	update := bson.M{"$set": sdk}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update SDK record", zap.Error(err), zap.String("sdkID", sdk.ID.Hex()))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("No SDK record found to update or user mismatch", zap.String("sdkID", sdk.ID.Hex()), zap.String("userID", sdk.UserID))
		return errors.New("sdk not found or permission denied")
	}
	r.logger.Info("SDK record updated successfully", zap.String("sdkID", sdk.ID.Hex()))
	return nil
}

// UpdateFields updates specific fields of an SDK record.
func (r *SDKRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": fields}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update SDK fields",
			zap.String("sdkID", id.Hex()),
			zap.Any("fields", fields),
			zap.Error(err))
		return err
	}
	r.logger.Info("SDK fields updated successfully", zap.String("sdkID", id.Hex()), zap.Any("fields", fields))
	return nil
}

// SoftDelete marks an SDK record as deleted.
// Ensures the SDK belongs to the user requesting deletion.
func (r *SDKRepository) SoftDelete(ctx context.Context, id primitive.ObjectID, userID string) error {
	filter := bson.M{"_id": id, "userId": userID}
	update := bson.M{"$set": bson.M{"isDeleted": true, "status": models.SDKStatusDeleted, "updatedAt": time.Now()}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to soft-delete SDK record", zap.Error(err), zap.String("sdkID", id.Hex()), zap.String("userID", userID))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("No SDK record found to soft-delete or user mismatch", zap.String("sdkID", id.Hex()), zap.String("userID", userID))
		return errors.New("sdk not found or permission denied for deletion")
	}
	r.logger.Info("SDK record soft-deleted successfully", zap.String("sdkID", id.Hex()))
	return nil
}

// SoftDeleteSDK marks an SDK record as deleted and sets the deletion timestamp.
// This is an example of how a specific soft delete method might look if you prefer it over generic UpdateFields.
func (r *SDKRepository) SoftDeleteSDK(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
			"status":    models.SDKStatusDeleted,
			"deletedAt": time.Now(), // Assuming you add DeletedAt to your model
			"updatedAt": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to soft delete SDK record", zap.Error(err), zap.String("sdkID", id.Hex()))
		return err
	}
	r.logger.Info("SDK record soft-deleted successfully", zap.String("sdkID", id.Hex()))
	return nil
}

// HardDelete permanently removes an SDK record from the database.
// Ensures the SDK belongs to the user requesting deletion.
// Use with caution. Consider if physical file deletion is also needed.
func (r *SDKRepository) HardDelete(ctx context.Context, id primitive.ObjectID, userID string) error {
	filter := bson.M{"_id": id, "userId": userID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to hard-delete SDK record", zap.Error(err), zap.String("sdkID", id.Hex()), zap.String("userID", userID))
		return err
	}
	if result.DeletedCount == 0 {
		r.logger.Warn("No SDK record found to hard-delete or user mismatch", zap.String("sdkID", id.Hex()), zap.String("userID", userID))
		return errors.New("sdk not found or permission denied for deletion")
	}
	r.logger.Info("SDK record hard-deleted successfully", zap.String("sdkID", id.Hex()))
	return nil
}

// Collection returns the underlying MongoDB collection for advanced queries/statistics
func (r *SDKRepository) Collection() *mongo.Collection {
	return r.collection
}

package repositories

// Adding a comment to trigger re-evaluation

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models" // Corrected import path
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PlatformSettingsRepository defines the interface for interacting with platform settings storage.
type PlatformSettingsRepository interface {
	GetSettings(ctx context.Context) (*models.PlatformSettings, error)
	UpdateSettings(ctx context.Context, settings map[string]interface{}) (*models.PlatformSettings, error)
}

// MongoPlatformSettingsRepository implements PlatformSettingsRepository for MongoDB.
type MongoPlatformSettingsRepository struct {
	collection *mongo.Collection
}

// NewMongoPlatformSettingsRepository creates a new MongoPlatformSettingsRepository.
func NewMongoPlatformSettingsRepository(db *mongo.Database) PlatformSettingsRepository {
	return &MongoPlatformSettingsRepository{
		collection: db.Collection("platform_settings"),
	}
}

// GetSettings retrieves the current platform settings.
// There should ideally be only one document for platform settings.
func (r *MongoPlatformSettingsRepository) GetSettings(ctx context.Context) (*models.PlatformSettings, error) {
	var settings models.PlatformSettings
	// Find the first document. If you have a specific ID or filter, use that.
	// For simplicity, assuming there's at most one settings document or we fetch the first one.
	// If you want to ensure only one, you might use a known ID.
	filter := bson.M{} // Empty filter to get the first/only document
	err := r.collection.FindOne(ctx, filter).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// No settings document found, return a default or empty struct
		return &models.PlatformSettings{Settings: make(map[string]interface{})}, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateSettings updates the platform settings.
// This uses an upsert operation to create the settings document if it doesn't exist,
// or update it if it does. We assume a single settings document.
func (r *MongoPlatformSettingsRepository) UpdateSettings(ctx context.Context, newSettings map[string]interface{}) (*models.PlatformSettings, error) {
	filter := bson.M{} // Empty filter to target the single settings document or create one if none exists.
	// For a more robust single-document approach, you could use a fixed known ID here.
	update := bson.M{
		"$set": bson.M{
			"settings":  newSettings,
			"updatedAt": time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedSettings models.PlatformSettings
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedSettings)
	if err != nil {
		return nil, err
	}
	return &updatedSettings, nil
}

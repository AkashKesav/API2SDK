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

// Repository defines the standard repository interface
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	GetByID(ctx context.Context, id models.ID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id models.ID) error
	List(ctx context.Context, filter interface{}, page, limit int) ([]*T, int64, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
	Exists(ctx context.Context, filter interface{}) (bool, error)
}

// SoftDeleteRepository extends Repository with soft delete functionality
type SoftDeleteRepository[T any] interface {
	Repository[T]
	SoftDelete(ctx context.Context, id models.ID) error
	Restore(ctx context.Context, id models.ID) error
	ListDeleted(ctx context.Context, page, limit int) ([]*T, int64, error)
	PermanentDelete(ctx context.Context, id models.ID) error
}

// BaseRepository provides common repository functionality
type BaseRepository[T any] struct {
	collection *mongo.Collection
	entityName string
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *mongo.Database, collectionName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		collection: db.Collection(collectionName),
		entityName: collectionName,
	}
}

// Create inserts a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	// Set timestamps if entity implements BaseModel
	if baseModel, ok := any(entity).(interface{ SetTimestamps() }); ok {
		baseModel.SetTimestamps()
	}

	// Set ID if not already set
	if idSetter, ok := any(entity).(interface{ SetID(models.ID) }); ok {
		if getter, ok := any(entity).(interface{ GetID() models.ID }); ok {
			if getter.GetID().IsZero() {
				idSetter.SetID(models.NewID())
			}
		}
	}

	_, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// GetByID retrieves an entity by ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id models.ID) (*T, error) {
	var entity T

	filter := bson.M{"_id": id.ObjectID()}

	// Add soft delete filter if entity supports it
	if _, ok := any(&entity).(models.SoftDeletable); ok {
		filter["isDeleted"] = bson.M{"$ne": true}
	}

	err := r.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entity, nil
}

// Update modifies an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	// Set updated timestamp
	if baseModel, ok := any(entity).(interface{ SetTimestamps() }); ok {
		baseModel.SetTimestamps()
	}

	// Increment version for optimistic locking
	if versionable, ok := any(entity).(interface{ IncrementVersion() }); ok {
		versionable.IncrementVersion()
	}

	var id primitive.ObjectID
	if getter, ok := any(entity).(interface{ GetID() models.ID }); ok {
		id = getter.GetID().ObjectID()
	} else {
		return ErrInvalidEntity
	}

	filter := bson.M{"_id": id}

	// Add version check for optimistic locking
	if versionGetter, ok := any(entity).(interface{ GetVersion() int64 }); ok {
		filter["version"] = versionGetter.GetVersion() - 1 // Previous version
	}

	update := bson.M{"$set": entity}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// Delete removes an entity
func (r *BaseRepository[T]) Delete(ctx context.Context, id models.ID) error {
	filter := bson.M{"_id": id.ObjectID()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// List retrieves entities with pagination
func (r *BaseRepository[T]) List(ctx context.Context, filter interface{}, page, limit int) ([]*T, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	// Add soft delete filter if applicable
	if filter == nil {
		filter = bson.M{}
	}

	if filterMap, ok := filter.(bson.M); ok {
		var entity T
		if _, ok := any(&entity).(models.SoftDeletable); ok {
			filterMap["isDeleted"] = bson.M{"$ne": true}
		}
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get entities
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var entities []*T
	if err = cursor.All(ctx, &entities); err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Count returns the number of entities matching the filter
func (r *BaseRepository[T]) Count(ctx context.Context, filter interface{}) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}

	// Add soft delete filter if applicable
	if filterMap, ok := filter.(bson.M); ok {
		var entity T
		if _, ok := any(&entity).(models.SoftDeletable); ok {
			filterMap["isDeleted"] = bson.M{"$ne": true}
		}
	}

	return r.collection.CountDocuments(ctx, filter)
}

// Exists checks if an entity exists
func (r *BaseRepository[T]) Exists(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SoftDelete marks an entity as deleted
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id models.ID) error {
	filter := bson.M{"_id": id.ObjectID()}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// Restore restores a soft-deleted entity
func (r *BaseRepository[T]) Restore(ctx context.Context, id models.ID) error {
	filter := bson.M{
		"_id":       id.ObjectID(),
		"isDeleted": true,
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": false,
			"updatedAt": time.Now(),
		},
		"$unset": bson.M{
			"deletedAt": "",
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// ListDeleted retrieves soft-deleted entities
func (r *BaseRepository[T]) ListDeleted(ctx context.Context, page, limit int) ([]*T, int64, error) {
	filter := bson.M{"isDeleted": true}
	return r.List(ctx, filter, page, limit)
}

// PermanentDelete permanently removes an entity
func (r *BaseRepository[T]) PermanentDelete(ctx context.Context, id models.ID) error {
	filter := bson.M{"_id": id.ObjectID()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// Common repository errors
var (
	ErrEntityNotFound = NewRepositoryError("ENTITY_NOT_FOUND", "Entity not found")
	ErrInvalidEntity  = NewRepositoryError("INVALID_ENTITY", "Invalid entity")
	ErrDuplicateKey   = NewRepositoryError("DUPLICATE_KEY", "Duplicate key error")
)

// RepositoryError represents a repository-specific error
type RepositoryError struct {
	Code    string
	Message string
}

func (e RepositoryError) Error() string {
	return e.Message
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(code, message string) RepositoryError {
	return RepositoryError{
		Code:    code,
		Message: message,
	}
}

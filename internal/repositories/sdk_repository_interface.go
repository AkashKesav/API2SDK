package repositories

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SDKRepositoryInterface defines the operations for SDK persistence.
type SDKRepositoryInterface interface {
	Create(ctx context.Context, sdk *models.SDK) (*models.SDK, error)
	Update(ctx context.Context, sdk *models.SDK) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.SDK, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error)
	GetByCollectionID(ctx context.Context, collectionID string) ([]*models.SDK, error)
	UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error
	SoftDelete(ctx context.Context, id primitive.ObjectID, userID string) error
	HardDelete(ctx context.Context, id primitive.ObjectID, userID string) error
}

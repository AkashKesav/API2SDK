package repositories

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type APILogRepository struct {
	collection *mongo.Collection
}

func NewAPILogRepository(db *mongo.Database) *APILogRepository {
	return &APILogRepository{collection: db.Collection("api_logs")}
}

func (r *APILogRepository) CountAPICallsCreatedAfter(ctx context.Context, after time.Time) (int64, error) {
	filter := bson.M{"createdAt": bson.M{"$gte": after}}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *APILogRepository) CreateAPILog(ctx context.Context, log *models.APILog) error {
	log.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

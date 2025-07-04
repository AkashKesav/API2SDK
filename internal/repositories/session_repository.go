package repositories

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{collection: db.Collection("sessions")}
}

func (r *SessionRepository) CountSessionsCreatedAfter(ctx context.Context, after time.Time) (int64, error) {
	filter := bson.M{"createdAt": bson.M{"$gte": after}}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	session.ID = primitive.NewObjectID()
	session.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *SessionRepository) DeleteSession(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

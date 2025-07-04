package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type APILog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	Endpoint  string             `bson:"endpoint" json:"endpoint"`
	Method    string             `bson:"method" json:"method"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

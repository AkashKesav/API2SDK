package services

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionService struct {
	Repo *repositories.SessionRepository
}

func NewSessionService(repo *repositories.SessionRepository) *SessionService {
	return &SessionService{Repo: repo}
}

func (s *SessionService) GetSessionCountSince(ctx context.Context, after time.Time) (int64, error) {
	return s.Repo.CountSessionsCreatedAfter(ctx, after)
}

func (s *SessionService) CreateSession(ctx context.Context, userID primitive.ObjectID) error {
	session := &models.Session{UserID: userID}
	return s.Repo.CreateSession(ctx, session)
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID primitive.ObjectID) error {
	return s.Repo.DeleteSession(ctx, sessionID)
}

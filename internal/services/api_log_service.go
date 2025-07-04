package services

import (
	"context"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type APILogService struct {
	Repo *repositories.APILogRepository
}

func NewAPILogService(repo *repositories.APILogRepository) *APILogService {
	return &APILogService{Repo: repo}
}

func (s *APILogService) GetAPICallCountSince(ctx context.Context, after time.Time) (int64, error) {
	return s.Repo.CountAPICallsCreatedAfter(ctx, after)
}

func (s *APILogService) LogAPICall(ctx context.Context, userID primitive.ObjectID, endpoint, method string) error {
	log := &models.APILog{UserID: userID, Endpoint: endpoint, Method: method}
	return s.Repo.CreateAPILog(ctx, log)
}

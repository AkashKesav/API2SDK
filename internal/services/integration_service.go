package services

import (
	"context"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IntegrationService defines the interface for managing integrations.
type IntegrationService interface {
	CreateIntegration(ctx context.Context, integration *models.Integration) (*models.Integration, error)
	GetIntegration(ctx context.Context, id primitive.ObjectID) (*models.Integration, error)
	ListIntegrations(ctx context.Context) ([]*models.Integration, error)
	UpdateIntegration(ctx context.Context, id primitive.ObjectID, integration *models.Integration) (*models.Integration, error)
}

// integrationService is the concrete implementation of IntegrationService.
type integrationService struct {
	repo repositories.IntegrationRepository
}

// NewIntegrationService creates a new IntegrationService.
func NewIntegrationService(repo repositories.IntegrationRepository) IntegrationService {
	return &integrationService{repo: repo}
}

// CreateIntegration creates a new integration.
func (s *integrationService) CreateIntegration(ctx context.Context, integration *models.Integration) (*models.Integration, error) {
	// The APIKey is already an EncryptedString, so no need to encrypt it here.
	// The BSON marshaller will handle it automatically.
	return s.repo.Create(ctx, integration)
}

// GetIntegration retrieves an integration by its ID.
func (s *integrationService) GetIntegration(ctx context.Context, id primitive.ObjectID) (*models.Integration, error) {
	return s.repo.GetByID(ctx, id)
}

// ListIntegrations retrieves all integrations.
func (s *integrationService) ListIntegrations(ctx context.Context) ([]*models.Integration, error) {
	return s.repo.GetAll(ctx)
}

// UpdateIntegration updates an existing integration.
func (s *integrationService) UpdateIntegration(ctx context.Context, id primitive.ObjectID, integration *models.Integration) (*models.Integration, error) {
	// The APIKey is already an EncryptedString, so no need to re-encrypt it here.
	// The BSON marshaller will handle it automatically.
	return s.repo.Update(ctx, id, integration)
}

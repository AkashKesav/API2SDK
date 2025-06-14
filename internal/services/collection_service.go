package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CollectionService struct {
	repo       *repositories.CollectionRepository
	logger     *zap.Logger
	sdkRepo    *repositories.SDKRepository
	sdkService *SDKService // Added SDKService dependency
}

// NewCollectionService creates a new collection service
func NewCollectionService(repo *repositories.CollectionRepository, logger *zap.Logger, sdkRepo *repositories.SDKRepository, sdkService *SDKService) *CollectionService {
	return &CollectionService{
		repo:       repo,
		logger:     logger,
		sdkRepo:    sdkRepo,
		sdkService: sdkService, // Store sdkService
	}
}

// CreateCollection creates a new collection
func (s *CollectionService) CreateCollection(req *models.CreateCollectionRequest, userID string) (*models.Collection, error) {
	collection := &models.Collection{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		PostmanData: req.PostmanData,
		Endpoints:   []models.Endpoint{},
	}

	return s.repo.Create(collection)
}

// GetAllCollections retrieves all collections
func (s *CollectionService) GetAllCollections() ([]*models.Collection, error) {
	return s.repo.GetAll()
}

// GetCollection retrieves a collection by ID
func (s *CollectionService) GetCollection(id string) (*models.Collection, error) {
	return s.repo.GetByID(id)
}

// UpdateCollection updates a collection
func (s *CollectionService) UpdateCollection(id string, req *models.UpdateCollectionRequest) (*models.Collection, error) {
	return s.repo.Update(id, req)
}

// DeleteCollection deletes a collection
func (s *CollectionService) DeleteCollection(id string) error {
	return s.repo.Delete(id)
}

// GetCollectionsByUserID retrieves collections by user ID
func (s *CollectionService) GetCollectionsByUserID(userID string) ([]*models.Collection, error) {
	return s.repo.GetByUserID(userID)
}

// GenerateOpenAPISpec generates an OpenAPI specification from a Postman collection
// using the SDKService's ConvertPostmanToOpenAPI method.
// It saves the spec to a temporary file and returns the path and the spec string.
func (s *CollectionService) GenerateOpenAPISpec(collectionID string) (string, string, error) {
	collection, err := s.repo.GetByID(collectionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get collection: %w", err)
	}

	postmanDataStr, ok := collection.PostmanData.(string)
	if !ok {
		return "", "", fmt.Errorf("PostmanData for collection %s is not a string", collectionID)
	}

	if postmanDataStr == "" {
		return "", "", fmt.Errorf("collection %s has no Postman data", collectionID)
	}

	s.logger.Info("Calling SDKService.ConvertPostmanToOpenAPI", zap.String("collectionID", collectionID))

	// Use an empty options string for now. This can be made configurable later.
	optionsJSON := "{}"

	openAPISpecString, err := s.sdkService.ConvertPostmanToOpenAPI(context.Background(), postmanDataStr, optionsJSON)
	if err != nil {
		s.logger.Error("Failed to convert Postman to OpenAPI via SDKService", zap.String("collectionID", collectionID), zap.Error(err))
		return "", "", fmt.Errorf("conversion from Postman to OpenAPI failed: %w", err)
	}

	s.logger.Info("Successfully converted Postman to OpenAPI via SDKService", zap.String("collectionID", collectionID))

	// Define the output path for the OpenAPI spec
	tempSpecDir := filepath.Join(os.TempDir(), "api2sdk_openapi_specs")
	if err := utils.EnsureDir(tempSpecDir); err != nil {
		return "", "", fmt.Errorf("failed to create temporary spec directory %s: %w", tempSpecDir, err)
	}
	finalOpenAPISpecFileName := fmt.Sprintf("openapi_spec_%s_%s.json", collectionID, uuid.New().String())
	finalOpenAPISpecFilePath := filepath.Join(tempSpecDir, finalOpenAPISpecFileName)

	// Write the generated OpenAPI spec string to the file
	if err := os.WriteFile(finalOpenAPISpecFilePath, []byte(openAPISpecString), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write generated OpenAPI spec to file %s: %w", finalOpenAPISpecFilePath, err)
	}

	return finalOpenAPISpecFilePath, openAPISpecString, nil
}

// GenerateSDKFromCollection generates an SDK for a given language from a Postman collection.
// It first converts the Postman collection to OpenAPI, then generates the SDK.
// Returns the path to the generated SDK and an error if any.
func (s *CollectionService) GenerateSDKFromCollection(ctx context.Context, userID string, collectionID string, language string, outputDir string) (string, error) {
	// Step 1: Generate OpenAPI spec from Postman collection
	openAPISpecPath, _, err := s.GenerateOpenAPISpec(collectionID)
	if err != nil {
		return "", fmt.Errorf("failed to generate OpenAPI spec for collection %s: %w", collectionID, err)
	}
	// Defer removal of the temporary OpenAPI spec file
	defer func() {
		if err := os.Remove(openAPISpecPath); err != nil {
			s.logger.Warn("Failed to remove temporary OpenAPI spec file", zap.String("path", openAPISpecPath), zap.Error(err))
		}
	}()

	// Step 2: Generate SDK from the OpenAPI spec
	// Use the injected s.sdkService

	// Construct a more specific output directory for the SDK, e.g., outputDir/collection_uuid/language_pkgname
	// This outputDir should be unique per generation request to avoid conflicts if GenerateSDK is called concurrently.
	// The `outputDir` parameter to this function is the base, e.g., /tmp/sdks
	// The `outputDir` parameter to `s.sdkService.GenerateSDK` should be the specific path for *this* SDK.
	uniqueTaskID := uuid.New().String() // Generate a unique ID for this task
	sdkInstanceOutputDir := filepath.Join(outputDir, uniqueTaskID)

	if err := utils.EnsureDir(sdkInstanceOutputDir); err != nil {
		return "", fmt.Errorf("failed to create SDK instance output directory %s: %w", sdkInstanceOutputDir, err)
	}

	// Derive a targetPackageName.
	var targetPackageName string
	collection, err := s.repo.GetByID(collectionID)
	if err != nil {
		s.logger.Warn("Failed to get collection by ID for package name derivation, using default.", zap.String("collectionID", collectionID), zap.Error(err))
		targetPackageName = utils.DerivePackageName(nil, language) // Pass nil for collection to use default
	} else {
		targetPackageName = utils.DerivePackageName(collection, language)
	}

	// Call the SDKService's GenerateSDK method
	// It now expects: ctx, userID, collectionID, openAPISpecPath, language, sdkInstanceOutputDir, targetPackageName
	generatedSDKPath, err := s.sdkService.GenerateSDK(ctx, userID, collectionID, openAPISpecPath, language, sdkInstanceOutputDir, targetPackageName)
	if err != nil {
		return "", fmt.Errorf("failed to generate SDK for collection %s: %w", collectionID, err)
	}

	return generatedSDKPath, nil
}

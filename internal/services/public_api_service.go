package services

import (
	"context" // Added import
	"fmt"
	"strings"

	"github.com/AkashKesav/API2SDK/internal/models" // Corrected path
	"github.com/AkashKesav/API2SDK/internal/repositories" // Corrected path
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type PublicAPIService struct {
	repository        *repositories.PublicAPIRepository // For fetching/managing list of public APIs
	logger            *zap.Logger
	postmanAPIService *PostmanAPIService // For fetching actual Postman collection data
	collectionService *CollectionService // For saving imported collections and generating OpenAPI specs
}

// NewPublicAPIService creates a new public API service.
// It requires a logger, PostmanAPIService, and CollectionService.
// The PublicAPIRepository is internally instantiated for now, but could also be injected.
func NewPublicAPIService(logger *zap.Logger, postmanAPIService *PostmanAPIService, collectionService *CollectionService) *PublicAPIService {
	return &PublicAPIService{
		repository:        repositories.NewPublicAPIRepository(), // Assuming this doesn't need DB for static list for now
		logger:            logger,
		postmanAPIService: postmanAPIService,
		collectionService: collectionService,
	}
}

// PostmanService returns the underlying PostmanAPIService
func (s *PublicAPIService) PostmanService() *PostmanAPIService {
	return s.postmanAPIService
}

// CollectionService returns the underlying CollectionService
func (s *PublicAPIService) CollectionService() *CollectionService {
	return s.collectionService
}

// GetAllPublicAPIs retrieves all public APIs from the repository.
func (s *PublicAPIService) GetAllPublicAPIs(query, category string, page, limit int) ([]models.PublicAPI, int64, error) {
	// This currently uses a static list. If it should use the repository, this needs to change.
	// For now, let's assume the static list is the source for "GetAll" and "GetPopular"
	// and the repository would be for admin-managed entries.
	// If repository is the source: return s.repository.GetAll(query, category, page, limit)
	s.logger.Info("GetAllPublicAPIs called, currently returns static list, not repository data.")

	// Simulate basic filtering for the static list for now
	allApis := s.GetPopularAPIs() // Using popular as the base for "all" for now
	var filteredApis []models.PublicAPI
	for _, api := range allApis {
		matchQuery := true
		if query != "" && !strings.Contains(strings.ToLower(api.Name), strings.ToLower(query)) && !strings.Contains(strings.ToLower(api.Description), strings.ToLower(query)) {
			matchQuery = false
		}
		matchCategory := true
		if category != "" && !strings.EqualFold(api.Category, category) {
			matchCategory = false
		}
		if matchQuery && matchCategory {
			filteredApis = append(filteredApis, api)
		}
	}

	total := int64(len(filteredApis))
	start := (page - 1) * limit
	end := start + limit
	if start > len(filteredApis) {
		return []models.PublicAPI{}, total, nil
	}
	if end > len(filteredApis) {
		end = len(filteredApis)
	}

	return filteredApis[start:end], total, nil
}

// GetPublicAPIByPostmanID retrieves a specific public API entry by its Postman Collection UID.
// It uses models.PublicAPI.
func (s *PublicAPIService) GetPublicAPIByPostmanID(postmanID string) (*models.PublicAPI, error) { // Renamed to avoid conflict, and matches existing usage
	apis := s.GetPopularAPIs() // Using popular as the base for now
	for i := range apis {
		if apis[i].PostmanID == postmanID { // Assuming ID here refers to PostmanID for the static list
			return &apis[i], nil
		}
	}
	// If using a repository for other IDs:
	// return s.repository.GetByID(id)
	return nil, fmt.Errorf("public API with PostmanID '%s' not found in the predefined list", postmanID)
}

// GetPublicAPIByID retrieves a specific public API entry by its actual DB ID or PostmanID.
func (s *PublicAPIService) GetPublicAPIByID(id string) (*models.PublicAPI, error) {
	s.logger.Info("Attempting to retrieve public API by ID", zap.String("id", id))

	// First, try to find the API by PostmanID from the predefined list
	if api, err := s.GetPublicAPIByPostmanID(id); err == nil {
		s.logger.Info("Found public API in predefined list by PostmanID", zap.String("id", id))
		return api, nil
	}

	// If not found in the predefined list, try to get it from the repository by its database ID
	s.logger.Info("Public API not found in predefined list, attempting to fetch from repository", zap.String("id", id))
	objectID, err := primitive.ObjectIDFromHex(id) // Convert string to ObjectID
	if err != nil {
		s.logger.Error("Failed to convert ID string to ObjectID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("invalid ID format: '%s'", id)
	}
	api, err := s.repository.GetByID(objectID) // Use ObjectID
	if err != nil {
		s.logger.Error("Failed to get public API from repository", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("public API with ID '%s' not found", id)
	}

	s.logger.Info("Successfully retrieved public API from repository", zap.String("id", id))
	return api, nil
}

// CreatePublicAPI creates a new public API entry in the repository.
func (s *PublicAPIService) CreatePublicAPI(req *models.CreatePublicAPIRequest) (*models.PublicAPI, error) {
	// return s.repository.Create(req)
	s.logger.Warn("CreatePublicAPI called, but repository interaction is not fully implemented in this stub.")
	return nil, fmt.Errorf("CreatePublicAPI not implemented via repository yet")
}

// UpdatePublicAPI updates an existing public API entry in the repository.
func (s *PublicAPIService) UpdatePublicAPI(id string, req *models.UpdatePublicAPIRequest) (*models.PublicAPI, error) {
	// return s.repository.Update(id, req)
	s.logger.Warn("UpdatePublicAPI called, but repository interaction is not fully implemented in this stub.", zap.String("id", id))
	return nil, fmt.Errorf("UpdatePublicAPI not implemented via repository yet")
}

// DeletePublicAPI deletes a public API entry from the repository.
func (s *PublicAPIService) DeletePublicAPI(id string) error {
	// return s.repository.Delete(id)
	s.logger.Warn("DeletePublicAPI called, but repository interaction is not fully implemented in this stub.", zap.String("id", id))
	return fmt.Errorf("DeletePublicAPI not implemented via repository yet")
}

// SearchPublicCollections searches Postman's public API for collections.
// This should likely use the postmanAPIService.
func (s *PublicAPIService) SearchPublicCollections(query string, limit int) (interface{}, error) {
	s.logger.Info("SearchPublicCollections called in PublicAPIService. This might delegate to PostmanAPIService.", zap.String("query", query), zap.Int("limit", limit))
	// Example delegation (actual method in PostmanAPIService might differ):
	// return s.postmanAPIService.SearchCollections(context.Background(), query, limit)
	return nil, fmt.Errorf("SearchPublicCollections via PostmanAPIService not fully implemented yet")
}

// GetPopularAPIs retrieves a list of popular public APIs.
// It uses models.PublicAPI which is the correct struct name.
func (s *PublicAPIService) GetPopularAPIs() []models.PublicAPI {
	// This list is static for now. In a real app, it might come from s.repository.
	return []models.PublicAPI{
		{PostmanID: "123e4567-e89b-12d3-a456-426614174000", Name: "Weather API", Category: "Weather", Description: "Get current weather and forecasts.", BaseURL: "https://api.weather.com/v1", PostmanURL: "https://www.postman.com/collections/your-weather-collection-id", Tags: []string{"weather", "forecast"}, IsActive: true},
		{PostmanID: "123e4567-e89b-12d3-a456-426614174001", Name: "Joke API", Category: "Fun", Description: "Get random jokes.", BaseURL: "https://api.jokes.com/", PostmanURL: "https://www.postman.com/collections/your-joke-collection-id", Tags: []string{"jokes", "fun", "entertainment"}, IsActive: true},
		{PostmanID: "27997570-454c3c24-59f3-4b65-ba31-a52196923985", Name: "SpaceX API (Public)", Category: "Space", Description: "Publicly available SpaceX API data (capsules, rockets, etc.). Uses Postman Echo for collection URL.", BaseURL: "https://api.spacexdata.com/v4", PostmanURL: "https://www.postman.com/collections/c29db9c4a73a1a4267cf", Tags: []string{"space", "rockets", "elon musk", "public"}, IsActive: true},
		{PostmanID: "123e4567-e89b-12d3-a456-426614174003", Name: "Public Transport API", Category: "Transportation", Description: "Real-time public transport information.", BaseURL: "https://api.transport.org/", PostmanURL: "https://www.postman.com/collections/your-transport-collection-id", Tags: []string{"transport", "transit", "maps"}, IsActive: true},
		{PostmanID: "3117935-7e5c794c-4374-4a3a-a3a1-73d0f77699e8", Name: "Fake Store API (Public Collection)", Category: "Testing", Description: "A fake API for testing and prototyping e-commerce applications.", BaseURL: "https://fakestoreapi.com", PostmanURL: "https://www.postman.com/collections/7e5c794c43744a3aa3a173d0f77699e8", Tags: []string{"testing", "ecommerce", "developer tools"}, IsActive: true},
	}
}

// GetCategories returns a list of unique categories for public APIs.
func (s *PublicAPIService) GetCategories() []string {
	apis := s.GetPopularAPIs()
	categoryMap := make(map[string]bool)
	var categories []string
	for _, api := range apis {
		if !categoryMap[api.Category] {
			categoryMap[api.Category] = true
			categories = append(categories, api.Category)
		}
	}
	return categories
}

// ImportAndSavePublicPostmanCollection fetches a Postman collection by its UID (from the public API list),
// then saves it as a new Collection for the user, and generates its OpenAPI specification.
// It uses PostmanAPIService.GetCollection to fetch the raw JSON by UID.
func (s *PublicAPIService) ImportAndSavePublicPostmanCollection(ctx context.Context, userID, postmanCollectionUID, collectionNameOverride, collectionBaseURL string) (*models.Collection, string, error) {
	s.logger.Info("Importing public Postman collection by UID",
		zap.String("userID", userID),
		zap.String("postmanCollectionUID", postmanCollectionUID),
		zap.String("collectionNameOverride", collectionNameOverride))

	// Fetch the raw Postman JSON data using the Postman Collection UID
	rawPostmanJSON, err := s.postmanAPIService.GetCollection(ctx, postmanCollectionUID)
	if err != nil {
		s.logger.Error("Failed to fetch Postman collection using GetCollection by UID",
			zap.String("postmanCollectionUID", postmanCollectionUID),
			zap.Error(err))
		return nil, "", fmt.Errorf("failed to fetch Postman collection data for UID '%s': %w", postmanCollectionUID, err)
	}

	// Determine the collection name
	actualCollectionName := collectionNameOverride
	if actualCollectionName == "" {
		// Try to get the name from the public API entry details if available
		publicAPIEntry, _ := s.GetPublicAPIByPostmanID(postmanCollectionUID)
		if publicAPIEntry != nil && publicAPIEntry.Name != "" {
			actualCollectionName = publicAPIEntry.Name
		} else {
			// Fallback name if no override and not found in public list (or name is empty)
			actualCollectionName = "Imported Public API - " + postmanCollectionUID
		}
	}

	// Create the collection using CollectionService
	createReq := &models.CreateCollectionRequest{
		Name:        actualCollectionName,
		Description: fmt.Sprintf("Imported from Public API list (UID: %s)", postmanCollectionUID),
		PostmanData: rawPostmanJSON,
		// UserID is passed to CreateCollection method
	}

	createdCollection, err := s.collectionService.CreateCollection(createReq, userID)
	if err != nil {
		s.logger.Error("Failed to create collection from public API import via CollectionService", zap.Error(err))
		return nil, "", fmt.Errorf("failed to save imported Postman collection: %w", err)
	}

	// Generate OpenAPI spec using CollectionService
	_, openAPISpecContent, err := s.collectionService.GenerateOpenAPISpec(createdCollection.ID.Hex())
	if err != nil {
		s.logger.Error("Failed to generate OpenAPI spec for imported public API collection via CollectionService",
			zap.String("collectionID", createdCollection.ID.Hex()),
			zap.Error(err))
		return createdCollection, "", fmt.Errorf("collection created (ID: %s), but OpenAPI spec generation failed: %w", createdCollection.ID.Hex(), err)
	}

	s.logger.Info("Successfully imported and converted public API to collection and OpenAPI spec",
		zap.String("collectionID", createdCollection.ID.Hex()),
		zap.String("collectionName", createdCollection.Name))

	return createdCollection, openAPISpecContent, nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	postmanAPIBaseURL = "https://api.getpostman.com"
)

// PostmanAPIService handles interactions with the Postman Public API.
type PostmanAPIService struct {
	logger     *zap.Logger
	httpClient *http.Client
	apiKey     string
}

// NewPostmanAPIService creates a new PostmanAPIService.
// It now directly takes the API key, simplifying dependency on PlatformSettingsService for just the key.
func NewPostmanAPIService(logger *zap.Logger, apiKey string) *PostmanAPIService {
	return &PostmanAPIService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// getPostmanAPIKey is now a direct field access, but kept for consistency if logic were added.
func (s *PostmanAPIService) getPostmanAPIKey() string {
	if s.apiKey == "" {
		s.logger.Warn("Postman API key is not configured in PostmanAPIService")
	}
	return s.apiKey
}

// PostmanCollection represents the structure of a Postman collection from the API.
// This might need to be expanded based on the actual API response.
type PostmanCollection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	UID  string `json:"uid"` // uid is often used as the collection ID
}

// PostmanWorkspace represents a workspace from the Postman API.
type PostmanWorkspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListPublicWorkspaces fetches public workspaces.
// Note: The Postman API endpoint for truly "public workspaces of others" might be different or require specific permissions.
// This example assumes an endpoint like /workspaces. Adjust if needed.
func (s *PostmanAPIService) ListPublicWorkspaces(ctx context.Context) ([]PostmanWorkspace, error) {
	apiKey := s.getPostmanAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("postman API key not configured") // Keep as is, more of a status
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/workspaces", postmanAPIBaseURL), nil)
	if err != nil {
		s.logger.Error("Failed to create request for public workspaces", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch public workspaces", zap.Error(err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Postman API returned non-OK status for public workspaces", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("postman API error: status %d", resp.StatusCode)
	}

	var result struct {
		Workspaces []PostmanWorkspace `json:"workspaces"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Failed to decode public workspaces response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Workspaces, nil
}

// ListCollectionsInWorkspace fetches collections for a given workspace ID.
func (s *PostmanAPIService) ListCollectionsInWorkspace(ctx context.Context, workspaceID string) ([]PostmanCollection, error) {
	apiKey := s.getPostmanAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("postman API key not configured") // Keep as is
	}

	url := fmt.Sprintf("%s/collections?workspace=%s", postmanAPIBaseURL, workspaceID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create request for collections in workspace", zap.String("workspaceID", workspaceID), zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch collections in workspace", zap.String("workspaceID", workspaceID), zap.Error(err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Postman API returned non-OK status for collections in workspace", zap.Int("status", resp.StatusCode), zap.String("workspaceID", workspaceID))
		return nil, fmt.Errorf("postman API error: status %d for workspace %s", resp.StatusCode, workspaceID)
	}

	var result struct {
		Collections []PostmanCollection `json:"collections"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Failed to decode collections in workspace response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Collections, nil
}

// GetCollection fetches a specific Postman collection by its UID.
// It returns the raw JSON string of the collection.
func (s *PostmanAPIService) GetCollection(ctx context.Context, collectionUID string) (string, error) {
	apiKey := s.getPostmanAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("postman API key not configured") // Keep as is
	}

	collectionURL := fmt.Sprintf("%s/collections/%s", postmanAPIBaseURL, collectionUID)
	s.logger.Info("Fetching Postman collection", zap.String("url", collectionURL))

	req, err := http.NewRequestWithContext(ctx, "GET", collectionURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request for GetCollection", zap.String("collectionUID", collectionUID), zap.Error(err))
		return "", fmt.Errorf("failed to create request for collection %s: %w", collectionUID, err)
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch collection from Postman API", zap.String("collectionUID", collectionUID), zap.Error(err))
		return "", fmt.Errorf("failed to execute request for collection %s: %w", collectionUID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for error logging
		s.logger.Error("Postman API returned non-OK status for GetCollection",
			zap.String("collectionUID", collectionUID),
			zap.Int("status", resp.StatusCode),
			zap.String("responseBody", string(bodyBytes)))
		return "", fmt.Errorf("postman API error for collection %s: status %d, body: %s", collectionUID, resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body for GetCollection", zap.String("collectionUID", collectionUID), zap.Error(err))
		return "", fmt.Errorf("failed to read response body for collection %s: %w", collectionUID, err)
	}

	// The response from /collections/{uid} is expected to be the collection JSON itself,
	// wrapped in a top-level "collection" key.
	var responseWrapper struct {
		Collection json.RawMessage `json:"collection"`
	}
	if err := json.Unmarshal(bodyBytes, &responseWrapper); err != nil {
		s.logger.Error("Failed to unmarshal collection wrapper from Postman API response", zap.String("collectionUID", collectionUID), zap.Error(err), zap.String("rawResponse", string(bodyBytes)))
		// Fallback: if the response is already the collection itself (no wrapper)
		// This might happen if the API behavior changes or if the endpoint is different.
		// We can try to validate if bodyBytes is a valid JSON object that looks like a collection.
		// For now, we assume the wrapper or return the raw bytes if unmarshal fails but it looks like JSON.
		if strings.HasPrefix(strings.TrimSpace(string(bodyBytes)), "{") {
			s.logger.Warn("GetCollection: response was not wrapped in 'collection' field, attempting to use raw body.", zap.String("collectionUID", collectionUID))
			// Validate if this raw body is a Postman collection structure (e.g. has 'info' and 'item' keys)
			var tempCollectionCheck map[string]interface{}
			if json.Unmarshal(bodyBytes, &tempCollectionCheck) == nil {
				if _, hasInfo := tempCollectionCheck["info"]; hasInfo {
					if _, hasItem := tempCollectionCheck["item"]; hasItem {
						s.logger.Info("GetCollection: Raw body appears to be a valid collection.", zap.String("collectionUID", collectionUID))
						return string(bodyBytes), nil
					}
				}
			}
			s.logger.Error("GetCollection: Raw body does not appear to be a valid collection after failing to unmarshal wrapper.", zap.String("collectionUID", collectionUID))
		}
		return "", fmt.Errorf("failed to unmarshal collection wrapper for %s and raw body is not a valid collection: %w. Raw response: %s", collectionUID, err, string(bodyBytes))
	}

	if len(responseWrapper.Collection) == 0 {
		s.logger.Error("Postman API response for GetCollection had an empty 'collection' field", zap.String("collectionUID", collectionUID), zap.String("rawResponse", string(bodyBytes)))
		return "", fmt.Errorf("postman API returned empty collection data for %s. Raw response: %s", collectionUID, string(bodyBytes))
	}

	return string(responseWrapper.Collection), nil
}

// SearchCollections searches for public collections on Postman based on a query and limit.
func (s *PostmanAPIService) SearchCollections(ctx context.Context, query string, limit int) ([]PostmanCollection, error) {
	apiKey := s.getPostmanAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("postman API key not configured")
	}

	// Construct the search URL with query parameters
	searchURL := fmt.Sprintf("%s/search?type=collection&q=%s", postmanAPIBaseURL, url.QueryEscape(query))
	if limit > 0 {
		searchURL += fmt.Sprintf("&perPage=%d", limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request for searching collections", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to create request for search: %w", err)
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to search collections on Postman API", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.logger.Error("Postman API returned non-OK status for search collections",
			zap.String("query", query),
			zap.Int("status", resp.StatusCode),
			zap.String("responseBody", string(bodyBytes)))
		return nil, fmt.Errorf("postman API error for search '%s': status %d, body: %s", query, resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			UID  string `json:"uid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Failed to decode search collections response", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Convert the result data to PostmanCollection slice
	collections := make([]PostmanCollection, len(result.Data))
	for i, item := range result.Data {
		collections[i] = PostmanCollection{
			ID:   item.ID,
			Name: item.Name,
			UID:  item.UID,
		}
	}

	s.logger.Info("Successfully searched collections", zap.String("query", query), zap.Int("count", len(collections)))
	return collections, nil
}

// ImportCollectionByPostmanURL fetches a Postman collection given its public URL,
// extracts the collection UID, and then calls GetCollection.
// Returns the raw JSON string of the collection and the determined collection name.
func (s *PostmanAPIService) ImportCollectionByPostmanURL(ctx context.Context, postmanURLStr string, defaultName string) (string, string, error) {
	parsedURL, err := url.Parse(postmanURLStr)
	if err != nil {
		s.logger.Error("Invalid Postman URL", zap.String("url", postmanURLStr), zap.Error(err))
		return "", "", fmt.Errorf("invalid Postman URL %s: %w", postmanURLStr, err)
	}

	// Expected path formats:
	// 1. /collections/COLLECTION_UID
	// 2. /<workspace-name>/collection/COLLECTION_UID
	// 3. /<user>/<workspace>/collection/COLLECTION_UID/<collection-name>
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	var collectionUID string

	// Method 1: /collections/UID format
	if len(pathParts) > 0 && pathParts[0] == "collections" && len(pathParts) > 1 {
		collectionUID = pathParts[1]
	} else {
		// Method 2: Find "collection" keyword in path and extract the UID after it
		for i, part := range pathParts {
			if part == "collection" && i+1 < len(pathParts) {
				collectionUID = pathParts[i+1]
				break
			}
		}
	}

	if collectionUID == "" {
		s.logger.Error("Could not extract collection UID from Postman URL path", zap.String("urlPath", parsedURL.Path), zap.Any("pathParts", pathParts))
		return "", "", fmt.Errorf("could not extract collection UID from URL path: %s", parsedURL.Path)
	}

	if collectionUID == "" {
		return "", "", fmt.Errorf("failed to extract a valid collection UID from URL: %s", postmanURLStr)
	}

	// Fetch the collection data using the extracted UID
	collectionJSON, err := s.GetCollection(ctx, collectionUID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get collection %s using extracted UID: %w", collectionUID, err)
	}

	// Try to parse the name from the collection JSON itself
	var pmanCollection struct {
		Info struct {
			Name string `json:"name"`
		} `json:"info"`
	}
	collectionName := defaultName
	if err := json.Unmarshal([]byte(collectionJSON), &pmanCollection); err == nil && pmanCollection.Info.Name != "" {
		collectionName = pmanCollection.Info.Name
	} else if defaultName == "" {
		collectionName = "Imported Collection " + collectionUID // Fallback name
	}

	return collectionJSON, collectionName, nil
}

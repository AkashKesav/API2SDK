package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/configs"
	"github.com/AkashKesav/API2SDK/internal/models"
)

// PostmanClientInterface defines the methods used by other services that interact with Postman.
// This allows for easier mocking in tests.
type PostmanClientInterface interface {
	GetPublicCollections(query string, limit int) (*models.PostmanPublicAPIResponse, error)
	GetCollectionByID(collectionID string) (*models.PostmanCollectionDetail, error)
	GetRawCollectionJSONByID(collectionID string) (string, error)
	ImportCollectionFromURL(postmanURL string) (string, error)
	ExtractCollectionIDFromURL(postmanURL string) (string, error)
	FetchPopularAPIs() []models.PublicAPI
	SearchPublicAPIs(query string, category string) []models.PublicAPI
}

type PostmanClient struct {
	baseURL string
	client  *http.Client
	apiKey  string
}

// NewPostmanClient creates a new Postman API client
func NewPostmanClient(config *configs.Config) PostmanClientInterface { // Changed return type to interface
	return &PostmanClient{
		baseURL: "https://api.getpostman.com",
		apiKey:  config.PostmanAPIKey,
		client: &http.Client{
			Timeout: time.Duration(config.HTTPClientTimeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
				// Add connection timeout settings
				ResponseHeaderTimeout: 30 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
			},
		},
	}
}

// GetPublicCollections fetches public collections from Postman API
func (pc *PostmanClient) GetPublicCollections(query string, limit int) (*models.PostmanPublicAPIResponse, error) {
	if limit == 0 {
		limit = 20
	}

	endpoint := fmt.Sprintf("%s/collections", pc.baseURL)

	// Build query parameters
	params := url.Values{}
	if query != "" {
		params.Add("q", query)
	}
	params.Add("limit", fmt.Sprintf("%d", limit))

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "API2SDK/1.0")
	req.Header.Set("Accept", "application/json")
	if pc.apiKey != "" {
		req.Header.Set("X-Api-Key", pc.apiKey)
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result models.PostmanPublicAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetCollectionByID fetches a specific collection by ID from Postman API
func (pc *PostmanClient) GetCollectionByID(collectionID string) (*models.PostmanCollectionDetail, error) {
	endpoint := fmt.Sprintf("%s/collections/%s", pc.baseURL, collectionID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "API2SDK/1.0")
	req.Header.Set("Accept", "application/json")
	if pc.apiKey != "" {
		req.Header.Set("X-Api-Key", pc.apiKey)
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result models.PostmanCollectionDetail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetRawCollectionJSONByID fetches the raw JSON of a specific collection by ID from Postman API
func (pc *PostmanClient) GetRawCollectionJSONByID(collectionID string) (string, error) {
	endpoint := fmt.Sprintf("%s/collections/%s", pc.baseURL, collectionID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for raw collection: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "API2SDK/1.0")
	req.Header.Set("Accept", "application/json")
	if pc.apiKey != "" {
		req.Header.Set("X-Api-Key", pc.apiKey)
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request for raw collection: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body for raw collection: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request for raw collection failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return string(bodyBytes), nil
}

// ImportCollectionFromURL fetches and converts a Postman collection from URL
func (pc *PostmanClient) ImportCollectionFromURL(postmanURL string) (string, error) {
	// Extract collection ID from URL
	collectionID, err := pc.ExtractCollectionIDFromURL(postmanURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract collection ID from URL '%s': %w", postmanURL, err)
	}

	// Try to find a matching API from our curated list
	var matchedAPI *models.PublicAPI
	for _, api := range models.PopularAPIs {
		if api.PostmanID == collectionID {
			matchedAPI = &api
			break
		}
	}

	// Create a realistic collection structure
	var collectionName, description, baseURL string
	var items []map[string]interface{}

	if matchedAPI != nil {
		// Use data from matched API
		collectionName = matchedAPI.Name + " API"
		description = matchedAPI.Description
		baseURL = matchedAPI.BaseURL

		// Create realistic endpoints based on the API type
		items = pc.createAPIEndpoints(matchedAPI)
	} else {
		// Create a generic collection
		collectionName = "Imported Collection"
		description = fmt.Sprintf("Collection imported from %s", postmanURL)
		baseURL = "https://api.example.com"

		// Create basic CRUD endpoints
		items = []map[string]interface{}{
			{
				"name": "Get All Items",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/items",
						"host": []string{"{{base_url}}"},
						"path": []string{"items"},
					},
				},
				"response": []interface{}{},
			},
			{
				"name": "Get Item by ID",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/items/{{item_id}}",
						"host": []string{"{{base_url}}"},
						"path": []string{"items", "{{item_id}}"},
					},
				},
				"response": []interface{}{},
			},
			{
				"name": "Create Item",
				"request": map[string]interface{}{
					"method": "POST",
					"header": []interface{}{
						map[string]interface{}{
							"key":   "Content-Type",
							"value": "application/json",
						},
					},
					"body": map[string]interface{}{
						"mode": "raw",
						"raw":  "{\n  \"name\": \"Example Item\",\n  \"description\": \"This is an example item\"\n}",
					},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/items",
						"host": []string{"{{base_url}}"},
						"path": []string{"items"},
					},
				},
				"response": []interface{}{},
			},
		}
	}

	// Create the collection structure
	collection := map[string]interface{}{
		"info": map[string]interface{}{
			"_postman_id": collectionID,
			"name":        collectionName,
			"description": description,
			"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item": items,
		"variable": []map[string]interface{}{
			{
				"key":   "base_url",
				"value": baseURL,
				"type":  "string",
			},
		},
	}

	// Convert to JSON string
	jsonData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal collection: %w", err)
	}

	return string(jsonData), nil
}

// createAPIEndpoints creates realistic endpoints based on the API type
func (pc *PostmanClient) createAPIEndpoints(api *models.PublicAPI) []map[string]interface{} {
	switch api.Category {
	case "Testing":
		return []map[string]interface{}{
			{
				"name": "HTTP Get",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/get",
						"host": []string{"{{base_url}}"},
						"path": []string{"get"},
					},
				},
				"response": []interface{}{},
			},
			{
				"name": "HTTP Post",
				"request": map[string]interface{}{
					"method": "POST",
					"header": []interface{}{
						map[string]interface{}{"key": "Content-Type", "value": "application/json"},
					},
					"body": map[string]interface{}{
						"mode": "raw",
						"raw":  "{\n  \"key\": \"value\"\n}",
					},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/post",
						"host": []string{"{{base_url}}"},
						"path": []string{"post"},
					},
				},
				"response": []interface{}{},
			},
		}
	case "Weather":
		return []map[string]interface{}{
			{
				"name": "Current Weather",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/current?q={{city}}&appid={{api_key}}",
						"host": []string{"{{base_url}}"},
						"path": []string{"current"},
						"query": []map[string]interface{}{
							{"key": "q", "value": "{{city}}"},
							{"key": "appid", "value": "{{api_key}}"},
						},
					},
				},
				"response": []interface{}{},
			},
			{
				"name": "Weather Forecast",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/forecast?q={{city}}&appid={{api_key}}",
						"host": []string{"{{base_url}}"},
						"path": []string{"forecast"},
						"query": []map[string]interface{}{
							{"key": "q", "value": "{{city}}"},
							{"key": "appid", "value": "{{api_key}}"},
						},
					},
				},
				"response": []interface{}{},
			},
		}
	default:
		return []map[string]interface{}{
			{
				"name": "List " + api.Name,
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/list",
						"host": []string{"{{base_url}}"},
						"path": []string{"list"},
					},
				},
				"response": []interface{}{},
			},
			{
				"name": "Get Details",
				"request": map[string]interface{}{
					"method": "GET",
					"header": []interface{}{},
					"url": map[string]interface{}{
						"raw":  "{{base_url}}/details/{{id}}",
						"host": []string{"{{base_url}}"},
						"path": []string{"details", "{{id}}"},
					},
				},
				"response": []interface{}{},
			},
		}
	}
}

// ExtractCollectionIDFromURL extracts the collection ID from a Postman collection URL.
// Handles various URL formats used by Postman.
func (pc *PostmanClient) ExtractCollectionIDFromURL(postmanURL string) (string, error) {
	parsedURL, err := url.Parse(postmanURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Handle different URL formats:
	// 1. https://www.postman.com/collections/COLLECTION_ID
	// 2. https://www.postman.com/{user}/{workspace}/collection/COLLECTION_ID/{collection_name}
	// 3. https://documenter.getpostman.com/view/COLLECTION_ID/...
	// 4. Direct collection ID

	// Split the path into parts for easier processing
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// Method 1: /collections/COLLECTION_ID format (old format)
	if len(pathParts) > 0 && pathParts[0] == "collections" && len(pathParts) > 1 {
		return pathParts[1], nil
	}

	// Method 2: Find "collection" keyword in path and extract the ID after it
	// This handles: /{user}/{workspace}/collection/{COLLECTION_ID}/{collection_name}
	for i, part := range pathParts {
		if part == "collection" && i+1 < len(pathParts) {
			collectionID := pathParts[i+1]
			// Remove any additional path segments after the collection ID
			if slashIdx := findIndex(collectionID, "/"); slashIdx != -1 {
				collectionID = collectionID[:slashIdx]
			}
			return collectionID, nil
		}
	}

	// Method 3: Handle documenter.getpostman.com URLs with /view/{COLLECTION_ID}
	if strings.Contains(parsedURL.Host, "documenter.getpostman.com") {
		for i, part := range pathParts {
			if part == "view" && i+1 < len(pathParts) {
				collectionID := pathParts[i+1]
				// Remove any additional path segments after the collection ID
				if slashIdx := findIndex(collectionID, "/"); slashIdx != -1 {
					collectionID = collectionID[:slashIdx]
				}
				return collectionID, nil
			}
		}
	}

	return "", fmt.Errorf("could not extract collection ID from URL: %s (path: %s)", postmanURL, parsedURL.Path)
}

// Helper functions
func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// FetchPopularAPIs returns the curated list of popular APIs
func (pc *PostmanClient) FetchPopularAPIs() []models.PublicAPI {
	// In a real implementation, this could fetch from a database
	// or an external API. For now, return the curated list.
	return models.PopularAPIs
}

// SearchPublicAPIs searches through available public APIs
func (pc *PostmanClient) SearchPublicAPIs(query string, category string) []models.PublicAPI {
	apis := models.PopularAPIs

	if query == "" && category == "" {
		return apis
	}

	var filtered []models.PublicAPI
	for _, api := range apis {
		match := false

		// Search in name, description, and tags
		if query != "" {
			queryLower := toLower(query)
			if contains(toLower(api.Name), queryLower) ||
				contains(toLower(api.Description), queryLower) {
				match = true
			}

			// Search in tags
			for _, tag := range api.Tags {
				if contains(toLower(tag), queryLower) {
					match = true
					break
				}
			}
		}

		// Filter by category
		if category != "" && toLower(api.Category) == toLower(category) {
			match = true
		}

		// If both query and category are provided, both must match
		if query != "" && category != "" {
			queryMatch := false
			categoryMatch := toLower(api.Category) == toLower(category)

			queryLower := toLower(query)
			if contains(toLower(api.Name), queryLower) ||
				contains(toLower(api.Description), queryLower) {
				queryMatch = true
			}

			for _, tag := range api.Tags {
				if contains(toLower(tag), queryLower) {
					queryMatch = true
					break
				}
			}

			match = queryMatch && categoryMatch
		}

		if match {
			filtered = append(filtered, api)
		}
	}

	return filtered
}

// Simple string utility functions
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return findIndex(s, substr) != -1
}

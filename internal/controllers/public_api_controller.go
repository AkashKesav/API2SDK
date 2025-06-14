package controllers

import (
	"fmt"
	"strconv"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type PublicAPIController struct {
	service *services.PublicAPIService
	logger  *zap.Logger
}

func NewPublicAPIController(service *services.PublicAPIService, logger *zap.Logger) *PublicAPIController {
	return &PublicAPIController{
		service: service,
		logger:  logger,
	}
}

// GetPublicAPIs handles GET /public-apis
func (pac *PublicAPIController) GetPublicAPIs(c fiber.Ctx) error {
	// Get query parameters
	query := c.Query("q", "")
	category := c.Query("category", "")
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "20")

	// Parse pagination parameters
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	// Get public APIs
	apis, total, err := pac.service.GetAllPublicAPIs(query, category, page, limit)
	if err != nil {
		pac.logger.Error("Failed to retrieve public APIs", zap.Error(err), zap.String("query", query), zap.String("category", category))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve public APIs", err.Error())
	}

	// Get categories for filtering
	categories := pac.service.GetCategories()

	return utils.SuccessResponse(c, "Public APIs retrieved successfully", fiber.Map{
		"apis":       apis,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"categories": categories,
	})
}

// GetPublicAPIByID handles GET /public-apis/:id  (Note: Renamed from GetPublicAPI to avoid conflict with struct name if it were a package func)
func (pac *PublicAPIController) GetPublicAPIByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Public API ID is required", "")
	}

	// Special case for popular APIs - return from curated list
	if id == "popular" {
		apis := pac.service.GetPopularAPIs()
		return utils.SuccessResponse(c, "Popular APIs retrieved successfully", apis)
	}

	api, err := pac.service.GetPublicAPIByID(id)
	if err != nil {
		pac.logger.Warn("Public API not found", zap.String("id", id), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Public API not found", err.Error())
	}

	return utils.SuccessResponse(c, "Public API retrieved successfully", api)
}

// CreatePublicAPI handles POST /public-apis
func (pac *PublicAPIController) CreatePublicAPI(c fiber.Ctx) error {
	var req models.CreatePublicAPIRequest

	if err := c.Bind().Body(&req); err != nil {
		pac.logger.Error("Invalid request body for CreatePublicAPI", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate required fields
	if req.Name == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "API name is required", "")
	}

	if req.BaseURL == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Base URL is required", "")
	}

	api, err := pac.service.CreatePublicAPI(&req)
	if err != nil {
		pac.logger.Error("Failed to create public API", zap.Any("request", req), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create public API", err.Error())
	}

	return utils.SuccessResponse(c, "Public API created successfully", api)
}

// UpdatePublicAPI handles PUT /public-apis/:id
func (pac *PublicAPIController) UpdatePublicAPI(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Public API ID is required", "")
	}

	var req models.UpdatePublicAPIRequest
	if err := c.Bind().Body(&req); err != nil {
		pac.logger.Error("Invalid request body for UpdatePublicAPI", zap.String("id", id), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	api, err := pac.service.UpdatePublicAPI(id, &req)
	if err != nil {
		pac.logger.Error("Failed to update public API", zap.String("id", id), zap.Any("request", req), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update public API", err.Error())
	}

	return utils.SuccessResponse(c, "Public API updated successfully", api)
}

// DeletePublicAPI handles DELETE /public-apis/:id
func (pac *PublicAPIController) DeletePublicAPI(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Public API ID is required", "")
	}

	err := pac.service.DeletePublicAPI(id)
	if err != nil {
		pac.logger.Error("Failed to delete public API", zap.String("id", id), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete public API", err.Error())
	}

	return utils.SuccessResponse(c, "Public API deleted successfully", nil)
}

// ImportFromPostman handles POST /public-apis/import
func (pac *PublicAPIController) ImportFromPostman(c fiber.Ctx) error {
	var req struct {
		PostmanURL string `json:"postman_url"`
		Name       string `json:"name,omitempty"` // Optional name override
	}

	if err := c.Bind().Body(&req); err != nil {
		pac.logger.Error("Invalid request body for ImportFromPostman", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.PostmanURL == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Postman URL is required", "")
	}

	userID := "" // Default to empty or handle unauthenticated case
	userIDVal := c.Locals("userID")
	if userIDVal != nil {
		if idStr, ok := userIDVal.(string); ok {
			userID = idStr
		} else {
			pac.logger.Warn("userID in context is not a string", zap.Any("userID", userIDVal))
			// Decide if this is an error or if userID can remain empty
		}
	}

	// Use the accessor methods on pac.service (PublicAPIService instance)
	rawJSON, collectionName, err := pac.service.PostmanService().ImportCollectionByPostmanURL(c.Context(), req.PostmanURL, req.Name)
	if err != nil {
		pac.logger.Error("Failed to import from Postman URL via service", zap.String("postmanURL", req.PostmanURL), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import from Postman URL", err.Error())
	}

	createReq := &models.CreateCollectionRequest{
		Name:        collectionName, // Name determined by ImportCollectionByPostmanURL
		Description: fmt.Sprintf("Imported from Postman URL: %s", req.PostmanURL),
		PostmanData: rawJSON,
	}

	createdCollection, err := pac.service.CollectionService().CreateCollection(createReq, userID)
	if err != nil {
		pac.logger.Error("Failed to save collection imported from Postman URL", zap.String("postmanURL", req.PostmanURL), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save imported collection", err.Error())
	}

	return utils.SuccessResponse(c, "Collection imported successfully from Postman URL", fiber.Map{
		"collection": createdCollection,
	})
}

// SearchPublicCollections handles GET /public-apis/search
func (pac *PublicAPIController) SearchPublicCollections(c fiber.Ctx) error {
	query := c.Query("q", "")
	limitStr := c.Query("limit", "20")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	// Search Postman's public collections
	result, err := pac.service.SearchPublicCollections(query, limit)
	if err != nil {
		pac.logger.Error("Failed to search public collections", zap.String("query", query), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to search public collections", err.Error())
	}

	return utils.SuccessResponse(c, "Public collections searched successfully", result)
}

// GetPopularAPIs handles GET /public-apis/popular
func (pac *PublicAPIController) GetPopularAPIs(c fiber.Ctx) error {
	apis := pac.service.GetPopularAPIs()
	return utils.SuccessResponse(c, "Popular APIs retrieved successfully", apis)
}

// GetCategories handles GET /public-apis/categories
func (pac *PublicAPIController) GetCategories(c fiber.Ctx) error {
	categories := pac.service.GetCategories()
	return utils.SuccessResponse(c, "Categories retrieved successfully", categories)
}

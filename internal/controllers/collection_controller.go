package controllers

import (
	"encoding/json"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CollectionController struct {
	service *services.CollectionService
	logger  *zap.Logger
}

func NewCollectionController(service *services.CollectionService, logger *zap.Logger) *CollectionController {
	return &CollectionController{
		service: service,
		logger:  logger,
	}
}

// GetUserCollections handles GET /collections - specific to a user
func (cc *CollectionController) GetUserCollections(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context or userID is not of type primitive.ObjectID")
		return utils.UnauthorizedResponse(c, "Unauthorized or invalid user ID type")
	}

	collections, err := cc.service.GetCollectionsByUserID(userID.Hex())
	if err != nil {
		cc.logger.Error("Failed to retrieve collections for user", zap.String("userID", userID.Hex()), zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to retrieve collections", err.Error())
	}

	return utils.SuccessResponse(c, "Collections retrieved successfully", fiber.Map{
		"collections": collections,
		"count":       len(collections),
	})
}

// GetCollectionByID handles GET /collections/:id
func (cc *CollectionController) GetCollectionByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequestResponse(c, "Collection ID is required", "")
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context or userID is not of type primitive.ObjectID")
		return utils.UnauthorizedResponse(c, "Unauthorized or invalid user ID type")
	}

	// UserID is checked for authorization but not passed to the service's GetCollection method
	collection, err := cc.service.GetCollection(id)
	if err != nil {
		cc.logger.Error("Failed to retrieve collection by ID", zap.String("collectionID", id), zap.String("userID", userID.Hex()), zap.Error(err)) // Log userID for context
		return utils.InternalServerErrorResponse(c, "Failed to retrieve collection", err.Error())
	}
	if collection == nil { // Handle case where collection is not found
		return utils.NotFoundResponse(c, "Collection not found")
	}
	// Optional: Add an explicit check: if collection.UserID != userID.Hex() { return unauthorized }

	return utils.SuccessResponse(c, "Collection retrieved successfully", collection)
}

// CreateCollection handles POST /collections
func (cc *CollectionController) CreateCollection(c fiber.Ctx) error {
	var req models.CreateCollectionRequest

	// Handle multipart form data from HTMX
	contentType := c.Get("Content-Type")
	if contentType != "" && (contentType == "application/x-www-form-urlencoded" ||
		(len(contentType) > 19 && contentType[:19] == "multipart/form-data")) {

		req.Name = c.FormValue("name")
		req.Description = c.FormValue("description")

		// Check if file was uploaded
		file, err := c.FormFile("file")
		if err == nil && file != nil {
			// Read file content
			fileContent, err := file.Open()
			if err != nil {
				cc.logger.Error("Failed to open uploaded file", zap.Error(err))
				return utils.BadRequestResponse(c, "Failed to read uploaded file", err.Error())
			}
			defer fileContent.Close()

			// Read file as string
			fileBytes := make([]byte, file.Size)
			_, err = fileContent.Read(fileBytes)
			if err != nil {
				cc.logger.Error("Failed to read file content", zap.Error(err))
				return utils.BadRequestResponse(c, "Failed to read file content", err.Error())
			}

			req.PostmanData = string(fileBytes)
		} else {
			// Check for JSON data in form
			postmanData := c.FormValue("postman_data")
			if postmanData != "" {
				req.PostmanData = postmanData
			}
		}
	} else {
		// Handle JSON request
		body := c.Body()
		if len(body) == 0 {
			cc.logger.Error("Request body is empty for CreateCollection")
			return utils.BadRequestResponse(c, "Invalid request body", "Request body is empty")
		}
		if err := json.Unmarshal(body, &req); err != nil {
			cc.logger.Error("Invalid request body for CreateCollection", zap.Error(err))
			return utils.BadRequestResponse(c, "Invalid request body", err.Error())
		}
	}

	if req.Name == "" {
		return utils.BadRequestResponse(c, "Collection name is required", "")
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context or userID is not of type primitive.ObjectID for CreateCollection")
		return utils.UnauthorizedResponse(c, "Unauthorized or invalid user ID type")
	}

	collection, err := cc.service.CreateCollection(&req, userID.Hex())
	if err != nil {
		cc.logger.Error("Failed to create collection", zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to create collection", err.Error())
	}

	return utils.CreatedResponse(c, "Collection created successfully", collection)
}

// UpdateCollection handles PUT /collections/:id
func (cc *CollectionController) UpdateCollection(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequestResponse(c, "Collection ID is required", "")
	}

	var req models.UpdateCollectionRequest
	body := c.Body()
	if len(body) == 0 {
		cc.logger.Error("Request body is empty for UpdateCollection", zap.String("collectionID", id))
		return utils.BadRequestResponse(c, "Invalid request body", "Request body is empty")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		cc.logger.Error("Invalid request body for UpdateCollection", zap.String("collectionID", id), zap.Error(err))
		return utils.BadRequestResponse(c, "Invalid request body", err.Error())
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context or userID is not of type primitive.ObjectID for UpdateCollection")
		return utils.UnauthorizedResponse(c, "Unauthorized or invalid user ID type")
	}

	// UserID is checked for authorization, but not passed to the service's UpdateCollection method.
	// First, get the existing collection to verify ownership if needed.
	existingCollection, err := cc.service.GetCollection(id)
	if err != nil {
		cc.logger.Error("Failed to retrieve collection for update check", zap.String("collectionID", id), zap.String("userID", userID.Hex()), zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to retrieve collection for update", err.Error())
	}
	if existingCollection == nil {
		return utils.NotFoundResponse(c, "Collection not found")
	}
	if existingCollection.UserID != userID.Hex() {
		cc.logger.Warn("User attempted to update collection they do not own", zap.String("collectionID", id), zap.String("ownerUserID", existingCollection.UserID), zap.String("requestingUserID", userID.Hex()))
		return utils.ForbiddenResponse(c, "Forbidden: You do not own this collection")
	}

	collection, err := cc.service.UpdateCollection(id, &req)
	if err != nil {
		cc.logger.Error("Failed to update collection", zap.String("collectionID", id), zap.String("userID", userID.Hex()), zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to update collection", err.Error())
	}

	return utils.SuccessResponse(c, "Collection updated successfully", collection)
}

// DeleteCollection handles DELETE /collections/:id
func (cc *CollectionController) DeleteCollection(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequestResponse(c, "Collection ID is required", "")
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context or userID is not of type primitive.ObjectID for DeleteCollection")
		return utils.UnauthorizedResponse(c, "Unauthorized or invalid user ID type")
	}

	// UserID is checked for authorization, but not passed to the service's DeleteCollection method.
	// First, get the existing collection to verify ownership.
	existingCollection, err := cc.service.GetCollection(id)
	if err != nil {
		cc.logger.Error("Failed to retrieve collection for delete check", zap.String("collectionID", id), zap.String("userID", userID.Hex()), zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to retrieve collection for delete", err.Error())
	}
	if existingCollection == nil {
		return utils.NotFoundResponse(c, "Collection not found")
	}
	if existingCollection.UserID != userID.Hex() {
		cc.logger.Warn("User attempted to delete collection they do not own", zap.String("collectionID", id), zap.String("ownerUserID", existingCollection.UserID), zap.String("requestingUserID", userID.Hex()))
		return utils.ForbiddenResponse(c, "Forbidden: You do not own this collection")
	}

	err = cc.service.DeleteCollection(id)
	if err != nil {
		cc.logger.Error("Failed to delete collection", zap.String("collectionID", id), zap.String("userID", userID.Hex()), zap.Error(err))
		return utils.InternalServerErrorResponse(c, "Failed to delete collection", err.Error())
	}

	return utils.SuccessResponse(c, "Collection deleted successfully", nil)
}

// GenerateOpenAPISpec handles POST /collections/:id/generate-openapi-spec
func (cc *CollectionController) GenerateOpenAPISpec(c fiber.Ctx) error {
	collectionID := c.Params("id")
	if collectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Collection ID is required"})
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context for GenerateOpenAPISpec")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if user owns the collection before generating spec
	existingCollection, err := cc.service.GetCollection(collectionID)
	if err != nil {
		cc.logger.Error("Failed to retrieve collection for OpenAPI spec generation check", zap.String("collectionID", collectionID), zap.String("userID", userID.Hex()), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve collection", "details": err.Error()})
	}
	if existingCollection == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Collection not found"})
	}
	if existingCollection.UserID != userID.Hex() {
		cc.logger.Warn("User attempted to generate spec for collection they do not own", zap.String("collectionID", collectionID), zap.String("ownerUserID", existingCollection.UserID), zap.String("requestingUserID", userID.Hex()))
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: You do not own this collection"})
	}

	_, specContent, err := cc.service.GenerateOpenAPISpec(collectionID) // Service returns filePath, specContent, error
	if err != nil {
		cc.logger.Error("Failed to generate OpenAPI spec from collection", zap.String("collectionID", collectionID), zap.String("userID", userID.Hex()), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate OpenAPI spec", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "OpenAPI spec generated successfully",
		"data":    specContent, // Return the spec content
	})
}

// CreateCollectionFromPublicAPI handles POST /collections/from-public-api
// This functionality is not currently implemented in the CollectionService.
// Commenting out for now to resolve compiler errors.
/*
func (cc *CollectionController) CreateCollectionFromPublicAPI(c fiber.Ctx) error {
	var req struct {
		PublicAPIID string `json:"public_api_id"`
		Name        string `json:"name"` // Optional: name for the new collection
	}
	if err := c.Bind().Body(&req); err != nil {
		cc.logger.Error("Invalid request body for CreateCollectionFromPublicAPI", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.PublicAPIID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Public API ID is required"})
	}

	userID, ok := c.Locals("userID").(primitive.ObjectID)
	if !ok {
		cc.logger.Error("Failed to get userID from context for CreateCollectionFromPublicAPI")
return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
}

// collection, err := cc.service.CreateCollectionFromPublicAPI(req.PublicAPIID, userID.Hex(), req.Name) // userID needs to be string
// if err != nil {
//  cc.logger.Error("Failed to create collection from public API", zap.String("publicAPIID", req.PublicAPIID), zap.Error(err))
//  return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create collection from public API", "details": err.Error()})
// }
// return c.Status(fiber.StatusCreated).JSON(fiber.Map{
//  "success": true,
//  "message": "Collection created successfully from public API",
//  "data":    collection,
// })
return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "This endpoint is not yet implemented"})
}
*/

// Placeholder for other methods if any, or remove if not needed.
// For example, if there was a global GetCollections (not user-specific)
// func (cc *CollectionController) GetAllCollections(c fiber.Ctx) error { ... }

package controllers

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// SDKController handles HTTP requests related to SDKs.
type SDKController struct {
	sdkService              services.SDKServiceInterface
	collectionService       *services.CollectionService // Added CollectionService
	platformSettingsService services.PlatformSettingsService
	logger                  *zap.Logger
	validate                *validator.Validate // Added validator instance
}

// NewSDKController creates a new SDKController.
func NewSDKController(sdkService services.SDKServiceInterface, collectionService *services.CollectionService, platformSettingsService services.PlatformSettingsService, logger *zap.Logger) *SDKController {
	return &SDKController{
		sdkService:              sdkService,
		collectionService:       collectionService, // Initialize CollectionService
		platformSettingsService: platformSettingsService,
		logger:                  logger,
		validate:                validator.New(), // Initialize validator
	}
}

// GenerateSDK handles the request to generate an SDK from a collection.
func (ctrl *SDKController) GenerateSDK(c fiber.Ctx) error {
	ctrl.logger.Info("GenerateSDK endpoint hit")

	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		ctrl.logger.Error("Failed to get userID from context")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	var req models.SDKGenerationRequest
	body := c.Body()
	if len(body) == 0 {
		ctrl.logger.Error("Request body is empty for SDK generation")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload", "Request body is empty")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		ctrl.logger.Error("Failed to parse request body for SDK generation", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if err := ctrl.validate.Struct(req); err != nil {
		ctrl.logger.Error("SDK generation request validation failed", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	_, err := ctrl.collectionService.GetCollectionByIDAndUser(context.Background(), req.CollectionID, userIDStr)
	if err != nil {
		ctrl.logger.Error("Failed to verify collection ownership or collection not found", zap.String("collectionID", req.CollectionID), zap.String("userID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Access to collection denied or collection not found", err.Error())
	}

	initialSDKRecord := &models.SDK{
		UserID:       userIDStr,
		CollectionID: req.CollectionID,
		Language:     req.Language,
		PackageName:  req.PackageName,
		Status:       models.SDKStatusPending,
	}

	createdRecord, err := ctrl.sdkService.CreateSDKRecord(context.Background(), initialSDKRecord)
	if err != nil {
		ctrl.logger.Error("Failed to create initial SDK record", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize SDK generation", err.Error())
	}
	recordID := createdRecord.ID

	go func() {
		bgCtx := context.Background()
		generatedSDK, genErr := ctrl.sdkService.GenerateSDK(bgCtx, &req, recordID)

		if genErr != nil {
			ctrl.logger.Error("SDK generation failed in goroutine", zap.Error(genErr), zap.String("recordID", recordID.Hex()))
			if updateErr := ctrl.sdkService.UpdateSDKStatus(bgCtx, recordID, models.SDKStatusFailed, genErr.Error()); updateErr != nil {
				ctrl.logger.Error("Failed to update SDK status to failed after generation error", zap.Error(updateErr), zap.String("recordID", recordID.Hex()))
			}
			return
		}

		ctrl.logger.Info("SDK generated successfully in goroutine", zap.String("recordID", recordID.Hex()), zap.Stringp("filePath", &generatedSDK.FilePath))
		if updateErr := ctrl.sdkService.UpdateSDKRecord(bgCtx, generatedSDK); updateErr != nil {
			ctrl.logger.Error("Failed to update SDK record after successful generation", zap.Error(updateErr), zap.String("recordID", recordID.Hex()))
		}
	}()

	return utils.SuccessResponse(c, "SDK generation started successfully. You will be notified upon completion.", createdRecord)
}

// GenerateMCP handles the request to generate an MCP server.
func (ctrl *SDKController) GenerateMCP(c fiber.Ctx) error {
	ctrl.logger.Info("GenerateMCP endpoint hit")

	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		ctrl.logger.Error("Failed to get userID from context")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	var req models.MCPGenerationRequest
	if err := c.Bind().Body(&req); err != nil {
		ctrl.logger.Error("Failed to parse request body for MCP generation", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if err := ctrl.validate.Struct(req); err != nil {
		ctrl.logger.Error("MCP generation request validation failed", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	_, err := ctrl.collectionService.GetCollectionByIDAndUser(context.Background(), req.CollectionID, userIDStr)
	if err != nil {
		ctrl.logger.Error("Failed to verify collection ownership or collection not found for MCP gen", zap.String("collectionID", req.CollectionID), zap.String("userID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Access to collection denied or collection not found", err.Error())
	}

	initialSDKRecord := &models.SDK{
		UserID:         userIDStr,
		CollectionID:   req.CollectionID,
		GenerationType: models.GenerationTypeMCP,
		Status:         models.SDKStatusPending,
		MCPTransport:   req.Transport,
		MCPPort:        req.Port,
	}

	createdRecord, err := ctrl.sdkService.CreateSDKRecord(context.Background(), initialSDKRecord)
	if err != nil {
		ctrl.logger.Error("Failed to create initial MCP record", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize MCP generation", err.Error())
	}
	recordID := createdRecord.ID

	go func() {
		bgCtx := context.Background()
		generatedMCP, genErr := ctrl.sdkService.GenerateMCP(bgCtx, &req, recordID)

		if genErr != nil {
			ctrl.logger.Error("MCP generation failed in goroutine", zap.Error(genErr), zap.String("recordID", recordID.Hex()))
			if updateErr := ctrl.sdkService.UpdateSDKStatus(bgCtx, recordID, models.SDKStatusFailed, genErr.Error()); updateErr != nil {
				ctrl.logger.Error("Failed to update MCP status to failed after generation error", zap.Error(updateErr), zap.String("recordID", recordID.Hex()))
			}
			return
		}

		ctrl.logger.Info("MCP generated successfully in goroutine", zap.String("recordID", recordID.Hex()), zap.Stringp("filePath", &generatedMCP.FilePath))
		if updateErr := ctrl.sdkService.UpdateSDKRecord(bgCtx, generatedMCP); updateErr != nil {
			ctrl.logger.Error("Failed to update MCP record after successful generation", zap.Error(updateErr), zap.String("recordID", recordID.Hex()))
		}
	}()

	return utils.SuccessResponse(c, "MCP generation started successfully. You will be notified upon completion.", createdRecord)
}

// GetSDKByID handles the request to retrieve an SDK by its ID.
func (ctrl *SDKController) GetSDKByID(c fiber.Ctx) error {
	sdkID := c.Params("id")
	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	ctrl.logger.Info("GetSDKByID request",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	objectSdkID, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		ctrl.logger.Error("Invalid SDK ID format", zap.String("sdkID", sdkID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid SDK ID format", err.Error())
	}

	sdk, err := ctrl.sdkService.GetSDKByID(context.Background(), objectSdkID, userIDStr)
	if err != nil {
		ctrl.logger.Error("Failed to retrieve SDK", zap.Error(err), zap.String("sdkID", sdkID), zap.String("userID", userIDStr))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "SDK not found or access denied", err.Error())
	}

	return utils.SuccessResponse(c, "SDK retrieved successfully", sdk)
}

// ListSDKsForUser handles the request to list all SDKs for the authenticated user.
func (ctrl *SDKController) ListSDKsForUser(c fiber.Ctx) error {
	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	ctrl.logger.Info("ListSDKsForUser request",
		zap.String("userID", userIDStr),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("statusFilter", statusFilter),
		zap.String("typeFilter", typeFilter))

	sdks, total, err := ctrl.sdkService.GetSDKsByUserID(context.Background(), userIDStr, page, limit)
	if err != nil {
		ctrl.logger.Error("Failed to list SDKs for user", zap.Error(err), zap.String("userID", userIDStr))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve SDKs", err.Error())
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	pagination := models.Pagination{
		CurrentPage: page,
		Limit:       limit,
		TotalItems:  total,
		TotalPages:  totalPages,
	}

	response := models.PaginatedSDKsResponse{
		SDKs:       sdks,
		Pagination: pagination,
	}

	return utils.SuccessResponse(c, "SDKs retrieved successfully", response)
}

// DownloadSDK handles the request to download an SDK.
func (ctrl *SDKController) DownloadSDK(c fiber.Ctx) error {
	sdkID := c.Params("id")
	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	ctrl.logger.Info("DownloadSDK request",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	objectSdkID, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		ctrl.logger.Error("Invalid SDK ID format for download", zap.String("sdkID", sdkID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid SDK ID format", err.Error())
	}

	sdk, serviceErr := ctrl.sdkService.GetSDKByID(context.Background(), objectSdkID, userIDStr)
	if serviceErr != nil {
		ctrl.logger.Error("Failed to retrieve SDK for download", zap.Error(serviceErr), zap.String("sdkID", sdkID), zap.String("userID", userIDStr))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "SDK not found or access denied", serviceErr.Error())
	}

	if sdk.Status != models.SDKStatusCompleted {
		ctrl.logger.Warn("Attempt to download SDK not in completed state", zap.String("sdkID", sdkID), zap.String("status", string(sdk.Status)))
		return utils.ErrorResponse(c, fiber.StatusConflict, "SDK is not ready for download", "SDK generation is not complete or has failed.")
	}

	if sdk.FilePath == "" {
		ctrl.logger.Error("SDK file path is empty for a completed SDK", zap.String("sdkID", sdkID))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "SDK file not available", "The SDK file path is missing.")
	}

	var downloadFilename string
	if sdk.GenerationType == models.GenerationTypeMCP {
		downloadFilename = "mcp-server-" + strings.ReplaceAll(strings.ToLower(sdk.CollectionID), " ", "-") + ".zip"
		if sdk.PackageName != "" {
			// Sanitize package name to prevent path traversal
			sanitizedPackageName := strings.ReplaceAll(sdk.PackageName, "..", "")
			sanitizedPackageName = strings.ReplaceAll(sanitizedPackageName, "/", "")
			sanitizedPackageName = strings.ReplaceAll(sanitizedPackageName, "\\", "")
			downloadFilename = strings.ReplaceAll(strings.ToLower(sanitizedPackageName), " ", "-") + ".zip"
		}
	} else {
		// Sanitize package name to prevent path traversal
		sanitizedPackageName := strings.ReplaceAll(sdk.PackageName, "..", "")
		sanitizedPackageName = strings.ReplaceAll(sanitizedPackageName, "/", "")
		sanitizedPackageName = strings.ReplaceAll(sanitizedPackageName, "\\", "")
		downloadFilename = strings.ReplaceAll(strings.ToLower(sanitizedPackageName), " ", "-") + ".zip"
	}

	ctrl.logger.Info("Attempting to download SDK file", zap.String("filePath", sdk.FilePath), zap.String("downloadAs", downloadFilename))
	return c.Download(sdk.FilePath, downloadFilename)
}

// DeleteSDK handles the request to delete an SDK.
func (ctrl *SDKController) DeleteSDK(c fiber.Ctx) error {
	sdkID := c.Params("id")
	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	ctrl.logger.Info("DeleteSDK request",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	objectSdkID, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		ctrl.logger.Error("Invalid SDK ID format for deletion", zap.String("sdkID", sdkID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid SDK ID format", err.Error())
	}

	serviceErr := ctrl.sdkService.DeleteSDK(context.Background(), objectSdkID, userIDStr)
	if serviceErr != nil {
		ctrl.logger.Error("Failed to delete SDK", zap.Error(serviceErr), zap.String("sdkID", sdkID), zap.String("userID", userIDStr))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete SDK", serviceErr.Error())
	}

	return utils.SuccessResponse(c, "SDK deleted successfully", nil)
}

// GetSupportedLanguages handles the request to get supported programming languages for SDK generation.
func (ctrl *SDKController) GetSupportedLanguages(c fiber.Ctx) error {
	ctrl.logger.Info("GetSupportedLanguages endpoint hit")

	// Define supported languages for SDK generation
	supportedLanguages := []map[string]interface{}{
		{
			"id":          "python",
			"name":        "Python",
			"description": "Generate Python SDK with requests library",
			"extension":   ".py",
		},
		{
			"id":          "javascript",
			"name":        "JavaScript",
			"description": "Generate JavaScript SDK for Node.js",
			"extension":   ".js",
		},
		{
			"id":          "typescript",
			"name":        "TypeScript",
			"description": "Generate TypeScript SDK with type definitions",
			"extension":   ".ts",
		},
		{
			"id":          "java",
			"name":        "Java",
			"description": "Generate Java SDK with OkHttp client",
			"extension":   ".java",
		},
		{
			"id":          "csharp",
			"name":        "C#",
			"description": "Generate C# SDK for .NET",
			"extension":   ".cs",
		},
		{
			"id":          "go",
			"name":        "Go",
			"description": "Generate Go SDK with net/http",
			"extension":   ".go",
		},
		{
			"id":          "php",
			"name":        "PHP",
			"description": "Generate PHP SDK with Guzzle HTTP client",
			"extension":   ".php",
		},
		{
			"id":          "ruby",
			"name":        "Ruby",
			"description": "Generate Ruby SDK with Faraday",
			"extension":   ".rb",
		},
	}

	return utils.SuccessResponse(c, "Supported languages retrieved successfully", supportedLanguages)
}

// GetSDKHistory handles the request to get SDK generation history for the authenticated user.
func (ctrl *SDKController) GetSDKHistory(c fiber.Ctx) error {
	userIDStr, ok := middleware.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal Error", "User ID not found in context")
	}

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	ctrl.logger.Info("GetSDKHistory request",
		zap.String("userID", userIDStr),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("statusFilter", statusFilter),
		zap.String("typeFilter", typeFilter))

	sdks, total, err := ctrl.sdkService.GetSDKsByUserID(context.Background(), userIDStr, page, limit)
	if err != nil {
		ctrl.logger.Error("Failed to get SDK history for user", zap.Error(err), zap.String("userID", userIDStr))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve SDK history", err.Error())
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	pagination := models.Pagination{
		CurrentPage: page,
		Limit:       limit,
		TotalItems:  total,
		TotalPages:  totalPages,
	}

	response := models.PaginatedSDKsResponse{
		SDKs:       sdks,
		Pagination: pagination,
	}

	return utils.SuccessResponse(c, "SDK history retrieved successfully", response)
}

package controllers

import (
	"math"
	"strconv"
	"time"

	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var validateSDK = validator.New() // Validator instance for SDK controller

type SDKController struct {
	logger            *zap.Logger
	sdkService        *services.SDKService
	collectionService *services.CollectionService
}

func NewSDKController(sdkService *services.SDKService, collectionService *services.CollectionService, logger *zap.Logger) *SDKController {
	return &SDKController{
		sdkService:        sdkService,
		collectionService: collectionService,
		logger:            logger,
	}
}

// GenerateSDK handles the request to generate an SDK from a collection.
func (ctrl *SDKController) GenerateSDK(c fiber.Ctx) error {
	ctrl.logger.Info("GenerateSDK endpoint hit")

	// Extract user ID from JWT token
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		ctrl.logger.Warn("GenerateSDK: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	var req models.GenerationRequest
	if err := c.Bind().Body(&req); err != nil {
		ctrl.logger.Error("Failed to parse request body", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate request
	if err := validateSDK.Struct(req); err != nil {
		ctrl.logger.Error("Invalid generation request", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	ctrl.logger.Info("SDK generation request received",
		zap.Any("request", req),
		zap.String("userID", userIDStr))

	// Create a unique output directory for this SDK generation
	outputBaseDir := "/tmp/api2sdk_outputs" // This could be configurable
	if req.OutputDirectory != "" {
		outputBaseDir = req.OutputDirectory
	}

	// Call the collection service to generate SDK from collection
	sdkPath, err := ctrl.collectionService.GenerateSDKFromCollection(
		c.Context(),
		userIDStr,
		req.CollectionID,
		req.Language,
		outputBaseDir,
	)

	if err != nil {
		ctrl.logger.Error("Failed to generate SDK", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "SDK generation failed", err.Error())
	}

	// Return success response
	response := models.GenerationResponse{
		Message:     "SDK generation completed successfully",
		Status:      "completed",
		OutputPath:  sdkPath,
		GeneratedAt: time.Now(),
	}

	ctrl.logger.Info("SDK generation completed",
		zap.String("outputPath", sdkPath),
		zap.String("userID", userIDStr))

	return utils.SuccessResponse(c, "SDK generated successfully", response)
}

// GetSDKHistory retrieves the user's SDK generation history with pagination.
func (ctrl *SDKController) GetSDKHistory(c fiber.Ctx) error {
	ctrl.logger.Info("GetSDKHistory endpoint hit")

	// Extract user ID from JWT token
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		ctrl.logger.Warn("GetSDKHistory: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 { // Cap the limit to prevent abuse
		limit = 10
	}

	ctrl.logger.Info("Retrieving SDK history",
		zap.String("userID", userIDStr),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Call service to get SDK history
	sdks, totalCount, err := ctrl.sdkService.GetSDKHistory(c.Context(), userIDStr, page, limit)
	if err != nil {
		ctrl.logger.Error("Failed to retrieve SDK history", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve SDK history", err.Error())
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Build response
	response := models.SDKHistoryResponse{
		SDKs:       sdks,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	ctrl.logger.Info("SDK history retrieved successfully",
		zap.String("userID", userIDStr),
		zap.Int("count", len(sdks)),
		zap.Int64("totalCount", totalCount))

	return utils.SuccessResponse(c, "SDK history retrieved successfully", response)
}

// DeleteSDK handles the soft deletion of an SDK.
func (ctrl *SDKController) DeleteSDK(c fiber.Ctx) error {
	ctrl.logger.Info("DeleteSDK endpoint hit")

	// Extract user ID from JWT token
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		ctrl.logger.Warn("DeleteSDK: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	// Get SDK ID from URL params
	sdkID := c.Params("id")
	if sdkID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "SDK ID is required", "")
	}

	ctrl.logger.Info("Deleting SDK",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	// Call service to delete SDK
	err := ctrl.sdkService.DeleteSDK(c.Context(), sdkID, userIDStr)
	if err != nil {
		ctrl.logger.Error("Failed to delete SDK", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete SDK", err.Error())
	}

	// Build response
	response := models.DeleteSDKResponse{
		Message: "SDK deleted successfully",
		Success: true,
	}

	ctrl.logger.Info("SDK deleted successfully",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	return utils.SuccessResponse(c, "SDK deleted successfully", response)
}

// DownloadSDK handles SDK file downloads.
func (ctrl *SDKController) DownloadSDK(c fiber.Ctx) error {
	ctrl.logger.Info("DownloadSDK endpoint hit")

	// Extract user ID from JWT token
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		ctrl.logger.Warn("DownloadSDK: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	// Get SDK ID from URL params
	sdkID := c.Params("id")
	if sdkID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "SDK ID is required", "")
	}

	ctrl.logger.Info("Download request for SDK",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	// Get SDK record from database to verify ownership and get file path

	// Validate SDK ID format
	_, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		ctrl.logger.Error("Invalid SDK ID format", zap.String("sdkID", sdkID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid SDK ID format", err.Error())
	}

	// TODO: Add GetSDKByID method to SDKService that verifies user ownership
	// For now, return a proper response indicating the feature needs additional implementation
	ctrl.logger.Info("SDK download requested but requires additional service method implementation",
		zap.String("sdkID", sdkID),
		zap.String("userID", userIDStr))

	// Return a structured response indicating the feature is recognized but needs completion
	return utils.ErrorResponse(c, fiber.StatusNotImplemented,
		"SDK download functionality is available but requires additional implementation",
		"The GetSDKByID service method and file streaming logic need to be implemented")
}

// GetSupportedLanguages returns the list of supported languages for SDK generation.
func (ctrl *SDKController) GetSupportedLanguages(c fiber.Ctx) error {
	ctrl.logger.Info("Request for supported languages")
	return utils.SuccessResponse(c, "Supported languages retrieved successfully", models.SupportedLanguages)
}

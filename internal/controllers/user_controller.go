package controllers

import (
	"strconv" // For converting string ID from path to uint

	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// UserController handles user-specific requests
type UserController struct {
	userService services.UserService
	logger      *zap.Logger
}

// NewUserController creates a new UserController
func NewUserController(userService services.UserService, logger *zap.Logger) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

var validateUser = validator.New() // Validator instance for user controller

// GetMe handles GET /api/v1/users/me - retrieves current authenticated user's details
func (uc *UserController) GetMe(c fiber.Ctx) error {
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		uc.logger.Warn("GetMe: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	user, err := uc.userService.GetUserProfile(c.Context(), userIDStr)
	if err != nil {
		uc.logger.Error("GetMe: Failed to get user by ID", zap.String("userID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}
	return utils.SuccessResponse(c, "Current user details retrieved successfully", user)
}

// UpdateMe handles PUT /api/v1/users/me - updates current authenticated user's details
func (uc *UserController) UpdateMe(c fiber.Ctx) error {
	userIDStr, ok := middleware.GetUserID(c)
	if !ok || userIDStr == "" {
		uc.logger.Warn("UpdateMe: UserID not found or invalid in context")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User ID not found. Please log in.", "")
	}

	var req models.UpdateUserProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate req if it has tags
	if err := validateUser.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	// The service should handle partial updates and not update password here.
	updatedUser, err := uc.userService.UpdateUserProfile(c.Context(), userIDStr, &req)
	if err != nil {
		uc.logger.Error("UpdateMe: Failed to update user", zap.String("userID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user details", err.Error())
	}
	return utils.SuccessResponse(c, "User details updated successfully", updatedUser)
}

// --- Admin User Management Methods ---

// GetAllUsers handles GET /api/v1/admin/users - retrieves all users (admin only)
func (uc *UserController) GetAllUsers(c fiber.Ctx) error {
	// This route is protected by AdminRequired middleware
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	users, total, err := uc.userService.GetAllUsers(c.Context(), page, limit)
	if err != nil {
		uc.logger.Error("GetAllUsers: Failed to get all users", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve users", err.Error())
	}
	return utils.SuccessResponse(c, "All users retrieved successfully", fiber.Map{"users": users, "total": total})
}

// GetUserByID handles GET /api/v1/admin/users/:id - retrieves a specific user by ID (admin only)
func (uc *UserController) GetUserByID(c fiber.Ctx) error {
	userIDStr := c.Params("id")
	_, err := strconv.ParseUint(userIDStr, 10, 32) // Validate format
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	user, err := uc.userService.GetUserProfile(c.Context(), userIDStr)
	if err != nil {
		uc.logger.Error("GetUserByID (admin): Failed to get user", zap.String("targetUserID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}
	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

// UpdateUser handles PUT /api/v1/admin/users/:id - updates a specific user by ID (admin only)
func (uc *UserController) UpdateUser(c fiber.Ctx) error {
	userIDStr := c.Params("id")
	_, err := strconv.ParseUint(userIDStr, 10, 32) // Validate format
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	var req models.UpdateUserProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validate req
	if err := validateUser.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	updatedUser, err := uc.userService.UpdateUserProfile(c.Context(), userIDStr, &req) // Service handles update logic
	if err != nil {
		uc.logger.Error("UpdateUser (admin): Failed to update user", zap.String("targetUserID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())
	}
	return utils.SuccessResponse(c, "User updated successfully by admin", updatedUser)
}

// DeleteUser handles DELETE /api/v1/admin/users/:id - deletes a specific user by ID (admin only)
func (uc *UserController) DeleteUser(c fiber.Ctx) error {
	userIDStr := c.Params("id")
	_, err := strconv.ParseUint(userIDStr, 10, 32) // Validate format
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	err = uc.userService.DeleteUserAsAdmin(c.Context(), userIDStr)
	if err != nil {
		uc.logger.Error("DeleteUser (admin): Failed to delete user", zap.String("targetUserID", userIDStr), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}
	return utils.SuccessResponse(c, "User deleted successfully by admin", nil)
}

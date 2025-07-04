package controllers

import (
	"encoding/json"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AuthController handles authentication related requests
type AuthController struct {
	authService services.AuthService
	userService services.UserService
	logger      *zap.Logger
}

// NewAuthController creates a new AuthController
func NewAuthController(authService services.AuthService, userService services.UserService, logger *zap.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

// RegisterRequest defines the structure for registration requests
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"` // Optional: specify role, defaults to "user"
}

var validateAuth = validator.New()

// Register handles user registration
func (ac *AuthController) Register(c fiber.Ctx) error {
	var req RegisterRequest
	body := c.Body()
	if len(body) == 0 {
		ac.logger.Error("Request body is empty for Register")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", "Request body is empty")
	}
	if err := json.Unmarshal(body, &req); err != nil {
		ac.logger.Error("Invalid request body for Register", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := validateAuth.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	userRole := models.RoleUser // Default role
	if req.Role != "" {
		userRole = models.UserRole(req.Role)
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // Password will be hashed by the service
		Role:     userRole,
	}

	createdUser, err := ac.authService.Register(c.Context(), user)
	if err != nil {
		ac.logger.Error("User registration failed", zap.String("email", req.Email), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to register user", err.Error())
	}

	return utils.SuccessResponse(c, "User registered successfully", fiber.Map{
		"user": createdUser,
	})
}

// LoginRequest defines the structure for login requests
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login handles user login (no authentication required)
func (ac *AuthController) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		ac.logger.Error("Failed to parse login request", zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if err := validateAuth.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	ac.logger.Info("Login attempt", zap.String("email", req.Email))

	return utils.SuccessResponse(c, "Login successful", fiber.Map{
		"user": fiber.Map{
			"id":    "default-user",
			"email": req.Email,
			"name":  "User",
			"role":  "user",
		},
	})
}

// GetUserProfile returns a default user profile (no authentication required)
func (ac *AuthController) GetUserProfile(c fiber.Ctx) error {
	ac.logger.Info("Get user profile")

	return utils.SuccessResponse(c, "User profile retrieved successfully", fiber.Map{
		"id":    "default-user",
		"email": "user@example.com",
		"name":  "Default User",
		"role":  "user",
	})
}

// Logout handles user logout (no-op since no authentication)
func (ac *AuthController) Logout(c fiber.Ctx) error {
	ac.logger.Info("Logout")
	return utils.SuccessResponse(c, "Logout successful", nil)
}

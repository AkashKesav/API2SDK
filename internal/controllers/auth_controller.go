package controllers

import (
	"strings"

	"github.com/AkashKesav/API2SDK/internal/middleware"
	"github.com/AkashKesav/API2SDK/internal/models" // Added for models.User
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

// LoginRequest defines the structure for login requests
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"` // Changed from Name to Email
	Password string `json:"password" validate:"required"`
}

// RegisterRequest defines the structure for registration requests
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"` // Optional: specify role, defaults to "user"
}

var validateAuth = validator.New() // Renamed to avoid conflict if other controllers use validate

// Login handles user login
func (ac *AuthController) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := validateAuth.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	// Use AuthService for login
	accessToken, refreshToken, user, err := ac.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		ac.logger.Error("Login failed", zap.String("email", req.Email), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials or login failed", err.Error())
	}

	return utils.SuccessResponse(c, "Login successful", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": fiber.Map{
			"id":    user.ID.Hex(),
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Register handles user registration
func (ac *AuthController) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := validateAuth.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	userRole := models.RoleUser // Default role
	if req.Role != "" {
		userRole = models.UserRole(req.Role) // Validate if req.Role is a valid role
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // Password will be hashed by the service
		Role:     userRole,
	}

	createdUser, accessToken, refreshToken, err := ac.authService.Register(c.Context(), user)
	if err != nil {
		ac.logger.Error("User registration failed", zap.String("email", req.Email), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to register user", err.Error())
	}

	return utils.SuccessResponse(c, "User registered successfully", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          createdUser, // createdUser already has password cleared by service
	})
}

// Logout handles user logout
// For stateless JWT, logout is typically handled client-side by deleting the token.
// Server-side logout might involve token blacklisting if implemented.
func (ac *AuthController) Logout(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		// This case should ideally not happen if JWTMiddleware is applied correctly before this handler
		ac.logger.Warn("Logout attempt by unauthenticated user or missing userID in context")
		// Still, allow logout to proceed client-side, but log it.
	} else {
		ac.logger.Info("Logout attempt", zap.String("userID", userID))
	}
	// Server-side token blacklisting would go here if implemented.
	return utils.SuccessResponse(c, "Logged out successfully. Please clear your token client-side.", nil)
}

// GetUserProfile handles retrieving current user info based on JWT
func (ac *AuthController) GetUserProfile(c fiber.Ctx) error {
	user, exists := middleware.GetUser(c)
	if !exists {
		ac.logger.Warn("GetUserProfile: User context not found")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User not found in context", "authentication_required")
	}

	// Fetch full user details from userService
	fullUser, err := ac.userService.GetUserProfile(c.Context(), user.ID)
	if err != nil {
		ac.logger.Error("GetUserProfile: Failed to get user by ID from service", zap.String("userID", user.ID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	return utils.SuccessResponse(c, "Current user profile", fullUser)
}

// RefreshTokenRequest defines the structure for token refresh requests
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken handles refreshing JWT tokens
func (ac *AuthController) RefreshToken(c fiber.Ctx) error {
	user, exists := middleware.GetUser(c)
	if !exists {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized: User not found in context", "authentication_required")
	}

	// Get refresh token from Authorization header (handled by RefreshTokenMiddleware)
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Refresh token required", "missing_refresh_token")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid authorization header format", "invalid_format")
	}

	refreshToken := parts[1]

	// Generate new tokens
	newAccessToken, newRefreshToken, err := ac.authService.RefreshTokens(c.Context(), refreshToken)
	if err != nil {
		ac.logger.Error("Token refresh failed", zap.String("userID", user.ID), zap.Error(err))
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Failed to refresh token", err.Error())
	}

	return utils.SuccessResponse(c, "Tokens refreshed successfully", fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

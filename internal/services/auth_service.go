package services

import (
	"context" // Added for repository calls
	"errors"  // Added for error handling
	"fmt"     // Added for error formatting

	"github.com/AkashKesav/API2SDK/configs"             // Added for config
	"github.com/AkashKesav/API2SDK/internal/middleware" // Added for GenerateJWT
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive" // Added for ObjectID conversion
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt" // For password hashing
)

// AuthService defines the interface for authentication related services
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, *models.User, error) // Return access token, refresh token, and user
	Register(ctx context.Context, user models.User) (*models.User, string, string, error)    // Return user, access token, and refresh token
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)          // Return new access token and refresh token
	ValidateRefreshToken(ctx context.Context, refreshToken string) (*models.User, error)     // Validate refresh token and return user
}

type authService struct {
	userRepository repositories.UserRepository
	logger         *zap.Logger
	config         *configs.Config // Added field for config
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(userRepo repositories.UserRepository, logger *zap.Logger, config *configs.Config) AuthService {
	return &authService{
		userRepository: userRepo,
		logger:         logger,
		config:         config, // Store config
	}
}

// Login handles user login, verifies credentials, and generates JWT tokens.
func (s *authService) Login(ctx context.Context, email, password string) (string, string, *models.User, error) {
	s.logger.Info("Attempting login", zap.String("email", email))

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Error finding user by email", zap.String("email", email), zap.Error(err))
		return "", "", nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		s.logger.Warn("User not found during login", zap.String("email", email))
		return "", "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.logger.Warn("Password mismatch", zap.String("email", email), zap.Error(err))
		return "", "", nil, errors.New("invalid email or password")
	}

	// Generate both access and refresh tokens
	accessToken, err := middleware.GenerateAccessToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.String("email", email), zap.Error(err))
		return "", "", nil, fmt.Errorf("could not generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.String("email", email), zap.Error(err))
		return "", "", nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	s.logger.Info("Login successful, tokens generated", zap.String("email", email))
	user.Password = "" // Clear password before returning
	return accessToken, refreshToken, user, nil
}

// Register handles user registration, creates the user, and generates JWT tokens.
func (s *authService) Register(ctx context.Context, user models.User) (*models.User, string, string, error) {
	s.logger.Info("Attempting registration", zap.String("email", user.Email))

	existingUser, err := s.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		s.logger.Error("Error checking for existing user during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, "", "", fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		s.logger.Warn("User already exists during registration attempt", zap.String("email", user.Email))
		return nil, "", "", errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, "", "", fmt.Errorf("could not hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Ensure default role if not provided
	if user.Role == "" {
		user.Role = models.RoleUser
	}

	userID, err := s.userRepository.Create(ctx, &user)
	if err != nil {
		s.logger.Error("Failed to create user during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, "", "", fmt.Errorf("could not create user: %w", err)
	}
	user.ID = userID // Set the returned ID

	// Generate both access and refresh tokens for the newly registered user
	accessToken, err := middleware.GenerateAccessToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate access token after registration", zap.String("email", user.Email), zap.Error(err))
		return &user, "", "", fmt.Errorf("user created, but could not generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate refresh token after registration", zap.String("email", user.Email), zap.Error(err))
		return &user, "", "", fmt.Errorf("user created, but could not generate refresh token: %w", err)
	}

	s.logger.Info("User registered successfully, tokens generated", zap.String("email", user.Email))
	user.Password = "" // Clear password before returning
	return &user, accessToken, refreshToken, nil
}

// RefreshTokens handles refreshing JWT tokens using a valid refresh token
func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	// Validate the refresh token
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return "", "", errors.New("invalid refresh token")
	}

	// Ensure it's a refresh token
	if claims.Type != "refresh" {
		s.logger.Warn("Wrong token type for refresh", zap.String("type", claims.Type))
		return "", "", errors.New("invalid token type")
	}

	// Convert string ID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		s.logger.Error("Invalid user ID in token", zap.String("userID", claims.UserID), zap.Error(err))
		return "", "", errors.New("invalid user ID")
	}

	// Verify user still exists
	user, err := s.userRepository.FindByID(ctx, userObjectID)
	if err != nil || user == nil {
		s.logger.Error("User not found during token refresh", zap.String("userID", claims.UserID), zap.Error(err))
		return "", "", errors.New("user not found")
	}

	// Generate new tokens
	newAccessToken, err := middleware.GenerateAccessToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate new access token", zap.String("userID", claims.UserID), zap.Error(err))
		return "", "", fmt.Errorf("could not generate new access token: %w", err)
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(user.ID.Hex(), user.Email, string(user.Role), s.config)
	if err != nil {
		s.logger.Error("Failed to generate new refresh token", zap.String("userID", claims.UserID), zap.Error(err))
		return "", "", fmt.Errorf("could not generate new refresh token: %w", err)
	}

	s.logger.Info("Tokens refreshed successfully", zap.String("userID", claims.UserID))
	return newAccessToken, newRefreshToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user
func (s *authService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*models.User, error) {
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Convert string ID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepository.FindByID(ctx, userObjectID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	user.Password = "" // Clear password
	return user, nil
}

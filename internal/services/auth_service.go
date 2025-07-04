package services

import (
	"context" // Added for repository calls
	"errors"  // Added for error handling
	"fmt"     // Added for error formatting

	"github.com/AkashKesav/API2SDK/configs" // Added for config
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories" // Added for ObjectID conversion
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt" // For password hashing
)

// AuthService defines the interface for authentication related services
type AuthService interface {
	Register(ctx context.Context, user models.User) (*models.User, error) // Return user
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

// Register handles user registration, creates the user.
func (s *authService) Register(ctx context.Context, user models.User) (*models.User, error) {
	s.logger.Info("Attempting registration", zap.String("email", user.Email))

	existingUser, err := s.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		s.logger.Error("Error checking for existing user during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		s.logger.Warn("User already exists during registration attempt", zap.String("email", user.Email))
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, fmt.Errorf("could not hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Ensure default role if not provided
	if user.Role == "" {
		user.Role = models.RoleUser
	}

	createdUserID, err := s.userRepository.Create(ctx, &user)
	if err != nil {
		s.logger.Error("Failed to create user during registration", zap.String("email", user.Email), zap.Error(err))
		return nil, fmt.Errorf("could not create user: %w", err)
	}
	user.ID = createdUserID // Set the returned ID

	s.logger.Info("User registered successfully", zap.String("email", user.Email))
	user.Password = "" // Clear password before returning
	return &user, nil
}

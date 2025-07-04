package services

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"go.uber.org/zap" // Added for logging
)

// UserService defines the interface for user-related operations.
type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string) (*models.User, error) // Changed from username to name
	GetUserProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserProfileRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
	DeleteUserAccount(ctx context.Context, userID string) error
	GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) // Changed to []*models.User
	UpdateUserRole(ctx context.Context, userID string, newRole models.UserRole) error
	DeleteUserAsAdmin(ctx context.Context, userID string) error
	GetTotalUsersCount(ctx context.Context) (int64, error)  // Added for admin stats
	GetRecentUsersCount(ctx context.Context) (int64, error) // Added for recent users count
}

// userServiceImpl implements the UserService interface.
type userServiceImpl struct {
	userRepo                repositories.UserRepository
	platformSettingsService PlatformSettingsService // Added PlatformSettingsService
	logger                  *zap.Logger             // Added logger
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo repositories.UserRepository, psService PlatformSettingsService, logger *zap.Logger) UserService { // Added PlatformSettingsService and logger
	return &userServiceImpl{
		userRepo:                userRepo,
		platformSettingsService: psService, // Store injected service
		logger:                  logger,    // Store injected logger
	}
}

// validatePassword checks if password meets complexity requirements
func (s *userServiceImpl) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters long")
	}

	// Check for at least one uppercase letter
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one special character
	if matched, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password); !matched {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// RegisterUser handles new user registration.
func (s *userServiceImpl) RegisterUser(ctx context.Context, name, email, password string) (*models.User, error) { // Changed from username to name
	existingUserByName, err := s.userRepo.FindByName(ctx, name) // Changed from FindByUsername to FindByName
	if err != nil {
		return nil, fmt.Errorf("error checking name: %w", err)
	}
	if existingUserByName != nil {
		return nil, fmt.Errorf("name already taken")
	}

	existingUserByEmail, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existingUserByEmail != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate password complexity
	if err := s.validatePassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:     name, // Changed from Username to Name
		Email:    email,
		Password: string(hashedPassword),
	}

	userID, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.ID = userID

	user.Password = ""
	return user, nil
}

// GetUserProfile retrieves a user's profile by their ID.
func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, objID) // Assumes FindByID exists in UserRepository
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	user.Password = "" // Ensure password is not returned
	return user, nil
}

// UpdateUserProfile updates a user's profile information.
func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserProfileRequest) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Fetch the existing user
	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if provided in the request
	// Example: (ensure models.UpdateUserProfileRequest has these fields)
	if req.Email != "" && req.Email != user.Email {
		// Optional: Add email validation and check for uniqueness if email is changeable
		existingUserByEmail, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("error checking email: %w", err)
		}
		if existingUserByEmail != nil && existingUserByEmail.ID != objID {
			return nil, fmt.Errorf("email already registered by another user")
		}
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	// Add other updatable fields from models.UpdateUserProfileRequest

	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil { // Assumes Update method exists and takes *models.User
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.Password = "" // Ensure password is not returned
	return user, nil
}

// ChangePassword allows a user to change their password.
func (s *userServiceImpl) ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("incorrect old password")
	}

	// Validate new password complexity
	if err := s.validatePassword(newPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.Password = string(hashedNewPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user) // Reusing Update or a specific UpdatePassword method
}

// DeleteUserAccount allows a user to delete their own account.
func (s *userServiceImpl) DeleteUserAccount(ctx context.Context, userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Optional: Add checks, e.g., ensure user is not an admin of critical resources, etc.

	return s.userRepo.Delete(ctx, objID) // Assumes Delete method exists in UserRepository
}

// GetAllUsers retrieves a paginated list of all users.
func (s *userServiceImpl) GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) { // Changed to []*models.User
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // Default limit
	}
	offset := (page - 1) * limit
	// Assuming FindAll does not require context or it's handled within the repository
	return s.userRepo.FindAll(ctx, offset, limit)
}

// UpdateUserRole updates the role of a specific user.
func (s *userServiceImpl) UpdateUserRole(ctx context.Context, userID string, newRole models.UserRole) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.Role = newRole
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// DeleteUserAsAdmin allows an admin to delete any user account.
func (s *userServiceImpl) DeleteUserAsAdmin(ctx context.Context, userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	return s.userRepo.Delete(ctx, objID) // Assumes Delete method exists
}

// GetTotalUsersCount retrieves the total number of registered users.
func (s *userServiceImpl) GetTotalUsersCount(ctx context.Context) (int64, error) {
	count, err := s.userRepo.CountAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get total users count from repository", zap.Error(err))
		return 0, fmt.Errorf("failed to get total users count: %w", err)
	}
	return count, nil
}

// GetRecentUsersCount returns the number of users registered in the last 7 days
func (s *userServiceImpl) GetRecentUsersCount(ctx context.Context) (int64, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	return s.userRepo.CountCreatedAfter(ctx, sevenDaysAgo)
}

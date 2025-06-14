package services

import (
	"context"
	"fmt"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap" // Added for logging
)

// DefaultJWTSecretKey is a fallback if no key is in settings.
// It's crucial that this is changed or that a key is always set in production.
var DefaultJWTSecretKey = []byte("your-very-secret-key-for-api2sdk-project-change-this-in-prod")

// Claims defines the JWT claims structure.
type Claims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"` // Changed from Username to Name
	jwt.RegisteredClaims
}

// UserService defines the interface for user-related operations.
type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string) (*models.User, error) // Changed from username to name
	LoginUser(ctx context.Context, name, password string) (*models.User, string, error)   // Changed from username to name
	GetUserProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateUserProfileRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error
	DeleteUserAccount(ctx context.Context, userID string) error
	GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, int64, error) // Changed to []*models.User
	UpdateUserRole(ctx context.Context, userID string, newRole models.UserRole) error
	DeleteUserAsAdmin(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshTokenString string) (string, error) // Added for token refresh
	GetTotalUsersCount(ctx context.Context) (int64, error)                       // Added for admin stats
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

func (s *userServiceImpl) getJWTSecretKey(ctx context.Context) []byte {
	settings, err := s.platformSettingsService.GetPlatformSettingsStruct(ctx)
	if err != nil {
		s.logger.Warn("Failed to get platform settings for JWT secret, using default key", zap.Error(err))
		return DefaultJWTSecretKey
	}
	if settings.JWTSecretKey == "" {
		s.logger.Warn("JWT secret key is not set in platform settings, using default key")
		return DefaultJWTSecretKey
	}
	return []byte(settings.JWTSecretKey)
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

// LoginUser handles user login and JWT generation.
func (s *userServiceImpl) LoginUser(ctx context.Context, name, password string) (*models.User, string, error) { // Changed from username to name
	user, err := s.userRepo.FindByName(ctx, name) // Changed from FindByUsername to FindByName
	if err != nil {
		return nil, "", fmt.Errorf("error fetching user: %w", err)
	}
	if user == nil {
		return nil, "", fmt.Errorf("invalid name or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", fmt.Errorf("invalid name or password")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID.Hex(),
		Name:   user.Name, // Changed from Username to Name
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "api2sdk",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := s.getJWTSecretKey(ctx) // Get JWT key from settings
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.Password = "" // Clear password before returning user object
	return user, tokenString, nil
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

	if len(newPassword) < 6 { // Example: Add password policy validation
		return fmt.Errorf("new password is too short")
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

// RefreshToken validates a refresh token and issues a new access token.
func (s *userServiceImpl) RefreshToken(ctx context.Context, refreshTokenString string) (string, error) {
	claims := &Claims{}
	jwtKey := s.getJWTSecretKey(ctx) // Get JWT key from settings

	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil // Use the key from settings
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", fmt.Errorf("invalid token signature")
		}
		// Check for expired token, though token.Valid should also catch this.
		// err can be jwt.ErrTokenExpired or other validation errors.
		return "", fmt.Errorf("could not parse refresh token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("refresh token is invalid or expired")
	}

	// Check if the user still exists and is active, etc. (optional)
	// _, err = s.userRepo.FindByID(ctx, claims.UserID) // Assuming claims.UserID is primitive.ObjectID
	// if err != nil {
	// 	return "", fmt.Errorf("user not found or inactive")
	// }

	// Create new access token
	newExpirationTime := time.Now().Add(1 * time.Hour) // Shorter lifespan for access tokens
	newClaims := &Claims{
		UserID: claims.UserID,
		Name:   claims.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(newExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "api2sdk",
		},
	}

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newAccessTokenString, err := newAccessToken.SignedString(jwtKey) // Use the same key
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessTokenString, nil
}

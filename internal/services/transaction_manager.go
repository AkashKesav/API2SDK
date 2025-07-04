package services

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// TransactionManager handles database transactions
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithSession(ctx context.Context, fn func(sessionCtx context.Context) error) error
}

// MongoTransactionManager implements TransactionManager for MongoDB
type MongoTransactionManager struct {
	client *mongo.Client
	logger *zap.Logger
}

// NewMongoTransactionManager creates a new MongoDB transaction manager
func NewMongoTransactionManager(client *mongo.Client, logger *zap.Logger) TransactionManager {
	return &MongoTransactionManager{
		client: client,
		logger: logger,
	}
}

// WithTransaction executes a function within a MongoDB transaction
func (tm *MongoTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		tm.logger.Error("Failed to start MongoDB session", zap.Error(err))
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// Define the transaction function
	transactionFn := func(sessionCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessionCtx)
	}

	// Execute the transaction with retry logic
	_, err = session.WithTransaction(ctx, transactionFn)
	if err != nil {
		tm.logger.Error("Transaction failed", zap.Error(err))
		return fmt.Errorf("transaction failed: %w", err)
	}

	tm.logger.Debug("Transaction completed successfully")
	return nil
}

// WithSession executes a function within a MongoDB session (without transaction)
func (tm *MongoTransactionManager) WithSession(ctx context.Context, fn func(sessionCtx context.Context) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		tm.logger.Error("Failed to start MongoDB session", zap.Error(err))
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	sessionCtx := mongo.NewSessionContext(ctx, session)
	return fn(sessionCtx)
}

// TransactionContext provides transaction-aware context
type TransactionContext struct {
	context.Context
	inTransaction bool
	session       mongo.Session
}

// NewTransactionContext creates a new transaction context
func NewTransactionContext(ctx context.Context, session mongo.Session) *TransactionContext {
	return &TransactionContext{
		Context:       ctx,
		inTransaction: true,
		session:       session,
	}
}

// IsInTransaction returns true if the context is within a transaction
func (tc *TransactionContext) IsInTransaction() bool {
	return tc.inTransaction
}

// GetSession returns the MongoDB session
func (tc *TransactionContext) GetSession() mongo.Session {
	return tc.session
}

// TransactionAware interface for services that need transaction support
type TransactionAware interface {
	SetTransactionManager(tm TransactionManager)
}

// BaseService provides common service functionality with transaction support
type BaseService struct {
	txManager TransactionManager
	logger    *zap.Logger
}

// NewBaseService creates a new base service
func NewBaseService(txManager TransactionManager, logger *zap.Logger) *BaseService {
	return &BaseService{
		txManager: txManager,
		logger:    logger,
	}
}

// SetTransactionManager sets the transaction manager
func (s *BaseService) SetTransactionManager(tm TransactionManager) {
	s.txManager = tm
}

// WithTransaction executes a function within a transaction
func (s *BaseService) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if s.txManager == nil {
		s.logger.Warn("No transaction manager available, executing without transaction")
		return fn(ctx)
	}
	return s.txManager.WithTransaction(ctx, fn)
}

// Example usage in a service
type UserServiceWithTransaction struct {
	*BaseService
	userRepo    UserRepository
	profileRepo ProfileRepository
}

func NewUserServiceWithTransaction(
	txManager TransactionManager,
	userRepo UserRepository,
	profileRepo ProfileRepository,
	logger *zap.Logger,
) *UserServiceWithTransaction {
	return &UserServiceWithTransaction{
		BaseService: NewBaseService(txManager, logger),
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

// CreateUserWithProfile creates a user and their profile in a single transaction
func (s *UserServiceWithTransaction) CreateUserWithProfile(ctx context.Context, userReq CreateUserRequest, profileReq CreateProfileRequest) error {
	return s.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create user
		user := &User{
			Email:    userReq.Email,
			Name:     userReq.Name,
			Password: userReq.Password,
		}

		createdUser, err := s.userRepo.Create(txCtx, user)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Create profile
		profile := &UserProfile{
			UserID:      createdUser.ID,
			FirstName:   profileReq.FirstName,
			LastName:    profileReq.LastName,
			PhoneNumber: profileReq.PhoneNumber,
		}

		_, err = s.profileRepo.Create(txCtx, profile)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		s.logger.Info("User and profile created successfully",
			zap.String("userID", createdUser.ID),
			zap.String("email", createdUser.Email))

		return nil
	})
}

// Placeholder types for the example
type User struct {
	ID       string
	Email    string
	Name     string
	Password string
}

type UserProfile struct {
	UserID      string
	FirstName   string
	LastName    string
	PhoneNumber string
}

type CreateUserRequest struct {
	Email    string
	Name     string
	Password string
}

type CreateProfileRequest struct {
	FirstName   string
	LastName    string
	PhoneNumber string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) (*UserProfile, error)
}

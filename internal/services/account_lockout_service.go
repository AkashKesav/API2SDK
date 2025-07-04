package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AccountLockoutService handles account lockout functionality
type AccountLockoutService interface {
	RecordFailedAttempt(ctx context.Context, identifier string) error
	IsAccountLocked(ctx context.Context, identifier string) (bool, time.Duration, error)
	ResetFailedAttempts(ctx context.Context, identifier string) error
	GetFailedAttempts(ctx context.Context, identifier string) (int, error)
	GetLockoutInfo(ctx context.Context, identifier string) (*LockoutInfo, error)
}

type accountLockoutServiceImpl struct {
	redisClient      *redis.Client
	logger           *zap.Logger
	maxAttempts      int
	lockoutDuration  time.Duration
	attemptWindow    time.Duration
	keyPrefix        string
	lockoutKeyPrefix string
}

// AccountLockoutConfig defines the configuration for account lockout
type AccountLockoutConfig struct {
	MaxAttempts     int           // Maximum failed attempts before lockout
	LockoutDuration time.Duration // How long to lock the account
	AttemptWindow   time.Duration // Time window for counting attempts
}

// DefaultAccountLockoutConfig provides sensible defaults
var DefaultAccountLockoutConfig = AccountLockoutConfig{
	MaxAttempts:     5,
	LockoutDuration: 30 * time.Minute,
	AttemptWindow:   15 * time.Minute,
}

// NewAccountLockoutService creates a new account lockout service
func NewAccountLockoutService(redisClient *redis.Client, logger *zap.Logger, config AccountLockoutConfig) AccountLockoutService {
	return &accountLockoutServiceImpl{
		redisClient:      redisClient,
		logger:           logger,
		maxAttempts:      config.MaxAttempts,
		lockoutDuration:  config.LockoutDuration,
		attemptWindow:    config.AttemptWindow,
		keyPrefix:        "failed_attempts:",
		lockoutKeyPrefix: "account_locked:",
	}
}

// RecordFailedAttempt records a failed login attempt
func (s *accountLockoutServiceImpl) RecordFailedAttempt(ctx context.Context, identifier string) error {
	attemptKey := s.keyPrefix + identifier
	lockoutKey := s.lockoutKeyPrefix + identifier

	// Use pipeline for atomic operations
	pipe := s.redisClient.Pipeline()

	// Increment failed attempts counter
	incrCmd := pipe.Incr(ctx, attemptKey)

	// Set expiration for attempts counter
	pipe.Expire(ctx, attemptKey, s.attemptWindow)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("Failed to record failed attempt",
			zap.String("identifier", identifier),
			zap.Error(err),
		)
		return fmt.Errorf("failed to record failed attempt: %w", err)
	}

	attempts := int(incrCmd.Val())

	// Check if we should lock the account
	if attempts >= s.maxAttempts {
		// Lock the account
		err = s.redisClient.Set(ctx, lockoutKey, "locked", s.lockoutDuration).Err()
		if err != nil {
			s.logger.Error("Failed to lock account",
				zap.String("identifier", identifier),
				zap.Error(err),
			)
			return fmt.Errorf("failed to lock account: %w", err)
		}

		s.logger.Warn("Account locked due to too many failed attempts",
			zap.String("identifier", identifier),
			zap.Int("attempts", attempts),
			zap.Duration("lockout_duration", s.lockoutDuration),
		)
	}

	s.logger.Info("Failed attempt recorded",
		zap.String("identifier", identifier),
		zap.Int("attempts", attempts),
		zap.Int("max_attempts", s.maxAttempts),
	)

	return nil
}

// IsAccountLocked checks if an account is currently locked
func (s *accountLockoutServiceImpl) IsAccountLocked(ctx context.Context, identifier string) (bool, time.Duration, error) {
	lockoutKey := s.lockoutKeyPrefix + identifier

	// Check if lockout key exists
	result := s.redisClient.Exists(ctx, lockoutKey)
	if result.Err() != nil {
		s.logger.Error("Failed to check account lockout status",
			zap.String("identifier", identifier),
			zap.Error(result.Err()),
		)
		return false, 0, fmt.Errorf("failed to check lockout status: %w", result.Err())
	}

	if result.Val() == 0 {
		// Account is not locked
		return false, 0, nil
	}

	// Get remaining TTL
	ttl := s.redisClient.TTL(ctx, lockoutKey)
	if ttl.Err() != nil {
		s.logger.Error("Failed to get lockout TTL",
			zap.String("identifier", identifier),
			zap.Error(ttl.Err()),
		)
		return true, 0, fmt.Errorf("failed to get lockout TTL: %w", ttl.Err())
	}

	return true, ttl.Val(), nil
}

// ResetFailedAttempts clears failed attempts for an account (called on successful login)
func (s *accountLockoutServiceImpl) ResetFailedAttempts(ctx context.Context, identifier string) error {
	attemptKey := s.keyPrefix + identifier
	lockoutKey := s.lockoutKeyPrefix + identifier

	// Use pipeline to delete both keys
	pipe := s.redisClient.Pipeline()
	pipe.Del(ctx, attemptKey)
	pipe.Del(ctx, lockoutKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("Failed to reset failed attempts",
			zap.String("identifier", identifier),
			zap.Error(err),
		)
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}

	s.logger.Info("Failed attempts reset",
		zap.String("identifier", identifier),
	)

	return nil
}

// GetFailedAttempts returns the current number of failed attempts
func (s *accountLockoutServiceImpl) GetFailedAttempts(ctx context.Context, identifier string) (int, error) {
	attemptKey := s.keyPrefix + identifier

	result := s.redisClient.Get(ctx, attemptKey)
	if result.Err() == redis.Nil {
		// No failed attempts recorded
		return 0, nil
	}

	if result.Err() != nil {
		s.logger.Error("Failed to get failed attempts count",
			zap.String("identifier", identifier),
			zap.Error(result.Err()),
		)
		return 0, fmt.Errorf("failed to get failed attempts: %w", result.Err())
	}

	attempts, err := strconv.Atoi(result.Val())
	if err != nil {
		s.logger.Error("Failed to parse failed attempts count",
			zap.String("identifier", identifier),
			zap.String("value", result.Val()),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to parse failed attempts: %w", err)
	}

	return attempts, nil
}

// GetLockoutInfo returns comprehensive lockout information
type LockoutInfo struct {
	IsLocked       bool          `json:"is_locked"`
	FailedAttempts int           `json:"failed_attempts"`
	MaxAttempts    int           `json:"max_attempts"`
	RemainingTime  time.Duration `json:"remaining_time"`
	AttemptsLeft   int           `json:"attempts_left"`
}

// GetLockoutInfo returns comprehensive lockout information for an account
func (s *accountLockoutServiceImpl) GetLockoutInfo(ctx context.Context, identifier string) (*LockoutInfo, error) {
	info := &LockoutInfo{
		MaxAttempts: s.maxAttempts,
	}

	// Check if account is locked
	isLocked, remainingTime, err := s.IsAccountLocked(ctx, identifier)
	if err != nil {
		return nil, err
	}

	info.IsLocked = isLocked
	info.RemainingTime = remainingTime

	// Get failed attempts count
	failedAttempts, err := s.GetFailedAttempts(ctx, identifier)
	if err != nil {
		return nil, err
	}

	info.FailedAttempts = failedAttempts
	info.AttemptsLeft = s.maxAttempts - failedAttempts
	if info.AttemptsLeft < 0 {
		info.AttemptsLeft = 0
	}

	return info, nil
}

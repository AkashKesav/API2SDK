package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheService provides caching functionality
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	FlushPattern(ctx context.Context, pattern string) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

type cacheServiceImpl struct {
	redisClient *redis.Client
	logger      *zap.Logger
	keyPrefix   string
}

// CacheConfig defines the configuration for cache service
type CacheConfig struct {
	KeyPrefix string
}

// DefaultCacheConfig provides sensible defaults
var DefaultCacheConfig = CacheConfig{
	KeyPrefix: "cache:",
}

// NewCacheService creates a new cache service
func NewCacheService(redisClient *redis.Client, logger *zap.Logger, config CacheConfig) CacheService {
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultCacheConfig.KeyPrefix
	}
	
	return &cacheServiceImpl{
		redisClient: redisClient,
		logger:      logger,
		keyPrefix:   config.KeyPrefix,
	}
}

// Set stores a value in cache with expiration
func (s *cacheServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := s.keyPrefix + key
	
	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		s.logger.Error("Failed to marshal cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	// Store in Redis
	err = s.redisClient.Set(ctx, fullKey, data, expiration).Err()
	if err != nil {
		s.logger.Error("Failed to set cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set cache value: %w", err)
	}
	
	s.logger.Debug("Cache value set",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
	)
	
	return nil
}

// Get retrieves a value from cache
func (s *cacheServiceImpl) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.keyPrefix + key
	
	// Get from Redis
	result := s.redisClient.Get(ctx, fullKey)
	if result.Err() == redis.Nil {
		return fmt.Errorf("cache miss for key: %s", key)
	}
	
	if result.Err() != nil {
		s.logger.Error("Failed to get cache value",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return fmt.Errorf("failed to get cache value: %w", result.Err())
	}
	
	// Deserialize JSON
	err := json.Unmarshal([]byte(result.Val()), dest)
	if err != nil {
		s.logger.Error("Failed to unmarshal cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	
	s.logger.Debug("Cache hit",
		zap.String("key", key),
	)
	
	return nil
}

// Delete removes keys from cache
func (s *cacheServiceImpl) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	
	// Add prefix to all keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.keyPrefix + key
	}
	
	// Delete from Redis
	result := s.redisClient.Del(ctx, fullKeys...)
	if result.Err() != nil {
		s.logger.Error("Failed to delete cache keys",
			zap.Strings("keys", keys),
			zap.Error(result.Err()),
		)
		return fmt.Errorf("failed to delete cache keys: %w", result.Err())
	}
	
	s.logger.Debug("Cache keys deleted",
		zap.Strings("keys", keys),
		zap.Int64("deleted_count", result.Val()),
	)
	
	return nil
}

// Exists checks if a key exists in cache
func (s *cacheServiceImpl) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.keyPrefix + key
	
	result := s.redisClient.Exists(ctx, fullKey)
	if result.Err() != nil {
		s.logger.Error("Failed to check cache key existence",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return false, fmt.Errorf("failed to check key existence: %w", result.Err())
	}
	
	return result.Val() > 0, nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (s *cacheServiceImpl) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	fullKey := s.keyPrefix + key
	
	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		s.logger.Error("Failed to marshal cache value for SetNX",
			zap.String("key", key),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}
	
	// Set only if not exists
	result := s.redisClient.SetNX(ctx, fullKey, data, expiration)
	if result.Err() != nil {
		s.logger.Error("Failed to SetNX cache value",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return false, fmt.Errorf("failed to SetNX cache value: %w", result.Err())
	}
	
	success := result.Val()
	s.logger.Debug("SetNX operation completed",
		zap.String("key", key),
		zap.Bool("success", success),
	)
	
	return success, nil
}

// Increment atomically increments a numeric value
func (s *cacheServiceImpl) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := s.keyPrefix + key
	
	result := s.redisClient.Incr(ctx, fullKey)
	if result.Err() != nil {
		s.logger.Error("Failed to increment cache value",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return 0, fmt.Errorf("failed to increment value: %w", result.Err())
	}
	
	s.logger.Debug("Cache value incremented",
		zap.String("key", key),
		zap.Int64("new_value", result.Val()),
	)
	
	return result.Val(), nil
}

// Expire sets expiration time for a key
func (s *cacheServiceImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := s.keyPrefix + key
	
	result := s.redisClient.Expire(ctx, fullKey, expiration)
	if result.Err() != nil {
		s.logger.Error("Failed to set cache key expiration",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return fmt.Errorf("failed to set expiration: %w", result.Err())
	}
	
	s.logger.Debug("Cache key expiration set",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
	)
	
	return nil
}

// FlushPattern deletes all keys matching a pattern
func (s *cacheServiceImpl) FlushPattern(ctx context.Context, pattern string) error {
	fullPattern := s.keyPrefix + pattern
	
	// Scan for keys matching pattern
	iter := s.redisClient.Scan(ctx, 0, fullPattern, 100).Iterator()
	var keys []string
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		s.logger.Error("Failed to scan cache keys",
			zap.String("pattern", pattern),
			zap.Error(err),
		)
		return fmt.Errorf("failed to scan keys: %w", err)
	}
	
	// Delete found keys
	if len(keys) > 0 {
		result := s.redisClient.Del(ctx, keys...)
		if result.Err() != nil {
			s.logger.Error("Failed to delete cache keys by pattern",
				zap.String("pattern", pattern),
				zap.Error(result.Err()),
			)
			return fmt.Errorf("failed to delete keys: %w", result.Err())
		}
		
		s.logger.Info("Cache keys deleted by pattern",
			zap.String("pattern", pattern),
			zap.Int("deleted_count", len(keys)),
		)
	}
	
	return nil
}

// GetTTL returns the remaining time to live for a key
func (s *cacheServiceImpl) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := s.keyPrefix + key
	
	result := s.redisClient.TTL(ctx, fullKey)
	if result.Err() != nil {
		s.logger.Error("Failed to get cache key TTL",
			zap.String("key", key),
			zap.Error(result.Err()),
		)
		return 0, fmt.Errorf("failed to get TTL: %w", result.Err())
	}
	
	return result.Val(), nil
}

// CacheWrapper provides a convenient way to cache function results
func (s *cacheServiceImpl) CacheWrapper(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	err := s.Get(ctx, key, &result)
	if err == nil {
		s.logger.Debug("Cache hit for wrapper",
			zap.String("key", key),
		)
		return result, nil
	}
	
	// Cache miss, execute function
	s.logger.Debug("Cache miss for wrapper, executing function",
		zap.String("key", key),
	)
	
	result, err = fn()
	if err != nil {
		return nil, err
	}
	
	// Store result in cache
	if cacheErr := s.Set(ctx, key, result, expiration); cacheErr != nil {
		s.logger.Warn("Failed to cache function result",
			zap.String("key", key),
			zap.Error(cacheErr),
		)
		// Don't return cache error, return the actual result
	}
	
	return result, nil
}
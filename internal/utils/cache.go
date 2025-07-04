package utils

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// Cache implements a simple in-memory cache with expiration
type Cache struct {
	items   map[string]CacheItem
	mutex   sync.RWMutex
	logger  *zap.Logger
	name    string
	metrics *CacheMetrics
}

// CacheMetrics tracks cache performance metrics
type CacheMetrics struct {
	Hits      int64
	Misses    int64
	Evictions int64
	mutex     sync.Mutex
}

// NewCache creates a new cache instance
func NewCache(name string, logger *zap.Logger) *Cache {
	cache := &Cache{
		items:   make(map[string]CacheItem),
		logger:  logger,
		name:    name,
		metrics: &CacheMetrics{},
	}

	// Start background cleanup
	go cache.startCleanupTask()

	return cache
}

// Set adds an item to the cache with the specified TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(ttl)
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}

	c.logger.Debug("Cache item set",
		zap.String("cache", c.name),
		zap.String("key", key),
		zap.Time("expiration", expiration))
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		c.metrics.recordMiss()
		return nil, false
	}

	// Check if the item has expired
	if time.Now().After(item.Expiration) {
		c.metrics.recordMiss()
		return nil, false
	}

	c.metrics.recordHit()
	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	c.logger.Debug("Cache item deleted", 
		zap.String("cache", c.name),
		zap.String("key", key))
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]CacheItem)
	c.logger.Info("Cache cleared", zap.String("cache", c.name))
}

// GetMetrics returns the current cache metrics
func (c *Cache) GetMetrics() (hits, misses, evictions int64) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()
	
	return c.metrics.Hits, c.metrics.Misses, c.metrics.Evictions
}

// startCleanupTask runs a periodic task to remove expired items
func (c *Cache) startCleanupTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired items from the cache
func (c *Cache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	evicted := 0

	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
			evicted++
		}
	}

	if evicted > 0 {
		c.metrics.recordEvictions(int64(evicted))
		c.logger.Debug("Expired cache items removed",
			zap.String("cache", c.name),
			zap.Int("count", evicted))
	}
}

// recordHit increments the hit counter
func (m *CacheMetrics) recordHit() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Hits++
}

// recordMiss increments the miss counter
func (m *CacheMetrics) recordMiss() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Misses++
}

// recordEvictions increments the eviction counter
func (m *CacheMetrics) recordEvictions(count int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Evictions += count
}

// CacheRegistry maintains a registry of caches
type CacheRegistry struct {
	caches map[string]*Cache
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewCacheRegistry creates a new cache registry
func NewCacheRegistry(logger *zap.Logger) *CacheRegistry {
	return &CacheRegistry{
		caches: make(map[string]*Cache),
		logger: logger,
	}
}

// Get returns a cache by name, creating it if it doesn't exist
func (r *CacheRegistry) Get(name string) *Cache {
	r.mutex.RLock()
	cache, exists := r.caches[name]
	r.mutex.RUnlock()

	if exists {
		return cache
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check again in case another goroutine created it while we were waiting
	cache, exists = r.caches[name]
	if exists {
		return cache
	}

	// Create new cache
	cache = NewCache(name, r.logger)
	r.caches[name] = cache
	return cache
}

// GetMetrics returns metrics for all caches
func (r *CacheRegistry) GetMetrics() map[string]map[string]int64 {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metrics := make(map[string]map[string]int64)
	for name, cache := range r.caches {
		hits, misses, evictions := cache.GetMetrics()
		metrics[name] = map[string]int64{
			"hits":      hits,
			"misses":    misses,
			"evictions": evictions,
		}
	}

	return metrics
}
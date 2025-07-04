package utils

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitClosed means the circuit is closed and operations are allowed
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen means the circuit is open and operations are not allowed
	CircuitOpen
	// CircuitHalfOpen means the circuit is testing if operations can be allowed again
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	state          CircuitBreakerState
	failureCount   int
	successCount   int
	lastStateChange time.Time
	mutex          sync.RWMutex
	logger         *zap.Logger

	// Configuration
	failureThreshold   int
	successThreshold   int
	resetTimeout       time.Duration
	halfOpenMaxRetries int
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	Name              string
	FailureThreshold  int
	SuccessThreshold  int
	ResetTimeout      time.Duration
	HalfOpenMaxRetries int
	Logger            *zap.Logger
}

// DefaultCircuitBreakerConfig provides sensible defaults
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	Name:              "default",
	FailureThreshold:  5,
	SuccessThreshold:  2,
	ResetTimeout:      30 * time.Second,
	HalfOpenMaxRetries: 1,
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.Logger == nil {
		logger, _ := zap.NewProduction()
		config.Logger = logger
	}

	return &CircuitBreaker{
		name:              config.Name,
		state:             CircuitClosed,
		failureCount:      0,
		successCount:      0,
		lastStateChange:   time.Now(),
		failureThreshold:  config.FailureThreshold,
		successThreshold:  config.SuccessThreshold,
		resetTimeout:      config.ResetTimeout,
		halfOpenMaxRetries: config.HalfOpenMaxRetries,
		logger:            config.Logger,
	}
}

// ErrCircuitOpen is returned when the circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit is open
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	// Record the result
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed through
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if reset timeout has elapsed
		if time.Since(cb.lastStateChange) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.transitionToHalfOpen()
			cb.mutex.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		// Only allow limited requests in half-open state
		return cb.successCount < cb.halfOpenMaxRetries
	default:
		return true
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case CircuitClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.transitionToClosed()
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.transitionToOpen()
		}
	case CircuitHalfOpen:
		cb.transitionToOpen()
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// transitionToOpen changes the state to open
func (cb *CircuitBreaker) transitionToOpen() {
	cb.state = CircuitOpen
	cb.lastStateChange = time.Now()
	cb.logger.Warn("Circuit breaker opened",
		zap.String("name", cb.name),
		zap.Int("failures", cb.failureCount))
}

// transitionToHalfOpen changes the state to half-open
func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = CircuitHalfOpen
	cb.lastStateChange = time.Now()
	cb.successCount = 0
	cb.logger.Info("Circuit breaker half-open",
		zap.String("name", cb.name),
		zap.Duration("after", time.Since(cb.lastStateChange)))
}

// transitionToClosed changes the state to closed
func (cb *CircuitBreaker) transitionToClosed() {
	cb.state = CircuitClosed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.logger.Info("Circuit breaker closed",
		zap.String("name", cb.name))
}

// CircuitBreakerRegistry maintains a registry of circuit breakers
type CircuitBreakerRegistry struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreakerRegistry creates a new registry
func NewCircuitBreakerRegistry(logger *zap.Logger) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// Get returns a circuit breaker by name, creating it if it doesn't exist
func (r *CircuitBreakerRegistry) Get(name string, config ...CircuitBreakerConfig) *CircuitBreaker {
	r.mutex.RLock()
	cb, exists := r.breakers[name]
	r.mutex.RUnlock()

	if exists {
		return cb
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check again in case another goroutine created it while we were waiting
	cb, exists = r.breakers[name]
	if exists {
		return cb
	}

	// Create new circuit breaker
	cfg := DefaultCircuitBreakerConfig
	cfg.Name = name
	cfg.Logger = r.logger
	
	if len(config) > 0 {
		cfg = config[0]
		cfg.Name = name
		if cfg.Logger == nil {
			cfg.Logger = r.logger
		}
	}

	cb = NewCircuitBreaker(cfg)
	r.breakers[name] = cb
	return cb
}

// GetStatus returns the status of all circuit breakers
func (r *CircuitBreakerRegistry) GetStatus() map[string]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	status := make(map[string]string)
	for name, cb := range r.breakers {
		state := "unknown"
		switch cb.GetState() {
		case CircuitClosed:
			state = "closed"
		case CircuitOpen:
			state = "open"
		case CircuitHalfOpen:
			state = "half-open"
		}
		status[name] = state
	}

	return status
}
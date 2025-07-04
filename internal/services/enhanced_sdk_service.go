package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// EnhancedSDKService provides an enhanced SDK service with improved reliability
type EnhancedSDKService struct {
	sdkRepo         repositories.SDKRepositoryInterface
	postmanClient   PostmanClientInterface
	logger          *zap.Logger
	cache           *utils.Cache
	circuitBreakers *utils.CircuitBreakerRegistry
	metrics         *utils.MetricsCollector
	workerPool      *utils.WorkerPool
}

// NewEnhancedSDKService creates a new enhanced SDK service
func NewEnhancedSDKService(
	sdkRepo repositories.SDKRepositoryInterface,
	postmanClient PostmanClientInterface,
	logger *zap.Logger,
) *EnhancedSDKService {
	// Initialize utilities
	cache := utils.NewCache("sdk_service", logger)
	circuitBreakers := utils.NewCircuitBreakerRegistry(logger)
	metrics := utils.GetGlobalMetricsCollector(logger)
	workerPool := utils.NewWorkerPool(5, 100, logger) // 5 workers, queue size 100

	// Start worker pool
	workerPool.Start()

	service := &EnhancedSDKService{
		sdkRepo:         sdkRepo,
		postmanClient:   postmanClient,
		logger:          logger,
		cache:           cache,
		circuitBreakers: circuitBreakers,
		metrics:         metrics,
		workerPool:      workerPool,
	}

	// Initialize metrics
	service.initializeMetrics()

	return service
}

// initializeMetrics sets up metrics for the service
func (s *EnhancedSDKService) initializeMetrics() {
	s.metrics.Counter("sdk_generation_requests_total", nil)
	s.metrics.Counter("sdk_generation_success_total", nil)
	s.metrics.Counter("sdk_generation_failure_total", nil)
	s.metrics.Histogram("sdk_generation_duration_ms", nil)
	s.metrics.Counter("mcp_generation_requests_total", nil)
	s.metrics.Counter("mcp_generation_success_total", nil)
	s.metrics.Counter("mcp_generation_failure_total", nil)
	s.metrics.Histogram("mcp_generation_duration_ms", nil)
	s.metrics.Gauge("active_generation_tasks", nil)
}

// SDKGenerationTask represents an SDK generation task
type SDKGenerationTask struct {
	id           string
	request      *models.SDKGenerationRequest
	recordID     primitive.ObjectID
	service      *EnhancedSDKService
	traceContext *utils.TraceContext
}

// ID returns the task ID
func (t *SDKGenerationTask) ID() string {
	return t.id
}

// Name returns the task name
func (t *SDKGenerationTask) Name() string {
	return "sdk_generation"
}

// Execute executes the SDK generation task
func (t *SDKGenerationTask) Execute(ctx context.Context) error {
	// Add trace context to the context
	ctx = utils.WithTraceContext(ctx, t.traceContext)

	t.traceContext.LogInfo("Starting SDK generation task")

	// Track active tasks
	t.service.metrics.Gauge("active_generation_tasks", nil).Add(1)
	defer t.service.metrics.Gauge("active_generation_tasks", nil).Add(-1)

	// Track generation duration
	timer := t.service.metrics.Histogram("sdk_generation_duration_ms", map[string]string{
		"language": t.request.Language,
	}).Timer()
	defer timer()

	// Execute the generation with retry logic
	retryConfig := utils.RetryConfig{
		MaxRetries:      2,
		InitialInterval: 1 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.1,
	}

	err := utils.RetryWithBackoff(ctx, t.service.logger, "sdk_generation", func() error {
		return t.service.executeSDKGeneration(ctx, t.request, t.recordID, t.traceContext)
	}, retryConfig)

	if err != nil {
		t.service.metrics.Counter("sdk_generation_failure_total", map[string]string{
			"language": t.request.Language,
		}).Inc()
		t.traceContext.LogError("SDK generation failed", err)
		return err
	}

	t.service.metrics.Counter("sdk_generation_success_total", map[string]string{
		"language": t.request.Language,
	}).Inc()
	t.traceContext.LogInfo("SDK generation completed successfully")
	return nil
}

// MCPGenerationTask represents an MCP generation task
type MCPGenerationTask struct {
	id           string
	request      *models.MCPGenerationRequest
	recordID     primitive.ObjectID
	service      *EnhancedSDKService
	traceContext *utils.TraceContext
}

// ID returns the task ID
func (t *MCPGenerationTask) ID() string {
	return t.id
}

// Name returns the task name
func (t *MCPGenerationTask) Name() string {
	return "mcp_generation"
}

// Execute executes the MCP generation task
func (t *MCPGenerationTask) Execute(ctx context.Context) error {
	// Add trace context to the context
	ctx = utils.WithTraceContext(ctx, t.traceContext)

	t.traceContext.LogInfo("Starting MCP generation task")

	// Track active tasks
	t.service.metrics.Gauge("active_generation_tasks", nil).Add(1)
	defer t.service.metrics.Gauge("active_generation_tasks", nil).Add(-1)

	// Track generation duration
	timer := t.service.metrics.Histogram("mcp_generation_duration_ms", map[string]string{
		"transport": t.request.Transport,
	}).Timer()
	defer timer()

	// Execute the generation with retry logic
	retryConfig := utils.RetryConfig{
		MaxRetries:      2,
		InitialInterval: 1 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.1,
	}

	err := utils.RetryWithBackoff(ctx, t.service.logger, "mcp_generation", func() error {
		return t.service.executeMCPGeneration(ctx, t.request, t.recordID, t.traceContext)
	}, retryConfig)

	if err != nil {
		t.service.metrics.Counter("mcp_generation_failure_total", map[string]string{
			"transport": t.request.Transport,
		}).Inc()
		t.traceContext.LogError("MCP generation failed", err)
		return err
	}

	t.service.metrics.Counter("mcp_generation_success_total", map[string]string{
		"transport": t.request.Transport,
	}).Inc()
	t.traceContext.LogInfo("MCP generation completed successfully")
	return nil
}

// GenerateSDKAsync submits an SDK generation task to the worker pool
func (s *EnhancedSDKService) GenerateSDKAsync(_ context.Context, req *models.SDKGenerationRequest, recordID primitive.ObjectID) error {
	s.metrics.Counter("sdk_generation_requests_total", map[string]string{
		"language": req.Language,
	}).Inc()

	// Create trace context
	traceContext := utils.NewTraceContext(s.logger)
	traceContext.SetTag("operation", "sdk_generation")
	traceContext.SetTag("language", req.Language)
	traceContext.SetTag("record_id", recordID.Hex())

	// Create task
	task := &SDKGenerationTask{
		id:           recordID.Hex(),
		request:      req,
		recordID:     recordID,
		service:      s,
		traceContext: traceContext,
	}

	// Submit to worker pool
	if !s.workerPool.Submit(task) {
		return utils.NewInternalError("Failed to submit SDK generation task to worker pool", nil)
	}

	traceContext.LogInfo("SDK generation task submitted to worker pool")
	return nil
}

// GenerateMCPAsync submits an MCP generation task to the worker pool
func (s *EnhancedSDKService) GenerateMCPAsync(_ context.Context, req *models.MCPGenerationRequest, recordID primitive.ObjectID) error {
	s.metrics.Counter("mcp_generation_requests_total", map[string]string{
		"transport": req.Transport,
	}).Inc()

	// Create trace context
	traceContext := utils.NewTraceContext(s.logger)
	traceContext.SetTag("operation", "mcp_generation")
	traceContext.SetTag("transport", req.Transport)
	traceContext.SetTag("record_id", recordID.Hex())

	// Create task
	task := &MCPGenerationTask{
		id:           recordID.Hex(),
		request:      req,
		recordID:     recordID,
		service:      s,
		traceContext: traceContext,
	}

	// Submit to worker pool
	if !s.workerPool.Submit(task) {
		return utils.NewInternalError("Failed to submit MCP generation task to worker pool", nil)
	}

	traceContext.LogInfo("MCP generation task submitted to worker pool")
	return nil
}

// executeSDKGeneration performs the actual SDK generation
func (s *EnhancedSDKService) executeSDKGeneration(ctx context.Context, req *models.SDKGenerationRequest, recordID primitive.ObjectID, tc *utils.TraceContext) error {
	// Get circuit breaker for Postman API
	postmanCB := s.circuitBreakers.Get("postman_api", utils.CircuitBreakerConfig{
		Name:             "postman_api",
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		Logger:           s.logger,
	})

	// Fetch collection data with circuit breaker protection
	var postmanJSON string
	err := postmanCB.Execute(func() error {
		var err error
		postmanJSON, err = s.postmanClient.GetRawCollectionJSONByID(req.CollectionID)
		if err != nil {
			return utils.NewExternalServiceError("postman", "Failed to fetch collection", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	tc.LogInfo("Postman collection fetched successfully")

	// Convert to OpenAPI with caching
	cacheKey := fmt.Sprintf("openapi:%s", req.CollectionID)
	var openAPIStr string

	if cached, found := s.cache.Get(cacheKey); found {
		openAPIStr = cached.(string)
		tc.LogInfo("OpenAPI spec retrieved from cache")
	} else {
		openAPIStr, err = s.convertPostmanToOpenAPI(ctx, postmanJSON, tc)
		if err != nil {
			return err
		}

		// Cache for 1 hour
		s.cache.Set(cacheKey, openAPIStr, time.Hour)
		tc.LogInfo("OpenAPI spec cached")
	}

	// Generate SDK based on language
	return s.generateSDKForLanguage(ctx, req, recordID, openAPIStr, tc)
}

// executeMCPGeneration performs the actual MCP generation
func (s *EnhancedSDKService) executeMCPGeneration(ctx context.Context, req *models.MCPGenerationRequest, recordID primitive.ObjectID, tc *utils.TraceContext) error {
	// Check if mcpgen is available
	if _, err := exec.LookPath("mcpgen"); err != nil {
		return utils.NewExternalServiceError("mcpgen", "mcpgen command not found in PATH", err)
	}

	// Get circuit breaker for Postman API
	postmanCB := s.circuitBreakers.Get("postman_api")

	// Fetch collection data with circuit breaker protection
	var postmanJSON string
	err := postmanCB.Execute(func() error {
		var err error
		postmanJSON, err = s.postmanClient.GetRawCollectionJSONByID(req.CollectionID)
		if err != nil {
			return utils.NewExternalServiceError("postman", "Failed to fetch collection", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	tc.LogInfo("Postman collection fetched successfully")

	// Convert to OpenAPI
	openAPIStr, err := s.convertPostmanToOpenAPI(ctx, postmanJSON, tc)
	if err != nil {
		return err
	}

	// Generate MCP server
	return s.generateMCPServer(ctx, req, recordID, openAPIStr, tc)
}

// convertPostmanToOpenAPI converts Postman collection to OpenAPI spec
func (s *EnhancedSDKService) convertPostmanToOpenAPI(ctx context.Context, postmanJSON string, tc *utils.TraceContext) (string, error) {
	childSpan := tc.NewChildSpan("postman_to_openapi_conversion")
	defer childSpan.Finish()

	childSpan.LogInfo("Converting Postman collection to OpenAPI")

	// We need access to the actual SDK service for conversion
	// For now, create a minimal SDK service instance just for conversion
	tempSDKService := &SDKService{
		logger: s.logger,
	}

	// Call the actual conversion function
	openAPIStr, err := tempSDKService.ConvertPostmanToOpenAPI(ctx, postmanJSON)
	if err != nil {
		childSpan.LogError("Postman to OpenAPI conversion failed", err)
		return "", utils.NewInternalError("Failed to convert Postman collection to OpenAPI", err)
	}

	childSpan.LogInfo("Postman to OpenAPI conversion completed successfully")
	return openAPIStr, nil
}

// generateSDKForLanguage generates SDK for a specific language
func (s *EnhancedSDKService) generateSDKForLanguage(_ context.Context, req *models.SDKGenerationRequest, recordID primitive.ObjectID, _ string, tc *utils.TraceContext) error {
	childSpan := tc.NewChildSpan("sdk_generation")
	defer childSpan.Finish()

	childSpan.SetTag("language", req.Language)
	childSpan.LogInfo("Generating SDK for language")

	// Create output directory
	outputDir := filepath.Join("generated_sdks", recordID.Hex())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return utils.NewInternalError("Failed to create output directory", err)
	}

	// Generate SDK based on language
	switch req.Language {
	default:
		return utils.NewValidationError("Unsupported language", fmt.Sprintf("Language '%s' is not supported", req.Language))
	}
}

// generateMCPServer generates an MCP server
func (s *EnhancedSDKService) generateMCPServer(ctx context.Context, req *models.MCPGenerationRequest, recordID primitive.ObjectID, openAPIStr string, tc *utils.TraceContext) error {
	childSpan := tc.NewChildSpan("mcp_server_generation")
	defer childSpan.Finish()

	childSpan.SetTag("transport", req.Transport)
	childSpan.LogInfo("Generating MCP server")

	// Create output directory
	outputDir := filepath.Join("generated_mcps", recordID.Hex())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return utils.NewInternalError("Failed to create output directory", err)
	}

	// Create temporary spec file
	tempSpecFile, err := os.CreateTemp("", "openapi_spec_*.json")
	if err != nil {
		return utils.NewInternalError("Failed to create temporary file", err)
	}
	defer os.Remove(tempSpecFile.Name())

	if _, err := tempSpecFile.Write([]byte(openAPIStr)); err != nil {
		return utils.NewInternalError("Failed to write OpenAPI spec", err)
	}
	tempSpecFile.Close()

	// Execute mcpgen command
	cmd := exec.CommandContext(ctx, "mcpgen", "generate",
		"--input", tempSpecFile.Name(),
		"--output", outputDir,
		"--transport", req.Transport,
		"--port", fmt.Sprintf("%d", req.Port),
		"--force")

	output, err := cmd.CombinedOutput()
	if err != nil {
		childSpan.LogError("mcpgen command failed", err, zap.String("output", string(output)))
		return utils.NewExternalServiceError("mcpgen", "MCP generation failed", err)
	}

	childSpan.LogInfo("MCP server generated successfully", zap.String("output", string(output)))
	return nil
}

// GetMetrics returns service metrics
func (s *EnhancedSDKService) GetMetrics() map[string]interface{} {
	return s.metrics.GetMetrics()
}

// GetCircuitBreakerStatus returns circuit breaker status
func (s *EnhancedSDKService) GetCircuitBreakerStatus() map[string]string {
	return s.circuitBreakers.GetStatus()
}

// GetWorkerPoolStatus returns worker pool status
func (s *EnhancedSDKService) GetWorkerPoolStatus() map[string]interface{} {
	return map[string]interface{}{
		"queue_size":     s.workerPool.QueueSize(),
		"queue_capacity": s.workerPool.QueueCapacity(),
		"worker_count":   s.workerPool.WorkerCount(),
	}
}

// Shutdown gracefully shuts down the service
func (s *EnhancedSDKService) Shutdown() {
	s.logger.Info("Shutting down enhanced SDK service")
	s.workerPool.Stop()
}

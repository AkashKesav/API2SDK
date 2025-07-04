package services

import (
	"context"
	"embed" // Ensure embed is imported
	"encoding/json"
	"fmt"
	"io"
	"log" // Added for init() logging
	"os"
	"os/exec"
	"path/filepath"
	"runtime" // Added for init() path resolution
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/utils" // Assuming utils.ErrNotFound, utils.ErrUnauthorized exist or handle errors appropriately
	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// SDKServiceInterface defines the operations for SDK and MCP generation and management.
type SDKServiceInterface interface {
	// CreateSDKRecord creates an initial record for an SDK or MCP generation task.
	CreateSDKRecord(ctx context.Context, sdk *models.SDK) (*models.SDK, error)

	// GenerateSDK handles the logic for generating an SDK.
	// It takes a request, converts Postman to OpenAPI, and invokes the generation process.
	// It updates the SDK record with status and file path upon completion or error.
	GenerateSDK(ctx context.Context, req *models.SDKGenerationRequest, recordID primitive.ObjectID) (*models.SDK, error)

	// GenerateMCP handles the logic for generating an MCP server.
	// Similar to GenerateSDK, it manages the generation process and updates the record.
	GenerateMCP(ctx context.Context, req *models.MCPGenerationRequest, recordID primitive.ObjectID) (*models.SDK, error)

	// GetSDKByID retrieves an SDK by its ID, ensuring user ownership.
	GetSDKByID(ctx context.Context, sdkID primitive.ObjectID, userID string) (*models.SDK, error)

	// GetSDKsByUserID retrieves a paginated list of SDKs for a specific user.
	GetSDKsByUserID(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error)

	// UpdateSDKStatus updates the status and an optional error message of an SDK record.
	UpdateSDKStatus(ctx context.Context, sdkID primitive.ObjectID, status models.SDKGenerationStatus, errorMessage string) error // Corrected to SDKGenerationStatus

	// UpdateSDKRecord updates an SDK record with new information, typically after successful generation.
	UpdateSDKRecord(ctx context.Context, sdk *models.SDK) error

	// DownloadSDK retrieves SDK metadata and file path for download, verifying ownership and status.
	DownloadSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) (*models.SDK, string, error) // Changed return to include *models.SDK

	// ConvertPostmanToOpenAPI converts a Postman collection JSON to OpenAPI v3 JSON.
	// It uses an embedded JavaScript bundle (Postman SDK) to perform the conversion.
	ConvertPostmanToOpenAPI(ctx context.Context, postmanCollectionJSON string) (string, error)

	// GetPyGenScript returns the embedded Python generation script filesystem.
	GetPyGenScript() embed.FS

	// GetPhpGenScript returns the embedded PHP generation script filesystem.
	GetPhpGenScript() embed.FS

	// GetPhpVendorZip returns the embedded PHP vendor zip filesystem.
	GetPhpVendorZip() embed.FS

	// DeleteSDK soft deletes an SDK record, verifying ownership.
	DeleteSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) error

	// GetTotalGeneratedSDKsCount returns the total number of generated SDKs (not soft-deleted)
	GetTotalGeneratedSDKsCount(ctx context.Context) (int64, error)
}

//go:embed jslibs/dist/bundle.js
var jsBundle []byte

//go:embed pylibs/generate_python_sdk.py
var PyGenScript embed.FS

//go:embed phplibs/generate_php_sdk.php
var PhpGenScript embed.FS

//go:embed phplibs/vendor.tar.gz
var PhpVendorZip embed.FS

//go:embed codegeners/openapi-generator-cli.jar
var openAPIGeneratorJar []byte

// GetPyGenScript returns the embedded Python generation script filesystem.
func (s *SDKService) GetPyGenScript() embed.FS { // Implemented for SDKService
	return s.pyGenScript
}

// GetPhpGenScript returns the embedded PHP generation script filesystem.
func (s *SDKService) GetPhpGenScript() embed.FS { // Implemented for SDKService
	return s.phpGenScript
}

// GetPhpVendorZip returns the embedded PHP vendor zip filesystem.
func (s *SDKService) GetPhpVendorZip() embed.FS { // Implemented for SDKService
	return s.phpVendorZip
}

// polyfillConsole provides a basic console object to the goja runtime.
func polyfillConsole(vm *goja.Runtime, logger *zap.Logger) {
	console := vm.NewObject()

	_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		logger.Info("[JS CONSOLE.LOG]", zap.Any("args", args))
		return goja.Undefined()
	})

	_ = console.Set("error", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		logger.Error("[JS CONSOLE.ERROR]", zap.Any("args", args))
		return goja.Undefined()
	})

	_ = console.Set("warn", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		logger.Warn("[JS CONSOLE.WARN]", zap.Any("args", args))
		return goja.Undefined()
	})

	_ = console.Set("debug", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		logger.Debug("[JS CONSOLE.DEBUG]", zap.Any("args", args))
		return goja.Undefined()
	})

	_ = vm.Set("console", console)
}

// init attempts to manually load jsBundle if go:embed failed or the file was empty.
// This serves as a fallback.
func init() {
	if len(jsBundle) == 0 {
		log.Println("INFO: jsBundle was not populated by go:embed or the embedded file is empty. Attempting manual load...")

		_, currentFilePath, _, ok := runtime.Caller(0)
		if !ok {
			log.Println("ERROR: Failed to determine current file path using runtime.Caller for manual jsBundle loading. Postman conversion may fail.")
			return
		}

		// Assuming jslibs/dist/bundle.js is relative to this sdk_service.go file's directory
		bundlePath := filepath.Join(filepath.Dir(currentFilePath), "jslibs", "dist", "bundle.js")

		content, err := os.ReadFile(bundlePath)
		if err != nil {
			log.Printf("ERROR: Failed to manually read jsBundle from %s: %v. Postman conversion will likely fail.", bundlePath, err)
			// jsBundle remains nil or empty, ConvertPostmanToOpenAPI will handle this.
			return
		}

		if len(content) == 0 {
			log.Printf("WARN: Manually loaded jsBundle from %s is empty. Postman conversion may fail.", bundlePath)
		} else {
			log.Printf("INFO: jsBundle manually loaded successfully from %s (%d bytes).", bundlePath, len(content))
		}
		jsBundle = content
	} else {
		log.Printf("INFO: jsBundle successfully embedded by go:embed (%d bytes).", len(jsBundle))
	}
}

// SDKService handles the business logic for SDK generation.
// It orchestrates the entire process of SDK generation, including
// interacting with the repository, managing temporary files, and
// invoking the appropriate code generation tools or scripts.
type SDKService struct {
	sdkRepo         repositories.SDKRepositoryInterface
	postmanClient   PostmanClientInterface
	logger          *zap.Logger
	openAPIGenPath  string   // Path to openapi-generator-cli.jar or executable
	pyGenScript     embed.FS // Embedded Python generation script
	phpGenScript    embed.FS // Embedded PHP generation script
	phpVendorZip    embed.FS // Embedded PHP vendor zip
	tempDirRootBase string   // Base for creating temporary directories for SDK generation
	jsRuntime       *goja.Runtime
	// jsErrorMutex              sync.Mutex // Remains removed
	// lastJSConversionHadErrors bool      // Remains removed
}

// NewSDKService creates a new SDKService.
// Note: mongoClient and dbName are currently for potential future use with direct DB interaction if needed,
// but core generation logic relies on sdkRepo for persistence.
func NewSDKService(
	sdkRepo repositories.SDKRepositoryInterface,
	postmanClient PostmanClientInterface,
	logger *zap.Logger,
	openAPIGenPath string,
	pyFS embed.FS, // Changed from pyGenScriptPath to pyFS to match struct
	phpFS embed.FS, // Changed from phpGenScriptPath to phpFS to match struct
	phpVendorFS embed.FS, // Changed from phpVendorZipPath to phpVendorFS
) (*SDKService, error) {
	if openAPIGenPath == "" {
		logger.Warn("OpenAPI Generator path not explicitly set. Relying on it being in PATH or pre-configured.")
	}

	tempDirRoot, err := os.MkdirTemp("", "sdk_service_temp_root_")
	if err != nil {
		logger.Error("Failed to create temporary root directory for SDK service", zap.Error(err))
		return nil, fmt.Errorf("failed to create temp root dir: %w", err)
	}
	logger.Info("SDKService initialized", zap.String("tempDirRoot", tempDirRoot))

	// Initialize goja runtime
	jsRuntime := goja.New()

	return &SDKService{
		sdkRepo:         sdkRepo,
		postmanClient:   postmanClient,
		logger:          logger,
		openAPIGenPath:  openAPIGenPath,
		pyGenScript:     pyFS,        // Correctly assign embed.FS
		phpGenScript:    phpFS,       // Correctly assign embed.FS
		phpVendorZip:    phpVendorFS, // Correctly assign embed.FS
		tempDirRootBase: tempDirRoot,
		jsRuntime:       jsRuntime,
	}, nil
}

// CreateSDKRecord creates an initial SDK record in the database.
func (s *SDKService) CreateSDKRecord(ctx context.Context, sdk *models.SDK) (*models.SDK, error) {
	s.logger.Info("Creating new SDK record", zap.String("userID", sdk.UserID), zap.String("collectionID", sdk.CollectionID), zap.String("language", sdk.Language))
	// The repository's Create method already sets ID, CreatedAt, UpdatedAt, IsDeleted
	createdSDK, err := s.sdkRepo.Create(ctx, sdk)
	if err != nil {
		s.logger.Error("Failed to create SDK record in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create SDK record: %w", err)
	}
	s.logger.Info("SDK record created successfully", zap.String("sdkID", createdSDK.ID.Hex()))
	return createdSDK, nil
}

// UpdateSDKStatus updates the status and error message of an SDK record.
func (s *SDKService) UpdateSDKStatus(ctx context.Context, sdkID primitive.ObjectID, status models.SDKGenerationStatus, errorMessage string) error { // Corrected to SDKGenerationStatus
	s.logger.Info("Updating SDK status", zap.String("sdkID", sdkID.Hex()), zap.String("status", string(status)))
	fields := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}
	if errorMessage != "" {
		fields["errorMessage"] = errorMessage
	}
	if status == models.SDKStatusFailed || status == models.SDKStatusCompleted {
		fields["finishedAt"] = time.Now()
	}

	err := s.sdkRepo.UpdateFields(ctx, sdkID, fields)
	if err != nil {
		s.logger.Error("Failed to update SDK status in repository", zap.Error(err), zap.String("sdkID", sdkID.Hex()))
		return fmt.Errorf("failed to update SDK status for %s: %w", sdkID.Hex(), err)
	}
	s.logger.Info("SDK status updated successfully", zap.String("sdkID", sdkID.Hex()))
	return nil
}

// UpdateSDKRecord updates an existing SDK record in the database.
// Typically used after successful generation to save file path and other details.
func (s *SDKService) UpdateSDKRecord(ctx context.Context, sdk *models.SDK) error {
	s.logger.Info("Updating full SDK record", zap.String("sdkID", sdk.ID.Hex()))
	sdk.UpdatedAt = time.Now() // Ensure UpdatedAt is set
	// If status is completed or failed, and FinishedAt is not set, set it.
	if (sdk.Status == models.SDKStatusCompleted || sdk.Status == models.SDKStatusFailed) && sdk.FinishedAt.IsZero() {
		sdk.FinishedAt = time.Now()
	}

	err := s.sdkRepo.Update(ctx, sdk)
	if err != nil {
		s.logger.Error("Failed to update SDK record in repository", zap.Error(err), zap.String("sdkID", sdk.ID.Hex()))
		return fmt.Errorf("failed to update SDK record %s: %w", sdk.ID.Hex(), err)
	}
	s.logger.Info("SDK record updated successfully", zap.String("sdkID", sdk.ID.Hex()))
	return nil
}

// GetSDKsByUserID retrieves a paginated list of SDKs for a specific user.
func (s *SDKService) GetSDKsByUserID(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error) {
	s.logger.Info("Fetching SDKs for user", zap.String("userID", userID), zap.Int("page", page), zap.Int("limit", limit))
	sdks, total, err := s.sdkRepo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		s.logger.Error("Failed to retrieve SDKs for user from repository", zap.Error(err), zap.String("userID", userID))
		return nil, 0, fmt.Errorf("failed to get SDKs for user %s: %w", userID, err)
	}
	s.logger.Info("SDKs retrieved for user", zap.String("userID", userID), zap.Int("count", len(sdks)), zap.Int64("total", total))
	return sdks, total, nil
}

// DownloadSDK retrieves SDK metadata and file path for download, verifying ownership and status.
// The third return value is the actual file path for download.
func (s *SDKService) DownloadSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) (*models.SDK, string, error) {
	s.logger.Info("Attempting to prepare SDK for download", zap.String("sdkID", sdkID.Hex()), zap.String("userID", userID))

	sdk, err := s.GetSDKByID(ctx, sdkID, userID) // Leverages existing ownership check
	if err != nil {
		// GetSDKByID already logs and returns descriptive errors (not found, unauthorized)
		return nil, "", err // Propagate error directly
	}

	if sdk.Status != models.SDKStatusCompleted {
		s.logger.Warn("Attempt to download SDK not in completed state", zap.String("sdkID", sdkID.Hex()), zap.String("status", string(sdk.Status)))
		return nil, "", fmt.Errorf("SDK %s is not yet ready or generation failed (status: %s)", sdkID.Hex(), sdk.Status)
	}

	if sdk.FilePath == "" {
		s.logger.Error("SDK record is completed but has no file path", zap.String("sdkID", sdkID.Hex()))
		return nil, "", fmt.Errorf("SDK %s is completed but file path is missing", sdkID.Hex())
	}
	s.logger.Info("SDK ready for download", zap.String("sdkID", sdkID.Hex()), zap.String("filePath", sdk.FilePath))
	return sdk, sdk.FilePath, nil
}

// GenerateSDK orchestrates the SDK generation process.
// It now accepts a recordID for the SDK record that was pre-created.
func (s *SDKService) GenerateSDK(ctx context.Context, genReq *models.SDKGenerationRequest, recordID primitive.ObjectID) (*models.SDK, error) {
	// Validate inputs
	if genReq == nil {
		return nil, fmt.Errorf("SDK generation request is nil")
	}
	if genReq.CollectionID == "" {
		return nil, fmt.Errorf("collection ID is required")
	}
	if genReq.Language == "" {
		return nil, fmt.Errorf("language is required")
	}
	if genReq.PackageName == "" {
		genReq.PackageName = "generated_sdk" // Set default if empty
	}

	s.logger.Info("Starting SDK generation process in service",
		zap.String("recordID", recordID.Hex()),
		zap.String("collectionID", genReq.CollectionID),
		zap.String("language", genReq.Language),
		zap.String("packageName", genReq.PackageName),
	)

	// Fetch the SDK record to update
	sdkRecord, err := s.sdkRepo.GetByID(ctx, recordID)
	if err != nil {
		s.logger.Error("Failed to fetch SDK record for generation", zap.String("recordID", recordID.Hex()), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch SDK record %s: %w", recordID.Hex(), err)
	}
	if sdkRecord == nil {
		s.logger.Error("SDK record not found for generation", zap.String("recordID", recordID.Hex()))
		return nil, fmt.Errorf("SDK record %s not found", recordID.Hex())
	}

	// Update status to InProgress
	sdkRecord.Status = models.SDKStatusInProgress
	sdkRecord.UpdatedAt = time.Now()
	if err := s.sdkRepo.Update(ctx, sdkRecord); err != nil {
		s.logger.Error("Failed to update SDK status to InProgress", zap.String("recordID", recordID.Hex()), zap.Error(err))
		// Continue generation, but log the error. The final status update will hopefully succeed.
	}

	// Create a temporary directory for generation with proper cleanup
	tempGenDir, err := utils.CreateTempDirForSDK(s.tempDirRootBase, recordID.Hex())
	if err != nil {
		s.logger.Error("Failed to create temp directory for SDK generation", zap.Error(err), zap.String("recordID", recordID.Hex()))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create temp dir: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord) // Attempt to update status
		return sdkRecord, err
	}
	defer func() {
		if removeErr := os.RemoveAll(tempGenDir); removeErr != nil {
			s.logger.Warn("Failed to clean up temp directory", zap.String("tempDir", tempGenDir), zap.Error(removeErr))
		}
	}()

	// Step 1: Fetch the collection data (e.g., Postman JSON) using genReq.CollectionID
	s.logger.Info("Fetching Postman collection data", zap.String("collectionID", genReq.CollectionID))
	postmanJSON, err := s.postmanClient.GetRawCollectionJSONByID(genReq.CollectionID)
	if err != nil {
		s.logger.Error("Failed to fetch Postman collection data", zap.String("collectionID", genReq.CollectionID), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to fetch collection data: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to fetch collection data: %w", err)
	}

	// Validate that we got valid JSON
	if strings.TrimSpace(postmanJSON) == "" {
		s.logger.Error("Received empty Postman collection data", zap.String("collectionID", genReq.CollectionID))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = "Received empty collection data"
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("received empty collection data for collection %s", genReq.CollectionID)
	}

	// Step 2: Convert Postman to OpenAPI if necessary
	s.logger.Info("Converting Postman collection to OpenAPI", zap.String("collectionID", genReq.CollectionID))
	openAPIStr, err := s.ConvertPostmanToOpenAPI(ctx, postmanJSON)
	if err != nil {
		s.logger.Error("Failed to convert Postman to OpenAPI", zap.String("collectionID", genReq.CollectionID), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to convert to OpenAPI: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to convert to OpenAPI: %w", err)
	}

	// Validate the OpenAPI spec
	if strings.TrimSpace(openAPIStr) == "" {
		s.logger.Error("OpenAPI conversion returned empty result", zap.String("collectionID", genReq.CollectionID))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = "OpenAPI conversion returned empty result"
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("OpenAPI conversion returned empty result for collection %s", genReq.CollectionID)
	}

	// Step 3: Write the OpenAPI spec to a file in tempGenDir for processing
	openAPIFilePath := filepath.Join(tempGenDir, "openapi.json")
	if err := os.WriteFile(openAPIFilePath, []byte(openAPIStr), 0644); err != nil {
		s.logger.Error("Failed to write OpenAPI spec to file", zap.String("filePath", openAPIFilePath), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to write OpenAPI spec: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}
	s.logger.Info("OpenAPI spec written to file", zap.String("filePath", openAPIFilePath))

	// Step 4: Invoke the appropriate generation tool based on genReq.Language
	var generatedSDKPath string
	finalSDKDir := filepath.Join("generated_sdks", recordID.Hex())
	if err := os.MkdirAll(finalSDKDir, 0755); err != nil {
		s.logger.Error("Failed to create final SDK directory", zap.String("dirPath", finalSDKDir), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create SDK directory: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to create SDK directory: %w", err)
	}

	// Validate language support
	supportedLanguages := []string{"go", "typescript", "python", "java", "csharp", "rust", "ruby", "php"}
	languageSupported := false
	for _, lang := range supportedLanguages {
		if genReq.Language == lang {
			languageSupported = true
			break
		}
	}
	if !languageSupported {
		err = fmt.Errorf("unsupported language: %s. Supported languages: %v", genReq.Language, supportedLanguages)
		s.logger.Error("Unsupported language requested", zap.String("language", genReq.Language), zap.Strings("supported", supportedLanguages))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = err.Error()
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, err
	}

	switch genReq.Language {
	case "go", "typescript", "python", "java", "csharp", "rust", "ruby":
		generatedSDKPath, err = s.generateWithOpenAPIGenerator(ctx, openAPIFilePath, genReq.Language, finalSDKDir, genReq.PackageName, tempGenDir, genReq.CollectionID)
	case "php":
		generatedSDKPath, err = s.generatePHPSDK(ctx, openAPIFilePath, finalSDKDir, genReq.PackageName, tempGenDir, genReq.CollectionID)
	default:
		err = fmt.Errorf("unsupported language: %s", genReq.Language)
	}

	if err != nil {
		s.logger.Error("Failed to generate SDK", zap.String("language", genReq.Language), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to generate SDK: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to generate SDK: %w", err)
	}

	// Validate that the generated SDK path exists
	if generatedSDKPath == "" {
		s.logger.Error("SDK generation completed but returned empty path", zap.String("language", genReq.Language))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = "SDK generation completed but returned empty path"
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("SDK generation completed but returned empty path for language %s", genReq.Language)
	}

	if _, err := os.Stat(generatedSDKPath); err != nil {
		s.logger.Error("Generated SDK file does not exist", zap.String("path", generatedSDKPath), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Generated SDK file does not exist: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("generated SDK file does not exist: %w", err)
	}

	// Update SDK record with the final path and status
	sdkRecord.FilePath = generatedSDKPath
	sdkRecord.Status = models.SDKStatusCompleted
	sdkRecord.ErrorMessage = "" // Clear any previous error message
	sdkRecord.FinishedAt = time.Now()
	sdkRecord.UpdatedAt = time.Now()

	s.logger.Info("SDK generation completed successfully",
		zap.String("recordID", recordID.Hex()),
		zap.String("filePath", sdkRecord.FilePath),
		zap.String("language", genReq.Language),
	)
	// The controller will call UpdateSDKRecord with this returned sdkRecord
	return sdkRecord, nil
}

// GenerateMCP orchestrates the MCP server generation process.
// It now accepts a recordID for the SDK record that was pre-created.
func (s *SDKService) GenerateMCP(ctx context.Context, genReq *models.MCPGenerationRequest, recordID primitive.ObjectID) (*models.SDK, error) {
	s.logger.Info("Starting MCP generation process in service",
		zap.String("recordID", recordID.Hex()),
		zap.String("collectionID", genReq.CollectionID),
		zap.String("transport", string(genReq.Transport)),
	)

	sdkRecord, err := s.sdkRepo.GetByID(ctx, recordID)
	if err != nil {
		s.logger.Error("Failed to fetch SDK record for MCP generation", zap.String("recordID", recordID.Hex()), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch SDK record %s for MCP: %w", recordID.Hex(), err)
	}
	if sdkRecord == nil {
		s.logger.Error("SDK record not found for MCP generation", zap.String("recordID", recordID.Hex()))
		return nil, fmt.Errorf("SDK record %s not found for MCP", recordID.Hex())
	}

	sdkRecord.Status = models.SDKStatusInProgress
	sdkRecord.UpdatedAt = time.Now()
	if err := s.sdkRepo.Update(ctx, sdkRecord); err != nil {
		s.logger.Error("Failed to update SDK status to InProgress for MCP", zap.String("recordID", recordID.Hex()), zap.Error(err))
		// Continue generation, log error.
	}

	// Create a temporary directory for generation
	tempGenDir, err := utils.CreateTempDirForSDK(s.tempDirRootBase, recordID.Hex())
	if err != nil {
		s.logger.Error("Failed to create temp directory for MCP generation", zap.Error(err), zap.String("recordID", recordID.Hex()))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create temp dir: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord) // Attempt to update status
		return sdkRecord, err
	}
	defer os.RemoveAll(tempGenDir) // Clean up temp directory

	// Step 1: Fetch the collection data (e.g., Postman JSON) using genReq.CollectionID
	s.logger.Info("Fetching Postman collection data for MCP", zap.String("collectionID", genReq.CollectionID))
	postmanJSON, err := s.postmanClient.GetRawCollectionJSONByID(genReq.CollectionID)
	if err != nil {
		s.logger.Error("Failed to fetch Postman collection data for MCP", zap.String("collectionID", genReq.CollectionID), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to fetch collection data for MCP: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to fetch collection data for MCP: %w", err)
	}

	// Step 2: Convert Postman to OpenAPI if necessary
	s.logger.Info("Converting Postman collection to OpenAPI for MCP", zap.String("collectionID", genReq.CollectionID))
	openAPIStr, err := s.ConvertPostmanToOpenAPI(ctx, postmanJSON)
	if err != nil {
		s.logger.Error("Failed to convert Postman to OpenAPI for MCP", zap.String("collectionID", genReq.CollectionID), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to convert to OpenAPI for MCP: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to convert to OpenAPI for MCP: %w", err)
	}

	// Step 3: Write the OpenAPI spec to a file in tempGenDir for processing
	openAPIFilePath := filepath.Join(tempGenDir, "openapi.json")
	if err := os.WriteFile(openAPIFilePath, []byte(openAPIStr), 0644); err != nil {
		s.logger.Error("Failed to write OpenAPI spec to file for MCP", zap.String("filePath", openAPIFilePath), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to write OpenAPI spec for MCP: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to write OpenAPI spec for MCP: %w", err)
	}
	s.logger.Info("OpenAPI spec written to file for MCP", zap.String("filePath", openAPIFilePath))

	// Step 4: Invoke the MCP generation tool based on genReq.Transport and other parameters
	finalMCPDir := filepath.Join("generated_mcps", recordID.Hex())
	if err := os.MkdirAll(finalMCPDir, 0755); err != nil {
		s.logger.Error("Failed to create final MCP directory", zap.String("dirPath", finalMCPDir), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create MCP directory: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to create MCP directory: %w", err)
	}

	generatedMCPPath, mcpID, err := s.GenerateMCPServer(ctx, sdkRecord.UserID, genReq.CollectionID, openAPIStr, finalMCPDir, string(genReq.Transport), genReq.Port)
	if err != nil {
		s.logger.Error("Failed to generate MCP server", zap.String("transport", string(genReq.Transport)), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to generate MCP server: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to generate MCP server: %w", err)
	}

	// Zip the generated MCP directory
	finalMCPPath := filepath.Join(finalMCPDir, "mcp_server.zip")
	if err := utils.ZipDirectory(generatedMCPPath, finalMCPPath); err != nil {
		s.logger.Error("Failed to zip MCP server", zap.String("sourceDir", generatedMCPPath), zap.Error(err))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to zip MCP server: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, fmt.Errorf("failed to zip MCP server: %w", err)
	}

	// Update SDK record with the final path and status
	sdkRecord.FilePath = finalMCPPath
	sdkRecord.Status = models.SDKStatusCompleted
	sdkRecord.ErrorMessage = ""
	sdkRecord.FinishedAt = time.Now()
	sdkRecord.UpdatedAt = time.Now()

	s.logger.Info("MCP generation completed successfully",
		zap.String("recordID", recordID.Hex()),
		zap.String("filePath", sdkRecord.FilePath),
		zap.String("mcpID", mcpID),
	)
	return sdkRecord, nil
}

// GetSDKByID retrieves an SDK by its ID and verifies user ownership.
func (s *SDKService) GetSDKByID(ctx context.Context, sdkID primitive.ObjectID, userID string) (*models.SDK, error) {
	s.logger.Info("Fetching SDK by ID for user", zap.String("sdkID", sdkID.Hex()), zap.String("userID", userID))
	sdk, err := s.sdkRepo.GetByID(ctx, sdkID)
	if err != nil {
		if err == mongo.ErrNoDocuments { // Assuming repository returns mongo.ErrNoDocuments
			s.logger.Warn("SDK not found", zap.String("sdkID", sdkID.Hex()))
			return nil, fmt.Errorf("SDK with ID %s not found", sdkID.Hex()) // Consider a utils.ErrNotFound
		}
		s.logger.Error("Failed to get SDK by ID from repository", zap.Error(err), zap.String("sdkID", sdkID.Hex()))
		return nil, fmt.Errorf("failed to retrieve SDK %s: %w", sdkID.Hex(), err)
	}

	if sdk.UserID != userID {
		s.logger.Warn("User not authorized to access SDK", zap.String("sdkID", sdkID.Hex()), zap.String("sdkUserID", sdk.UserID), zap.String("requestUserID", userID))
		return nil, fmt.Errorf("user not authorized to access SDK %s", sdkID.Hex()) // Consider a utils.ErrUnauthorized
	}

	s.logger.Info("SDK retrieved successfully by ID for user", zap.String("sdkID", sdkID.Hex()))
	return sdk, nil
}

// DeleteSDK soft deletes an SDK record, verifying ownership.
// It also attempts to remove the generated SDK file/directory.
func (s *SDKService) DeleteSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) error {
	s.logger.Info("Attempting to delete SDK", zap.String("sdkID", sdkID.Hex()), zap.String("userID", userID))

	// Verify ownership first
	sdk, err := s.sdkRepo.GetByID(ctx, sdkID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logger.Warn("SDK not found for deletion", zap.String("sdkID", sdkID.Hex()))
			return fmt.Errorf("SDK with ID %s not found", sdkID.Hex()) // Or a more specific error like utils.ErrNotFound
		}
		s.logger.Error("Failed to get SDK by ID from repository for deletion", zap.Error(err), zap.String("sdkID", sdkID.Hex()))
		return fmt.Errorf("failed to retrieve SDK %s for deletion: %w", sdkID.Hex(), err)
	}

	if sdk.UserID != userID {
		s.logger.Warn("User not authorized to delete SDK",
			zap.String("sdkID", sdkID.Hex()),
			zap.String("recordUserID", sdk.UserID),
			zap.String("requestUserID", userID))
		return fmt.Errorf("user not authorized to delete SDK %s", sdkID.Hex()) // Or utils.ErrUnauthorized
	}

	// Soft delete the record in the database
	// The repository's SoftDelete might also take userID for an additional check, which is fine.
	err = s.sdkRepo.SoftDelete(ctx, sdkID, userID)
	if err != nil {
		s.logger.Error("Failed to soft delete SDK record in repository", zap.Error(err), zap.String("sdkID", sdkID.Hex()))
		return fmt.Errorf("failed to delete SDK record %s: %w", sdkID.Hex(), err)
	}
	s.logger.Info("SDK record soft deleted successfully", zap.String("sdkID", sdkID.Hex()))

	// Attempt to delete the SDK file/directory if FilePath is set
	if sdk.FilePath != "" {
		// Determine if it's a file or directory.
		// For simplicity, let's assume FilePath might point to a zip file or a directory containing the SDK.
		// A common pattern is that the FilePath points to a zip file, and the actual SDK was generated in a directory
		// which might be a sibling or parent of this zip file, or the zip is the root.
		// If sdk.FilePath is a directory, os.RemoveAll is appropriate. If it's a file, os.Remove would be better,
		// but os.RemoveAll also works for single files.
		// Let's assume the stored FilePath is the primary artifact (e.g. a zip file or the main generated folder).
		// A more robust solution might involve storing the generation directory separately if cleanup is complex.

		// We'll try to remove the path stored in sdk.FilePath.
		// If it's a directory, os.RemoveAll will work. If it's a file, os.Remove would be better,
		// but os.RemoveAll also works for single files.
		s.logger.Info("Attempting to delete SDK artifact from filesystem", zap.String("filePath", sdk.FilePath))
		if err := os.RemoveAll(sdk.FilePath); err != nil {
			// Log the error but don't fail the whole operation, as the DB record is deleted.
			// This could be due to permissions, file not found (already cleaned up?), etc.
			s.logger.Error("Failed to delete SDK artifact from filesystem",
				zap.String("sdkID", sdkID.Hex()),
				zap.String("filePath", sdk.FilePath),
				zap.Error(err))
			// Optionally, you could update the SDK record to note that file deletion failed,
			// but since it's soft-deleted, this might be overkill.
		} else {
			s.logger.Info("SDK artifact deleted successfully from filesystem", zap.String("filePath", sdk.FilePath))
		}
	} else {
		s.logger.Info("No file path set for SDK, skipping filesystem deletion.", zap.String("sdkID", sdkID.Hex()))
	}

	return nil
}

// ConvertPostmanToOpenAPI converts a Postman collection JSON to OpenAPI v3 JSON.
// It uses an embedded JavaScript bundle (Postman SDK) to perform the conversion.
func (s *SDKService) ConvertPostmanToOpenAPI(ctx context.Context, postmanCollectionJSON string) (string, error) {
	s.logger.Info("Starting Postman to OpenAPI conversion")

	// Validate input
	if strings.TrimSpace(postmanCollectionJSON) == "" {
		return "", fmt.Errorf("postman collection JSON is empty")
	}

	// Validate that the input is valid JSON
	var postmanCollection interface{}
	if err := json.Unmarshal([]byte(postmanCollectionJSON), &postmanCollection); err != nil {
		return "", fmt.Errorf("invalid Postman collection JSON: %w", err)
	}

	// Check if it looks like a Postman collection
	if collectionMap, ok := postmanCollection.(map[string]interface{}); ok {
		if info, hasInfo := collectionMap["info"]; hasInfo {
			if infoMap, isInfoMap := info.(map[string]interface{}); isInfoMap {
				if _, hasSchema := infoMap["schema"]; !hasSchema {
					s.logger.Warn("Postman collection may be missing schema information")
				}
			}
		} else {
			return "", fmt.Errorf("invalid Postman collection: missing 'info' field")
		}
	} else {
		return "", fmt.Errorf("invalid Postman collection: root must be an object")
	}

	if len(jsBundle) == 0 {
		s.logger.Error("JavaScript bundle (jsBundle) is nil or empty. This indicates an issue with embedding or loading the bundle.js file.")
		return "", fmt.Errorf("critical: JavaScript bundle (jsBundle) not loaded or empty")
	}

	// Create a new VM for this conversion to avoid state issues
	vm := goja.New()
	polyfillConsole(vm, s.logger)

	// Add timeout for JavaScript execution
	jsCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Set up a goroutine to cancel execution if context times out
	done := make(chan bool, 1)
	go func() {
		select {
		case <-jsCtx.Done():
			vm.Interrupt("timeout")
		case <-done:
		}
	}()
	defer func() { done <- true }()

	s.logger.Debug("Executing JavaScript bundle for Postman conversion", zap.Int("bundle_size_bytes", len(jsBundle)))
	_, err := vm.RunString(string(jsBundle))
	if err != nil {
		s.logger.Error("Failed to execute JavaScript bundle", zap.Error(err))
		return "", fmt.Errorf("failed to execute JavaScript bundle: %w", err)
	}
	s.logger.Debug("JavaScript bundle executed successfully")

	convertFunc, ok := goja.AssertFunction(vm.Get("myCustomP2OFunction"))
	if !ok {
		s.logger.Error("myCustomP2OFunction function not found in JavaScript bundle")
		return "", fmt.Errorf("myCustomP2OFunction function not found in JavaScript bundle")
	}
	s.logger.Debug("myCustomP2OFunction function retrieved successfully")

	s.logger.Debug("Calling myCustomP2OFunction JavaScript function with Postman collection and options for JSON output")
	// Pass options to ensure JSON output format
	optionsMap := map[string]interface{}{
		"outputFormat":                "json",
		"validate":                    true,
		"schemaFaker":                 false,
		"requestParametersResolution": "Example",
		"exampleParametersResolution": "Example",
		"folderStrategy":              "Tags",
	}
	optionsBytes, marshalErr := json.Marshal(optionsMap)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal options for JS call", zap.Error(marshalErr))
		return "", fmt.Errorf("failed to marshal options for JS call: %w", marshalErr)
	}
	optionsJSONString := string(optionsBytes)

	jsPromiseValue, err := convertFunc(goja.Undefined(), vm.ToValue(postmanCollectionJSON), vm.ToValue(optionsJSONString))
	if err != nil {
		s.logger.Error("Failed to call myCustomP2OFunction", zap.Error(err))
		return "", fmt.Errorf("failed to call myCustomP2OFunction: %w", err)
	}

	s.logger.Debug("myCustomP2OFunction called successfully, waiting for promise resolution")

	promiseCh := s.waitForPromise(vm, jsPromiseValue)
	var result promiseResult

	select {
	case result = <-promiseCh:
		s.logger.Debug("Promise resolved")
	case <-jsCtx.Done():
		return "", fmt.Errorf("JavaScript conversion timed out after 2 minutes")
	}

	if result.Error != nil {
		s.logger.Error("Promise rejected during Postman to OpenAPI conversion", zap.Error(result.Error))
		return "", fmt.Errorf("conversion failed: %w", result.Error)
	}

	if result.Value == nil || goja.IsUndefined(result.Value) || goja.IsNull(result.Value) {
		s.logger.Error("Conversion promise resolved with null/undefined value")
		return "", fmt.Errorf("conversion returned null/undefined result")
	}

	resultStr := result.Value.String()
	if resultStr == "" {
		s.logger.Error("Conversion promise resolved with empty string")
		return "", fmt.Errorf("conversion returned empty result")
	}

	s.logger.Debug("Conversion result received", zap.Int("result_length", len(resultStr)))

	// Validate that the result is valid JSON
	var openAPISpec interface{}
	if err := json.Unmarshal([]byte(resultStr), &openAPISpec); err != nil {
		s.logger.Error("Conversion result is not valid JSON", zap.Error(err))
		return "", fmt.Errorf("conversion result is not valid JSON: %w", err)
	}

	// Basic validation that it looks like an OpenAPI spec
	if specMap, ok := openAPISpec.(map[string]interface{}); ok {
		if openapi, hasOpenAPI := specMap["openapi"]; hasOpenAPI {
			if openapiStr, isString := openapi.(string); isString {
				if !strings.HasPrefix(openapiStr, "3.") {
					s.logger.Warn("Generated OpenAPI spec may not be version 3.x", zap.String("version", openapiStr))
				}
			}
		} else {
			s.logger.Warn("Generated spec may not be valid OpenAPI: missing 'openapi' field")
		}

		if _, hasInfo := specMap["info"]; !hasInfo {
			s.logger.Warn("Generated OpenAPI spec may be missing 'info' field")
		}

		if _, hasPaths := specMap["paths"]; !hasPaths {
			s.logger.Warn("Generated OpenAPI spec may be missing 'paths' field")
		}
	}

	s.logger.Info("Postman to OpenAPI conversion completed successfully",
		zap.Int("input_length", len(postmanCollectionJSON)),
		zap.Int("output_length", len(resultStr)))

	return resultStr, nil
}

// promiseResult is a helper struct to pass promise results over a channel.
type promiseResult struct {
	Value goja.Value
	Error error
}

// waitForPromise waits for a goja.Promise to resolve or reject and sends the result to a channel.
// It now accepts the goja.Runtime (vm) to create callbacks in the correct context.
func (s *SDKService) waitForPromise(vm *goja.Runtime, promiseVal goja.Value) <-chan promiseResult { // Changed p *goja.Promise to promiseVal goja.Value
	ch := make(chan promiseResult, 1)

	go func() {
		defer close(ch)

		// Cast the goja.Value to *goja.Object to call Get("then")
		promiseObject := promiseVal.ToObject(vm)
		if promiseObject == nil {
			ch <- promiseResult{Error: fmt.Errorf("promiseVal.ToObject(vm) returned nil")}
			return
		}

		thenMethodVal := promiseObject.Get("then")
		thenMethod, ok := goja.AssertFunction(thenMethodVal)
		if !ok {
			ch <- promiseResult{Error: fmt.Errorf("'then' method not found on promise object or not a function")}
			return
		}

		onFulfilled := func(call goja.FunctionCall) goja.Value {
			s.logger.Debug("Promise onFulfilled called", zap.Int("num_args", len(call.Arguments)))
			if len(call.Arguments) > 0 {
				ch <- promiseResult{Value: call.Argument(0)}
			} else {
				ch <- promiseResult{Value: goja.Undefined()} // Or handle as an error: Error: fmt.Errorf("onFulfilled called with no arguments")
			}
			return goja.Undefined()
		}

		onRejected := func(call goja.FunctionCall) goja.Value {
			s.logger.Debug("Promise onRejected called", zap.Int("num_args", len(call.Arguments)))
			var err error
			if len(call.Arguments) > 0 {
				arg0 := call.Argument(0)
				if gojaErr, isGojaErr := arg0.Export().(error); isGojaErr {
					err = fmt.Errorf("JavaScript error: %w", gojaErr)
				} else if jsErrObj, isJsErrObj := arg0.(*goja.Object); isJsErrObj {
					// Try to get message and stack from JS error object
					errMsg := jsErrObj.Get("message")
					stack := jsErrObj.Get("stack")
					if errMsg != nil && !goja.IsUndefined(errMsg) && !goja.IsNull(errMsg) {
						err = fmt.Errorf("JavaScript rejected: %s (stack: %s)", errMsg.String(), stack)
					} else {
						err = fmt.Errorf("JavaScript rejected with object: %s", jsErrObj.String())
					}
				} else {
					err = fmt.Errorf("JavaScript rejected with: %s", arg0.String())
				}
			} else {
				err = fmt.Errorf("JavaScript rejected with no reason")
			}
			ch <- promiseResult{Error: err}
			return goja.Undefined()
		}

		// Create goja.Callable functions using the passed vm context.
		// The goja.Runtime.ToValue method can convert Go functions to goja.Value,
		// which, if they match the function signature, are callable from JS.
		fulfilledCallback := vm.ToValue(onFulfilled)
		rejectedCallback := vm.ToValue(onRejected)

		// Call the 'then' method on the promise object itself (this context for JS)
		// with the two callbacks.
		_, callErr := thenMethod(promiseObject, fulfilledCallback, rejectedCallback) // `this` should be the promiseObject
		if callErr != nil {
			ch <- promiseResult{Error: fmt.Errorf("error calling 'then' method on promise object: %w", callErr)}
			return
		}

		// Goja's event loop needs to run for promises to resolve.
		// If the `vm` is the same as `s.jsRuntime`, and `s.jsRuntime` has an active event loop (e.g., via Loop()),
		// this should work. If `vm` is a temporary one, it might need its own loop management if `Then` doesn't block
		// or if callbacks are not immediate.
		// For now, assume goja handles this scheduling internally when `Then` is called.
		// If issues persist, we might need to explicitly run vm.Loop() if it's a new/separate vm.
	}()

	return ch
}

// GenerateMCPServer generates an MCP server project from an OpenAPI specification string using the mcpgen Go module.
// This is a separate method that was previously part of the monolithic GenerateSDK.
// It's kept for potential direct use but the main flow uses GenerateMCP via the interface.
func (s *SDKService) GenerateMCPServer(ctx context.Context, userID, collectionID, openAPISpecContent, outputDir, transport string, port int) (string, string, error) {
	startTime := time.Now()
	s.logger.Info("Starting MCP server generation using mcpgen",
		zap.String("userID", userID),
		zap.String("collectionID", collectionID),
		zap.String("outputDir", outputDir),
		zap.String("transport", transport),
		zap.Int("port", port),
	)

	mcpRecord := &models.SDK{
		UserID:         userID,
		CollectionID:   collectionID,
		GenerationType: models.GenerationTypeMCP,
		MCPTransport:   transport,
		MCPPort:        port,
		Status:         models.SDKStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdRecord, err := s.sdkRepo.Create(ctx, mcpRecord)
	if err != nil {
		s.logger.Error("Failed to create MCP record", zap.Error(err))
		return "", "", fmt.Errorf("failed to create MCP record: %w", err)
	}
	mcpRecord = createdRecord
	s.logger.Info("MCP record created", zap.String("sdkID", mcpRecord.ID.Hex()))

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		s.logger.Error("Failed to create output directory", zap.String("outputDir", outputDir), zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to create output directory %s: %s", outputDir, err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	tempSpecFile, err := os.CreateTemp("", "openapi_spec_*.json")
	if err != nil {
		s.logger.Error("Failed to create temporary file for OpenAPI spec", zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to create temp spec file: %s", err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to create temp spec file: %w", err)
	}
	defer os.Remove(tempSpecFile.Name())
	if _, err := tempSpecFile.Write([]byte(openAPISpecContent)); err != nil {
		s.logger.Error("Failed to write OpenAPI spec to temporary file", zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to write spec to temp file: %s", err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to write spec to temp file: %w", err)
	}
	tempSpecFile.Close()

	// Check if mcpgen is available in PATH
	_, err = exec.LookPath("mcpgen")
	if err != nil {
		errorMsg := "mcpgen command not found in PATH. Please install mcpgen CLI tool."
		s.logger.Error(errorMsg, zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), errorMsg)
		return "", mcpRecord.ID.Hex(), fmt.Errorf("%s", errorMsg)
	}

	cmd := exec.Command("mcpgen", "generate",
		"--input", tempSpecFile.Name(),
		"--output", outputDir,
		"--transport", transport,
		"--port", fmt.Sprintf("%d", port),
		"--force")

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Failed to generate MCP server", zap.Error(err), zap.String("output", string(output)))
		errorMsg := fmt.Sprintf("mcpgen command failed: %s, output: %s. Ensure 'mcpgen' is installed and in PATH.", err.Error(), string(output))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), errorMsg)
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to generate MCP server: %w, output: %s", err, string(output))
	}

	s.logger.Info("MCP server generation completed", zap.String("outputDir", outputDir), zap.String("commandOutput", string(output)))
	s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusCompleted, time.Since(startTime).Milliseconds(), "")

	return outputDir, mcpRecord.ID.Hex(), nil
}

// generateWithOpenAPIGenerator uses the OpenAPI Generator CLI to generate SDKs.
// It abstracts the command execution and error handling for SDK generation.
func (s *SDKService) generateWithOpenAPIGenerator(ctx context.Context, openAPISpecPath, language, outputDir, packageName, tempDir, collectionID string) (string, error) {
	// Validate inputs
	if openAPISpecPath == "" || language == "" || outputDir == "" || packageName == "" {
		return "", fmt.Errorf("missing required parameters for SDK generation")
	}

	// Validate OpenAPI spec file exists and is readable
	if _, err := os.Stat(openAPISpecPath); err != nil {
		return "", fmt.Errorf("OpenAPI spec file not accessible: %w", err)
	}

	s.logger.Info("Generating SDK with OpenAPI Generator",
		zap.String("language", language),
		zap.String("openAPISpecPath", openAPISpecPath),
		zap.String("outputDir", outputDir),
		zap.String("packageName", packageName),
	)

	var cmd *exec.Cmd
	var generatedSDKPath string

	generatorJarPath := s.openAPIGenPath
	if generatorJarPath == "" || generatorJarPath == "openapi-generator-cli.jar" {
		tempJarFile, err := os.CreateTemp(tempDir, "openapi-generator-cli-*.jar")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file for generator JAR: %w", err)
		}
		defer tempJarFile.Close()
		defer os.Remove(tempJarFile.Name())

		if _, err := tempJarFile.Write(openAPIGeneratorJar); err != nil {
			return "", fmt.Errorf("failed to write embedded JAR to temp file: %w", err)
		}
		generatorJarPath = tempJarFile.Name()
		s.logger.Info("Using embedded OpenAPI Generator JAR", zap.String("path", generatorJarPath))
	}

	// Get configurable values or use defaults
	gitUserID := "api2sdk-generated"
	organizationName := "com.api2sdk"
	vendorName := "api2sdk"

	switch language {
	case "go":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "go",
			"-o", outputDir,
			"--package-name", packageName,
			"--git-user-id", gitUserID,
			"--git-repo-id", fmt.Sprintf("%s-go", packageName),
			"--additional-properties", "packageUrl=github.com/"+gitUserID+"/"+packageName,
		)
		generatedSDKPath = outputDir

	case "typescript", "typescript-axios":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "typescript-axios",
			"-o", outputDir,
			"--additional-properties", fmt.Sprintf("npmName=%s,supportsES6=true,usePromises=true,npmVersion=1.0.0", packageName),
		)
		generatedSDKPath = outputDir

	case "python":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "python",
			"-o", outputDir,
			"--additional-properties", fmt.Sprintf("packageName=%s,projectName=%s,packageVersion=1.0.0,library=urllib3", packageName, packageName),
		)
		generatedSDKPath = outputDir

	case "php":
		pascalCasePackageName := utils.ConvertToPascalCase(packageName)
		composerProjectName := utils.ConvertToSnakeCase(packageName)

		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "php",
			"-o", outputDir,
			"--additional-properties",
			fmt.Sprintf("composerVendorName=%s,composerProjectName=%s,invokerPackage=%s,variableNamingConvention=camelCase,packageVersion=1.0.0",
				vendorName, composerProjectName, pascalCasePackageName),
		)
		generatedSDKPath = outputDir

	case "java":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "java",
			"-o", outputDir,
			"--artifact-id", packageName,
			"--group-id", organizationName,
			"--api-package", fmt.Sprintf("%s.%s.api", organizationName, packageName),
			"--model-package", fmt.Sprintf("%s.%s.model", organizationName, packageName),
			"--library", "native",
			"--additional-properties", "artifactVersion=1.0.0",
		)
		generatedSDKPath = outputDir

	case "csharp", "csharp-netcore":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "csharp",
			"-o", outputDir,
			"--package-name", utils.ConvertToPascalCase(packageName),
			"--additional-properties", fmt.Sprintf("targetFramework=net6.0,packageName=%s,packageVersion=1.0.0", utils.ConvertToPascalCase(packageName)),
		)
		generatedSDKPath = outputDir

	case "rust":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "rust",
			"-o", outputDir,
			"--package-name", utils.ConvertToSnakeCase(packageName),
			"--additional-properties", fmt.Sprintf("packageName=%s,packageVersion=1.0.0", utils.ConvertToSnakeCase(packageName)),
		)
		generatedSDKPath = outputDir

	case "ruby":
		pascalCasePackageName := utils.ConvertToPascalCase(packageName)
		snakeCasePackageName := utils.ConvertToSnakeCase(packageName)
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "ruby",
			"-o", outputDir,
			"--additional-properties",
			fmt.Sprintf("moduleName=%s,gemName=%s,gemVersion=1.0.0", pascalCasePackageName, snakeCasePackageName),
		)
		generatedSDKPath = outputDir

	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	// Add timeout context for generation
	genCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	cmd = exec.CommandContext(genCtx, cmd.Args[0], cmd.Args[1:]...)

	s.logger.Info("Executing SDK generation command",
		zap.String("language", language),
		zap.String("command", cmd.String()))

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		s.logger.Error("SDK generation command failed",
			zap.String("language", language),
			zap.Error(cmdErr),
			zap.String("output", string(output)),
		)

		// Check for specific error types
		if genCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("SDK generation timed out after 15 minutes for language %s", language)
		}
		return "", fmt.Errorf("generation command failed for %s: %s. Output: %s", language, cmdErr, string(output))
	}

	s.logger.Info("SDK generation command completed successfully",
		zap.String("language", language),
		zap.String("output", string(output)),
	)

	// Verify the generated SDK path exists and contains files
	fi, statErr := os.Stat(generatedSDKPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			s.logger.Error("Generated SDK path does not exist after generation",
				zap.String("language", language),
				zap.String("expectedPath", generatedSDKPath),
				zap.String("generatorOutput", string(output)),
			)
			return "", fmt.Errorf("generated path %s does not exist for language %s. Output: %s", generatedSDKPath, language, string(output))
		}
		s.logger.Error("Error stating generated SDK path", zap.String("language", language), zap.String("path", generatedSDKPath), zap.Error(statErr))
		return "", fmt.Errorf("error stating generated path %s for %s: %w", generatedSDKPath, language, statErr)
	} else if !fi.IsDir() {
		s.logger.Warn("Generated SDK path is not a directory. This might be okay for some generators/languages.",
			zap.String("language", language),
			zap.String("path", generatedSDKPath),
		)
	}

	// Check if the directory contains generated files
	if fi.IsDir() {
		entries, err := os.ReadDir(generatedSDKPath)
		if err != nil {
			return "", fmt.Errorf("failed to read generated SDK directory: %w", err)
		}
		if len(entries) == 0 {
			return "", fmt.Errorf("SDK generation completed but no files were generated for language %s", language)
		}
	}

	zipFileName := fmt.Sprintf("%s_%s_sdk.zip", collectionID, language)
	zipFilePath := filepath.Join(outputDir, zipFileName)

	s.logger.Info("Zipping generated SDK",
		zap.String("language", language),
		zap.String("sourceDir", generatedSDKPath),
		zap.String("zipFilePath", zipFilePath),
	)

	if err := utils.ZipDirectory(generatedSDKPath, zipFilePath); err != nil {
		s.logger.Error("Failed to zip generated SDK",
			zap.String("language", language),
			zap.String("sourceDir", generatedSDKPath),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to zip %s SDK: %w", language, err)
	}

	s.logger.Info("SDK successfully generated and zipped",
		zap.String("language", language),
		zap.String("zipPath", zipFilePath),
	)

	return zipFilePath, nil
}

// generatePHPSDK generates a PHP SDK using a custom script.
// This method is kept for reference and should align with the current PHP generation approach.
func (s *SDKService) generatePHPSDK(ctx context.Context, openAPISpecPath, outputDir, packageName, tempDir, collectionID string) (string, error) {
	// Validate inputs
	if openAPISpecPath == "" || outputDir == "" || packageName == "" {
		return "", fmt.Errorf("missing required parameters for PHP SDK generation")
	}

	// Validate OpenAPI spec file exists and is readable
	if _, err := os.Stat(openAPISpecPath); err != nil {
		return "", fmt.Errorf("OpenAPI spec file not accessible: %w", err)
	}

	// Create a temporary file for the PHP script
	tempPhpScriptFile, err := os.CreateTemp(tempDir, "generate_php_sdk_*.php")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for PHP script: %w", err)
	}
	defer tempPhpScriptFile.Close()
	defer os.Remove(tempPhpScriptFile.Name())

	// Write the embedded PHP script to the temp file
	scriptContent, err := s.phpGenScript.ReadFile("generate_php_sdk.php")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded PHP script: %w", err)
	}
	if _, err := tempPhpScriptFile.Write(scriptContent); err != nil {
		return "", fmt.Errorf("failed to write embedded PHP script to temp file: %w", err)
	}

	// Create namespace from package name
	namespace := utils.ConvertToPascalCase(packageName)
	if namespace == "" {
		namespace = "GeneratedSDK"
	}

	// Prepare the command to run the PHP script with positional arguments
	// The PHP script expects: openApiSpecPath, outputDir, namespace, packageName
	args := []string{
		tempPhpScriptFile.Name(),
		openAPISpecPath,
		outputDir,
		namespace,
		packageName,
	}

	// Add timeout context for PHP execution
	phpCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(phpCtx, "php", args...)
	s.logger.Info("Executing PHP SDK generation command",
		zap.String("command", cmd.String()),
		zap.String("namespace", namespace),
		zap.String("packageName", packageName))

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		s.logger.Error("PHP SDK generation command failed",
			zap.Error(cmdErr),
			zap.String("output", string(output)),
			zap.String("namespace", namespace))

		// Check for specific error types
		if phpCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("PHP SDK generation timed out after 10 minutes")
		}
		return "", fmt.Errorf("failed to generate PHP SDK: %w. Output: %s", cmdErr, string(output))
	}

	s.logger.Info("PHP SDK generation command completed successfully",
		zap.String("output", string(output)),
		zap.String("namespace", namespace))

	// Verify the output directory was created and contains files
	if _, err := os.Stat(outputDir); err != nil {
		return "", fmt.Errorf("PHP SDK generation completed but output directory not found: %w", err)
	}

	// Check if the output directory contains generated files
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read generated PHP SDK directory: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("PHP SDK generation completed but no files were generated")
	}

	// Zip the generated SDK directory
	zipFileName := fmt.Sprintf("%s_php_sdk.zip", collectionID)
	zipFilePath := filepath.Join(outputDir, zipFileName)

	if err := utils.ZipDirectory(outputDir, zipFilePath); err != nil {
		s.logger.Error("Failed to zip generated PHP SDK",
			zap.String("sourceDir", outputDir),
			zap.Error(err))
		return "", fmt.Errorf("failed to zip PHP SDK: %w", err)
	}

	s.logger.Info("PHP SDK successfully generated and zipped",
		zap.String("zipPath", zipFilePath),
		zap.String("namespace", namespace))

	return zipFilePath, nil
}

// GetSDKGenerationStatus retrieves the status of an SDK generation task.
// This method might be redundant if GetSDKByID serves the same purpose for status checking.
func (s *SDKService) GetSDKGenerationStatus(ctx context.Context, sdkIDString string) (*models.SDK, error) {
	sdkID, err := primitive.ObjectIDFromHex(sdkIDString)
	if err != nil {
		return nil, fmt.Errorf("invalid SDK ID format: %w", err)
	}
	return s.sdkRepo.GetByID(ctx, sdkID)
}

// DownloadSDKFile is the older download method, potentially redundant with the interface's DownloadSDK.
// Kept for reference. The interface method `DownloadSDK` returns (*models.SDK, string, error)
// which is then used by the controller to send the file.
func (s *SDKService) DownloadSDKFile(ctx context.Context, sdkIDString string) (io.ReadCloser, string, error) {
	sdkID, err := primitive.ObjectIDFromHex(sdkIDString)
	if err != nil {
		return nil, "", fmt.Errorf("invalid SDK ID format: %w", err)
	}
	record, err := s.sdkRepo.GetByID(ctx, sdkID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get SDK record %s: %w", sdkIDString, err)
	}

	if record.Status != models.SDKStatusCompleted {
		return nil, "", fmt.Errorf("SDK %s is not yet completed. Status: %s", sdkIDString, record.Status)
	}

	if record.FilePath == "" {
		return nil, "", fmt.Errorf("SDK %s completed but has no file path associated", sdkIDString)
	}

	file, err := os.Open(record.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open SDK file %s: %w", record.FilePath, err)
	}

	return file, filepath.Base(record.FilePath), nil
}

// DeleteGeneratedSDKsForCollection removes all SDKs associated with a collectionID.
func (s *SDKService) DeleteGeneratedSDKsForCollection(ctx context.Context, collectionIDString string) error {
	sdks, err := s.sdkRepo.GetByCollectionID(ctx, collectionIDString)
	if err != nil {
		return fmt.Errorf("failed to get SDKs for collection %s: %w", collectionIDString, err)
	}

	var firstError error
	for _, sdk := range sdks {
		if sdk.FilePath != "" {
			if err := os.Remove(sdk.FilePath); err != nil {
				s.logger.Error("Failed to delete SDK file", zap.String("path", sdk.FilePath), zap.Error(err))
				if firstError == nil {
					firstError = fmt.Errorf("failed to delete file %s: %w", sdk.FilePath, err)
				}
			}
		}
		if err := s.sdkRepo.SoftDelete(ctx, sdk.ID, sdk.UserID); err != nil {
			s.logger.Error("Failed to soft delete SDK record", zap.String("sdkID", sdk.ID.Hex()), zap.Error(err))
			if firstError == nil {
				firstError = fmt.Errorf("failed to soft delete SDK record %s: %w", sdk.ID.Hex(), err)
			}
		}
	}
	return firstError
}

// GetSDKHistory retrieves the generation history for a user.
func (s *SDKService) GetSDKHistory(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error) {
	s.logger.Info("Fetching SDK history for user",
		zap.String("userID", userID),
		zap.Int("page", page),
		zap.Int("limit", limit))

	return s.sdkRepo.GetByUserID(ctx, userID, page, limit)
}

// UpdateSDKRecordStatus updates the status and other details of an SDK record.
func (s *SDKService) UpdateSDKRecordStatus(ctx context.Context, sdkID string, status models.SDKGenerationStatus, generationTimeMillis int64, failureReason string) error {
	s.logger.Info("Updating SDK record status (direct call)",
		zap.String("sdkID", sdkID),
		zap.String("status", string(status)),
		zap.Int64("generationTimeMillis", generationTimeMillis),
		zap.String("failureReason", failureReason),
	)

	if sdkID == "" {
		s.logger.Warn("UpdateSDKRecordStatus called with empty sdkID (direct call)")
		return fmt.Errorf("sdkID cannot be empty when updating status")
	}

	objID, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		s.logger.Error("Invalid sdkID format for status update (direct call)", zap.String("sdkID", sdkID), zap.Error(err))
		return fmt.Errorf("invalid sdkID format '%s': %w", sdkID, err)
	}

	updateFields := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}

	if generationTimeMillis > 0 {
		updateFields["generationTimeMs"] = generationTimeMillis
	}

	if status == models.SDKStatusFailed && failureReason != "" {
		updateFields["errorMessage"] = failureReason
	}

	if status == models.SDKStatusCompleted {
		updateFields["errorMessage"] = ""       // Clear error message on success
		updateFields["finishedAt"] = time.Now() // Set finishedAt on completion
	}

	return s.sdkRepo.UpdateFields(ctx, objID, updateFields)
}

// GetTotalGeneratedSDKsCount returns the total number of generated SDKs (not soft-deleted)
func (s *SDKService) GetTotalGeneratedSDKsCount(ctx context.Context) (int64, error) {
	filter := bson.M{"isDeleted": bson.M{"$ne": true}}
	count, err := s.sdkRepo.Collection().CountDocuments(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to count generated SDKs", zap.Error(err))
		return 0, err
	}
	return count, nil
}

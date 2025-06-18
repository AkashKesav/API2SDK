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
	GetSDKsByUserID(ctx context.Context, userID string, page, limit int, statusFilter, typeFilter string) ([]*models.SDK, int64, error)

	// UpdateSDKStatus updates the status and an optional error message of an SDK record.
	UpdateSDKStatus(ctx context.Context, sdkID primitive.ObjectID, status models.SDKGenerationStatus, errorMessage string) error // Corrected to SDKGenerationStatus

	// UpdateSDKRecord updates an SDK record with new information, typically after successful generation.
	UpdateSDKRecord(ctx context.Context, sdk *models.SDK) error

	// DownloadSDK retrieves SDK metadata and file path for download, verifying ownership and status.
	DownloadSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) (*models.SDK, string, error) // Changed return to include *models.SDK

	// ConvertPostmanToOpenAPI converts a Postman collection JSON to OpenAPI v3 JSON.
	ConvertPostmanToOpenAPI(ctx context.Context, postmanCollectionJSON string) (string, error)

	// GetPyGenScript returns the embedded Python generation script filesystem.
	GetPyGenScript() embed.FS

	// GetPhpGenScript returns the embedded PHP generation script filesystem.
	GetPhpGenScript() embed.FS

	// GetPhpVendorZip returns the embedded PHP vendor zip filesystem.
	GetPhpVendorZip() embed.FS

	// DeleteSDK soft deletes an SDK record, verifying ownership.
	DeleteSDK(ctx context.Context, sdkID primitive.ObjectID, userID string) error
}

//go:embed jslibs/dist/bundle.js
var jsBundle []byte

//go:embed pylibs/generate_python_sdk.py
var pyGenScript embed.FS

//go:embed phplibs/generate_php_sdk.php
var phpGenScript embed.FS

//go:embed phplibs/vendor.tar.gz
var phpVendorZip embed.FS

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
	mongoClient     *mongo.Client
	dbName          string
	postmanClient   *PostmanClient
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
	mongoClient *mongo.Client,
	dbName string,
	postmanClient *PostmanClient,
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
		mongoClient:     mongoClient,
		dbName:          dbName,
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
func (s *SDKService) GetSDKsByUserID(ctx context.Context, userID string, page, limit int, statusFilter, typeFilter string) ([]*models.SDK, int64, error) {
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
	s.logger.Info("Starting SDK generation process in service",
		zap.String("recordID", recordID.Hex()),
		zap.String("collectionID", genReq.CollectionID),
		zap.String("language", genReq.Language), // Corrected: SDKGenerationRequest has Language directly
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
	sdkRecord.Status = models.SDKStatusInProgress // Corrected: Use SDKStatusInProgress from models
	sdkRecord.UpdatedAt = time.Now()
	if err := s.sdkRepo.Update(ctx, sdkRecord); err != nil {
		s.logger.Error("Failed to update SDK status to InProgress", zap.String("recordID", recordID.Hex()), zap.Error(err))
		// Continue generation, but log the error. The final status update will hopefully succeed.
	}

	// Placeholder for actual generation logic.
	// This would involve:
	// 1. Fetching the collection data (e.g., Postman JSON) using genReq.CollectionID.
	//    This might involve another service call, e.g., a CollectionService.
	//    For now, we'll assume we get a Postman JSON string from somewhere.
	//    Example: postmanJSON, err := s.collectionService.GetPostmanJSON(ctx, genReq.CollectionID)
	//    if err != nil { return sdkRecord, err } // return sdkRecord so status can be updated to failed

	// 2. Converting Postman to OpenAPI if necessary (using s.ConvertPostmanToOpenAPI).
	//    openAPIStr, err := s.ConvertPostmanToOpenAPI(ctx, postmanJSON)
	//    if err != nil { sdkRecord.Status = models.SDKStatusFailed; sdkRecord.ErrorMessage = err.Error(); s.sdkRepo.Update(ctx, sdkRecord); return sdkRecord, err }

	// 3. Creating a temporary directory for generation.
	tempGenDir, err := utils.CreateTempDirForSDK(s.tempDirRootBase, recordID.Hex()) // Corrected: Use utils.CreateTempDirForSDK
	if err != nil {
		s.logger.Error("Failed to create temp directory for SDK generation", zap.Error(err), zap.String("recordID", recordID.Hex()))
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create temp dir: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord) // Attempt to update status
		return sdkRecord, err
	}
	defer os.RemoveAll(tempGenDir) // Clean up temp directory

	// 4. Writing the OpenAPI spec to a file in tempGenDir.
	//    openAPIFilePath := filepath.Join(tempGenDir, "openapi.json")
	//    if err := os.WriteFile(openAPIFilePath, []byte(openAPIStr), 0644); err != nil { ... }

	// 5. Invoking the appropriate generation tool (OpenAPI Generator, custom script) based on genReq.Language.
	//    This is the complex part. Example for a hypothetical Python generator:
	//    generatedSDKPath, err := s.generatePythonSDK(ctx, openAPIFilePath, tempGenDir, genReq.PackageName)
	//    if err != nil { sdkRecord.Status = models.SDKStatusFailed; sdkRecord.ErrorMessage = err.Error(); s.sdkRepo.Update(ctx, sdkRecord); return sdkRecord, err }

	// For now, simulate generation success after a delay
	s.logger.Info("Simulating SDK generation...", zap.String("recordID", recordID.Hex()), zap.String("language", genReq.Language)) // Corrected
	time.Sleep(5 * time.Second) // Simulate work

	// Assume generation was successful and produced a file.
	// The actual file path would be determined by the generation process.
	// It should be a persistent path, not in tempGenDir, or copied from tempGenDir.
	// For this example, let's assume it's stored in a structured way.
	finalSDKPath := filepath.Join("generated_sdks", recordID.Hex(), fmt.Sprintf("%s_sdk.zip", genReq.Language)) // Corrected

	// Create dummy SDK file for simulation
	err = os.MkdirAll(filepath.Dir(finalSDKPath), 0755)
	if err != nil {
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create dir for dummy SDK: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, err
	}
	dummyFile, err := os.Create(finalSDKPath)
	if err != nil {
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create dummy SDK: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, err
	}
	dummyFile.WriteString("This is a dummy SDK for " + genReq.Language) // Corrected
	dummyFile.Close()

	sdkRecord.FilePath = finalSDKPath
	sdkRecord.Status = models.SDKStatusCompleted
	sdkRecord.ErrorMessage = "" // Clear any previous error message
	sdkRecord.FinishedAt = time.Now()
	sdkRecord.UpdatedAt = time.Now()

	s.logger.Info("SDK generation completed successfully (simulated)",
		zap.String("recordID", recordID.Hex()),
		zap.String("filePath", sdkRecord.FilePath),
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

	sdkRecord.Status = models.SDKStatusInProgress // Corrected: Use SDKStatusInProgress from models
	sdkRecord.UpdatedAt = time.Now()
	if err := s.sdkRepo.Update(ctx, sdkRecord); err != nil {
		s.logger.Error("Failed to update SDK status to InProgress for MCP", zap.String("recordID", recordID.Hex()), zap.Error(err))
		// Continue generation, log error.
	}

	// Placeholder for actual MCP generation logic.
	s.logger.Info("Simulating MCP generation...", zap.String("recordID", recordID.Hex()))
	time.Sleep(5 * time.Second) // Simulate work

	finalMCPPath := filepath.Join("generated_mcps", recordID.Hex(), "mcp_server.zip")

	err = os.MkdirAll(filepath.Dir(finalMCPPath), 0755)
	if err != nil {
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create dir for dummy MCP: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, err
	}
	dummyFile, err := os.Create(finalMCPPath)
	if err != nil {
		sdkRecord.Status = models.SDKStatusFailed
		sdkRecord.ErrorMessage = fmt.Sprintf("Failed to create dummy MCP: %s", err.Error())
		s.sdkRepo.Update(ctx, sdkRecord)
		return sdkRecord, err
	}
	dummyFile.WriteString("This is a dummy MCP server")
	dummyFile.Close()

	sdkRecord.FilePath = finalMCPPath
	sdkRecord.Status = models.SDKStatusCompleted
	sdkRecord.ErrorMessage = ""
	sdkRecord.FinishedAt = time.Now()
	sdkRecord.UpdatedAt = time.Now()

	s.logger.Info("MCP generation completed successfully (simulated)",
		zap.String("recordID", recordID.Hex()),
		zap.String("filePath", sdkRecord.FilePath),
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

// ConvertPostmanToOpenAPI converts a Postman collection JSON string to an OpenAPI v3 string.
// It uses an embedded JavaScript bundle (Postman SDK) to perform the conversion.
func (s *SDKService) ConvertPostmanToOpenAPI(ctx context.Context, postmanCollectionJSON string) (string, error) {
	s.logger.Info("Starting Postman to OpenAPI conversion")

	if len(jsBundle) == 0 {
		s.logger.Error("JavaScript bundle (jsBundle) is nil or empty. This indicates an issue with embedding or loading the bundle.js file.")
		return "", fmt.Errorf("critical: JavaScript bundle (jsBundle) not loaded or empty")
	}

	vm := goja.New()
	polyfillConsole(vm, s.logger)

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
	optionsMap := map[string]string{"outputFormat": "json"}
	optionsBytes, marshalErr := json.Marshal(optionsMap) // Use standard json.Marshal
	if marshalErr != nil {
		s.logger.Error("Failed to marshal options for JS call", zap.Error(marshalErr))
		return "", fmt.Errorf("failed to marshal options for JS call: %w", marshalErr)
	}
	optionsJSONString := string(optionsBytes)

	jsPromiseValue, err := convertFunc(goja.Undefined(), vm.ToValue(postmanCollectionJSON), vm.ToValue(optionsJSONString))
	if err != nil {
		s.logger.Error("JavaScript function call failed", zap.Error(err))
		return "", fmt.Errorf("JavaScript function call failed: %w", err)
	}

	if jsPromiseValue == nil || goja.IsUndefined(jsPromiseValue) || goja.IsNull(jsPromiseValue) {
		s.logger.Error("JavaScript function returned nil or undefined, expected a Promise.")
		return "", fmt.Errorf("JavaScript function returned nil or undefined")
	}

	// Validate that jsPromiseValue looks like a promise.
	tempPromiseObj := jsPromiseValue.ToObject(vm)
	if tempPromiseObj == nil {
		s.logger.Error("JavaScript function result cannot be converted to object, expected Promise-like.", zap.String("type", jsPromiseValue.ExportType().String()))
		return "", fmt.Errorf("JavaScript function result is not an object, type: %s", jsPromiseValue.ExportType().String())
	}
	thenVal := tempPromiseObj.Get("then")
	_, isThenCallable := goja.AssertFunction(thenVal)
	if thenVal == nil || goja.IsUndefined(thenVal) || goja.IsNull(thenVal) || !isThenCallable {
		s.logger.Error("JavaScript function result does not have a callable 'then' method, not a Promise.", zap.Any("then_type", thenVal.ExportType().String()))
		return "", fmt.Errorf("JavaScript function result is not a Promise (no callable 'then' method)")
	}

	s.logger.Debug("JavaScript function call successful, proceeding to wait for promise.")
	select {
	case result := <-s.waitForPromise(vm, jsPromiseValue): // Pass jsPromiseValue (goja.Value)
		s.logger.Debug("JavaScript Promise resolved or rejected channel received", zap.Any("raw_js_result_struct", result))

		if result.Error != nil {
			s.logger.Error("JavaScript Promise rejected", zap.Error(result.Error))
			return "", fmt.Errorf("JavaScript Promise rejected: %w", result.Error)
		}

		if result.Value == nil || goja.IsUndefined(result.Value) || goja.IsNull(result.Value) {
			s.logger.Error("JavaScript Promise resolved with nil, undefined, or null result")
			return "", fmt.Errorf("JavaScript Promise resolved with nil, undefined, or null")
		}

		exportedVal := result.Value.Export()
		if exportedVal == nil {
			s.logger.Error("Exported JavaScript result is nil after conversion from goja.Value")
			return "", fmt.Errorf("exported JavaScript result is nil")
		}

		openAPISpec, ok := exportedVal.(string)
		if !ok {
			s.logger.Error("JavaScript Promise result is not a string",
				zap.String("actual_type", fmt.Sprintf("%T", exportedVal)),
				zap.Any("actual_value", exportedVal))
			return "", fmt.Errorf("JavaScript Promise result is not a string, got type %T", exportedVal)
		}

		if openAPISpec == "" {
			s.logger.Warn("JavaScript conversion returned an empty string. This might be a valid empty OpenAPI spec or indicate an issue with the input Postman collection or the conversion logic.")
		}

		s.logger.Info("Postman to OpenAPI conversion successful")
		return openAPISpec, nil

	case <-ctx.Done():
		s.logger.Error("Context cancelled while waiting for JavaScript Promise")
		return "", fmt.Errorf("context cancelled while waiting for JavaScript Promise: %w", ctx.Err())
		// Add a timeout to prevent indefinite blocking, if desired.
		// case <-time.After(30 * time.Second):
		// s.logger.Error("Timeout waiting for JavaScript Promise")
		// return "", fmt.Errorf("timeout waiting for JavaScript Promise")
	}
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

// GenerateMCPServer generates an MCP server project from an OpenAPI specification string.
// This is a separate method that was previously part of the monolithic GenerateSDK.
// It's kept for potential direct use but the main flow uses GenerateMCP via the interface.
func (s *SDKService) GenerateMCPServer(ctx context.Context, userID, collectionID, openAPISpecContent, outputDir, transport string, port int) (string, string, error) {
	startTime := time.Now()
	s.logger.Info("Starting MCP server generation (direct call)",
		zap.String("userID", userID),
		zap.String("collectionID", collectionID),
		zap.String("outputDir", outputDir),
		zap.String("transport", transport),
		zap.Int("port", port),
	)

	// Record the start of MCP generation
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
		s.logger.Error("Failed to create MCP record before generation (direct call)", zap.Error(err))
		return "", "", fmt.Errorf("failed to create MCP record: %w", err) // Return empty sdkID if create fails
	} else {
		mcpRecord = createdRecord
		s.logger.Info("MCP record created (direct call)", zap.String("sdkID", mcpRecord.ID.Hex()))
	}

	// Create a temporary file for the OpenAPI spec content
	tempSpecFile, err := os.CreateTemp(s.tempDirRootBase, "openapi_spec_*.json")
	if err != nil {
		s.logger.Error("Failed to create temporary file for OpenAPI spec (direct call)", zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to create temporary spec file: %s", err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to create temporary spec file: %w", err)
	}
	defer tempSpecFile.Close()
	defer os.Remove(tempSpecFile.Name()) // Clean up the temp spec file

	if _, err := tempSpecFile.WriteString(openAPISpecContent); err != nil {
		s.logger.Error("Failed to write OpenAPI spec to temporary file (direct call)", zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to write spec to temp file: %s", err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to write spec to temp file: %w", err)
	}
	openAPISpecPath := tempSpecFile.Name()
	s.logger.Info("OpenAPI spec written to temporary file (direct call)", zap.String("path", openAPISpecPath))

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		s.logger.Error("Failed to create output directory for MCP server (direct call)", zap.String("outputDir", outputDir), zap.Error(err))
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("failed to create output directory %s: %s", outputDir, err.Error()))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Prepare the openapi-mcp-generator command
	args := []string{
		"--input", openAPISpecPath,
		"--output", outputDir,
		"--transport", transport,
		"--force", // Overwrite existing files in the output directory
	}
	if transport == "web" || transport == "streamable-http" {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}

	cmd := exec.CommandContext(ctx, "openapi-mcp-generator", args...)
	s.logger.Info("Executing openapi-mcp-generator command (direct call)", zap.String("command", cmd.String()))

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		s.logger.Error("openapi-mcp-generator command failed (direct call)",
			zap.Error(cmdErr),
			zap.String("output", string(output)),
		)
		s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusFailed, time.Since(startTime).Milliseconds(), fmt.Sprintf("generation command failed for MCP: %s. Output: %s", cmdErr, string(output)))
		return "", mcpRecord.ID.Hex(), fmt.Errorf("failed to generate MCP server: %w. Output: %s", cmdErr, string(output))
	}

	s.logger.Info("openapi-mcp-generator command completed successfully (direct call)", zap.String("output", string(output)))

	s.UpdateSDKRecordStatus(ctx, mcpRecord.ID.Hex(), models.SDKStatusCompleted, time.Since(startTime).Milliseconds(), "") // Empty error string for success

	return outputDir, mcpRecord.ID.Hex(), nil
}

// GenerateSDKWithDetails is the older, more detailed SDK generation function.
// It's kept for reference or if a more granular, non-interface-driven call is needed internally.
// The primary SDK generation path should use the GenerateSDK method defined in the SDKServiceInterface.
func (s *SDKService) GenerateSDKWithDetails(ctx context.Context, userID, collectionID, openAPISpecPath, language, outputDir, targetPackageName string) (string, string, error) {
	startTime := time.Now()
	tempDir, err := os.MkdirTemp(s.tempDirRootBase, "sdkgen_*")
	if err != nil {
		s.logger.Error("Failed to create temporary directory for SDK generation (details)", zap.Error(err))
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	// defer os.RemoveAll(tempDir) // Clean up: Defer removal of the temporary directory for this task. Commented out for debugging.
	s.logger.Info("Temporary directory for SDK generation created (details)", zap.String("path", tempDir))

	s.logger.Info("Generating SDK (details)",
		zap.String("language", language),
		zap.String("openAPISpecPath", openAPISpecPath),
		zap.String("outputDir", outputDir),
		zap.String("packageName", targetPackageName),
		zap.String("collectionID", collectionID),
		zap.String("userID", userID),
	)

	sdkRecord := &models.SDK{
		UserID:         userID,
		CollectionID:   collectionID,
		GenerationType: models.GenerationTypeSDK,
		Language:       language,
		PackageName:    targetPackageName,
		Status:         models.SDKStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	createdRecord, err := s.sdkRepo.Create(ctx, sdkRecord)
	if err != nil {
		s.logger.Error("Failed to create SDK record before generation (details)", zap.Error(err))
		return "", "", fmt.Errorf("failed to create SDK record: %w", err)
	} else {
		sdkRecord = createdRecord
		s.logger.Info("SDK record created (details)", zap.String("sdkID", sdkRecord.ID.Hex()))
	}

	var generatedSDKPath string
	var genErr error

	switch language {
	case "go", "typescript", "python", "java", "csharp", "rust", "ruby":
		generatedSDKPath, genErr = s.generateWithOpenAPIGenerator(ctx, openAPISpecPath, language, outputDir, targetPackageName, tempDir, collectionID)
	case "php":
		generatedSDKPath, genErr = s.generatePHPSDK(ctx, openAPISpecPath, outputDir, targetPackageName, tempDir, collectionID)
	default:
		genErr = fmt.Errorf("unsupported language: %s", language)
	}

	generationDuration := time.Since(startTime)

	if genErr != nil {
		s.logger.Error("SDK generation failed (details)", zap.String("language", language), zap.Error(genErr))
		s.UpdateSDKRecordStatus(ctx, createdRecord.ID.Hex(), models.SDKStatusFailed, generationDuration.Milliseconds(), genErr.Error())
		return "", createdRecord.ID.Hex(), fmt.Errorf("generation failed for %s: %w", language, genErr)
	}

	s.logger.Info("SDK generated successfully (details)",
		zap.String("language", language),
		zap.String("path", generatedSDKPath))

	s.UpdateSDKRecordStatus(ctx, createdRecord.ID.Hex(), models.SDKStatusCompleted, generationDuration.Milliseconds(), "")

	return generatedSDKPath, createdRecord.ID.Hex(), nil
}

// generateWithOpenAPIGenerator uses the OpenAPI Generator CLI to generate SDKs.
// It abstracts the command execution and error handling for SDK generation.
func (s *SDKService) generateWithOpenAPIGenerator(ctx context.Context, openAPISpecPath, language, outputDir, packageName, tempDir, collectionID string) (string, error) { // Added collectionID
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

	switch language {
	case "go":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "go",
			"-o", outputDir,
			"--package-name", packageName,
			"--git-user-id", "YourGitHubUser",
			"--git-repo-id", fmt.Sprintf("%s-go", packageName),
		)
		generatedSDKPath = outputDir

	case "typescript", "typescript-axios":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "typescript-axios",
			"-o", outputDir,
			"--additional-properties=npmName="+packageName+",supportsES6=true,usePromises=true",
		)
		generatedSDKPath = outputDir

	case "python":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "python",
			"-o", outputDir,
			"--additional-properties", fmt.Sprintf("packageName=%s,projectName=%s,packageVersion=1.0.0", packageName, packageName),
		)
		generatedSDKPath = outputDir

	case "php":
		pascalCasePackageName := utils.ConvertToPascalCase(packageName)
		composerProjectName := utils.ConvertToSnakeCase(packageName)
		composerVendorName := "myvendor"

		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "php",
			"-o", outputDir,
			"--additional-properties",
			fmt.Sprintf("composerVendorName=%s,composerProjectName=%s,invokerPackage=%s,variableNamingConvention=camelCase",
				composerVendorName, composerProjectName, pascalCasePackageName),
		)
		generatedSDKPath = outputDir

	case "java":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "java",
			"-o", outputDir,
			"--artifact-id", packageName,
			"--group-id", "org.example.api",
			"--api-package", fmt.Sprintf("org.example.api.%s.api", packageName),
			"--model-package", fmt.Sprintf("org.example.api.%s.model", packageName),
			"--library", "native",
		)
		generatedSDKPath = outputDir

	case "csharp", "csharp-netcore":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "csharp",
			"-o", outputDir,
			"--package-name", packageName,
			"--additional-properties", "targetFramework=net6.0,packageName="+packageName,
		)
		generatedSDKPath = outputDir

	case "rust":
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "rust",
			"-o", outputDir,
			"--package-name", packageName,
		)
		generatedSDKPath = outputDir

	case "ruby":
		pascalCasePackageName := utils.ConvertToPascalCase(packageName)
		cmd = exec.Command("java", "-jar", generatorJarPath, "generate",
			"-i", openAPISpecPath,
			"-g", "ruby",
			"-o", outputDir,
			"--additional-properties",
			fmt.Sprintf("moduleName=%s,gemName=%s,gemVersion=1.0.0", pascalCasePackageName, packageName),
		)
		generatedSDKPath = outputDir

	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	s.logger.Info("Executing SDK generation command (details)", zap.String("language", language), zap.String("command", cmd.String()))
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		s.logger.Error("SDK generation command failed (details)",
			zap.String("language", language),
			zap.Error(cmdErr),
			zap.String("output", string(output)),
		)
		return "", fmt.Errorf("generation command failed for %s: %s. Output: %s", language, cmdErr, string(output))
	}

	s.logger.Info("SDK generation command completed successfully (details)",
		zap.String("language", language),
		zap.String("output", string(output)),
	)

	fi, statErr := os.Stat(generatedSDKPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			s.logger.Error("Generated SDK path does not exist after generation (details)",
				zap.String("language", language),
				zap.String("expectedPath", generatedSDKPath),
				zap.String("generatorOutput", string(output)),
			)
			return "", fmt.Errorf("generated path %s does not exist for language %s. Output: %s", generatedSDKPath, language, string(output))
		}
		s.logger.Error("Error stating generated SDK path (details)", zap.String("language", language), zap.String("path", generatedSDKPath), zap.Error(statErr))
		return "", fmt.Errorf("error stating generated path %s for %s: %w", generatedSDKPath, language, statErr)
	} else if !fi.IsDir() {
		s.logger.Warn("Generated SDK path is not a directory (details). This might be okay for some generators/languages.",
			zap.String("language", language),
			zap.String("path", generatedSDKPath),
		)
	}

	zipFileName := fmt.Sprintf("%s_%s_sdk.zip", collectionID, language)
	zipFilePath := filepath.Join(outputDir, zipFileName)

	s.logger.Info("Zipping generated SDK (details)",
		zap.String("language", language),
		zap.String("sourceDir", generatedSDKPath),
		zap.String("zipFilePath", zipFilePath),
	)

	if err := utils.ZipDirectory(generatedSDKPath, zipFilePath); err != nil {
		s.logger.Error("Failed to zip generated SDK (details)",
			zap.String("language", language),
			zap.String("sourceDir", generatedSDKPath),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to zip %s SDK: %w", language, err)
	}

	s.logger.Info("SDK successfully generated and zipped (details)",
		zap.String("language", language),
		zap.String("zipPath", zipFilePath),
	)

	return zipFilePath, nil
}

// generatePHPSDK generates a PHP SDK using a custom script.
// This method is kept for reference and should align with the current PHP generation approach.
func (s *SDKService) generatePHPSDK(ctx context.Context, openAPISpecPath, outputDir, packageName, tempDir, collectionID string) (string, error) { // Added collectionID
	s.logger.Info("Generating PHP SDK with custom script",
		zap.String("openAPISpecPath", openAPISpecPath),
		zap.String("outputDir", outputDir),
		zap.String("packageName", packageName),
	)

	// Create a temporary file for the Python script
	tempPyScriptFile, err := os.CreateTemp(tempDir, "generate_python_sdk_*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for PHP script: %w", err)
	}
	defer tempPyScriptFile.Close()
	defer os.Remove(tempPyScriptFile.Name())

	// Write the embedded Python script to the temp file
	scriptContent, err := s.pyGenScript.ReadFile("generate_python_sdk.py")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded Python script: %w", err)
	}
	if _, err := tempPyScriptFile.Write(scriptContent); err != nil {
		return "", fmt.Errorf("failed to write embedded Python script to temp file: %w", err)
	}

	// Prepare the command to run the Python script
	// Example: python3 /path/to/generate_python_sdk.py --input /path/to/openapi.json --output /path/to/output/dir --package_name=my_package
	args := []string{
		"--input", openAPISpecPath,
		"--output", outputDir,
		"--package_name", packageName,
	}

	cmdArgs := []string{tempPyScriptFile.Name()}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, "python3", cmdArgs...)
	s.logger.Info("Executing Python SDK generation command for PHP (misconfiguration?)", zap.String("command", cmd.String()))

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		s.logger.Error("Python SDK generation command failed for PHP (misconfiguration?)",
			zap.Error(cmdErr),
			zap.String("output", string(output)),
		)
		return "", fmt.Errorf("failed to generate PHP SDK via Python script: %w. Output: %s", cmdErr, string(output))
	}

	s.logger.Info("Python SDK generation command completed successfully for PHP (misconfiguration?)", zap.String("output", string(output)))

	// Zip the generated SDK directory
	zipFileName := fmt.Sprintf("%s_php_sdk.zip", collectionID) // Use collectionID for unique zip name
	zipFilePath := filepath.Join(outputDir, zipFileName)

	s.logger.Info("Zipping generated PHP SDK",
		zap.String("sourceDir", outputDir),
		zap.String("zipFilePath", zipFilePath),
	)

	if err := utils.ZipDirectory(outputDir, zipFilePath); err != nil {
		s.logger.Error("Failed to zip generated PHP SDK",
			zap.String("sourceDir", outputDir),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to zip PHP SDK: %w", err)
	}

	s.logger.Info("PHP SDK successfully generated and zipped",
		zap.String("zipPath", zipFilePath),
	)

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
		if err := s.sdkRepo.SoftDelete(ctx, sdk.ID, sdk.UserID);
			err != nil {
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
		updateFields["errorMessage"] = "" // Clear error message on success
		updateFields["finishedAt"] = time.Now() // Set finishedAt on completion
	}

	return s.sdkRepo.UpdateFields(ctx, objID, updateFields)
}

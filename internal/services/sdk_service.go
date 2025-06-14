package services

import (
	"archive/tar"
	"compress/gzip"
	"context"
	_ "embed" // Corrected embed import
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"github.com/AkashKesav/API2SDK/internal/utils"
	"github.com/dop251/goja"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

//go:embed jslibs/openapi-typescript.min.js
var openapiTypescriptMinJS string // Keep as string for goja

//go:embed jslibs/dist/bundle.js
var jsBundle string // Bundled JS with postman-to-openapi and openapi-typescript

//go:embed pylibs/generate_python_sdk.py
var generatePythonSDKScript string // Keep as string for PyRun_StringUTF8

//go:embed phplibs/vendor.tar.gz
var phpVendorTarGz []byte // Keep as []byte for direct use

//go:embed phplibs/generate_php_sdk.php
var generatePHPSDKScriptBytes []byte // Changed to []byte for os.WriteFile

//go:embed codegeners/openapi-generator-cli.jar
var openapiGeneratorCLIJar []byte // Keep as []byte for os.WriteFile

type SDKService struct {
	logger  *zap.Logger
	sdkRepo *repositories.SDKRepository // Added SDKRepository dependency
}

func NewSDKService(logger *zap.Logger, sdkRepo *repositories.SDKRepository) *SDKService {
	return &SDKService{
		logger:  logger,
		sdkRepo: sdkRepo,
	}
}

// GenerateSDK generates an SDK using a Go-native library or embedded interpreters.
func (s *SDKService) GenerateSDK(
	ctx context.Context,
	userID string,
	collectionID string, // Can be empty if not from a collection
	openAPISpecPath string,
	language string,
	outputDir string, // This is the base output dir for the SDK
	targetPackageName string,
) (string, error) { // Return final SDK path and error

	startTime := time.Now()

	// Create an initial SDK record
	sdkRecord := &models.SDK{
		UserID:       userID,
		CollectionID: collectionID,
		PackageName:  targetPackageName,
		Language:     language,
		Status:       models.SDKStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdRecord, err := s.sdkRepo.Create(ctx, sdkRecord)
	if err != nil {
		s.logger.Error("Failed to create initial SDK record", zap.Error(err))
	}
	if createdRecord != nil {
		sdkRecord = createdRecord // Use the record with DB-assigned ID
	}

	// Defer function to update SDK record status on completion or failure
	var generationError error
	var finalSDKOutputPath string // This will be the path to the generated SDK

	defer func() {
		if sdkRecord.ID.IsZero() { // If initial creation failed and we don't have an ID
			s.logger.Error("cannot update SDK record as it was not created", zap.Error(generationError))
			return
		}

		sdkRecord.GenerationTime = time.Since(startTime).Milliseconds()
		sdkRecord.UpdatedAt = time.Now()

		if generationError != nil {
			sdkRecord.Status = models.SDKStatusFailed
			sdkRecord.ErrorMessage = generationError.Error()
			s.logger.Error("SDK generation failed, updating record",
				zap.String("sdkID", sdkRecord.ID.Hex()),
				zap.Error(generationError))
		} else {
			sdkRecord.Status = models.SDKStatusCompleted
			sdkRecord.FilePath = finalSDKOutputPath // Store the actual path of the generated SDK
			sdkRecord.GeneratedAt = time.Now()
			s.logger.Info("SDK generation successful, updating record",
				zap.String("sdkID", sdkRecord.ID.Hex()),
				zap.String("path", sdkRecord.FilePath))
		}

		if err := s.sdkRepo.Update(ctx, sdkRecord); err != nil {
			s.logger.Error("failed to update SDK record", zap.Error(err), zap.String("sdkID", sdkRecord.ID.Hex()))
		}
	}()

	// Ensure the base output directory for this specific SDK generation exists
	if err := utils.EnsureDir(outputDir); err != nil {
		generationError = fmt.Errorf("failed to ensure output directory %s exists: %w", outputDir, err)
		return "", generationError
	}
	finalSDKOutputPath = outputDir // Initially, the output dir itself is the "path"

	// Go SDK Generation
	if strings.ToLower(language) == "go" {
		goPackagePath := filepath.Join(outputDir, targetPackageName)
		if err := utils.EnsureDir(goPackagePath); err != nil {
			generationError = fmt.Errorf("failed to ensure Go package directory '%s' exists: %w", goPackagePath, err)
			return "", generationError
		}
		s.logger.Info("Generating Go SDK", zap.String("spec", openAPISpecPath), zap.String("path", goPackagePath), zap.String("pkg", targetPackageName))

		// Load the OpenAPI specification
		spec, err := util.LoadSwagger(openAPISpecPath)
		if err != nil {
			generationError = fmt.Errorf("error loading swagger spec '%s': %w", openAPISpecPath, err)
			return "", generationError
		}

		outputFilePath := filepath.Join(goPackagePath, "client.gen.go")
		opts := codegen.Configuration{
			PackageName: targetPackageName,
			Generate: codegen.GenerateOptions{
				EchoServer:   false,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: codegen.OutputOptions{
				SkipFmt:   false,
				SkipPrune: false,
			},
		}

		// Generate the code
		code, err := codegen.Generate(spec, opts)
		if err != nil {
			generationError = fmt.Errorf("error generating Go client SDK with oapi-codegen: %w", err)
			return "", generationError
		}

		// Write the generated code to the output file
		if err := os.WriteFile(outputFilePath, []byte(code), 0644); err != nil {
			generationError = fmt.Errorf("error writing generated Go client SDK to file '%s': %w", outputFilePath, err)
			return "", generationError
		}

		finalSDKOutputPath = outputFilePath // Specific file for Go
		s.logger.Info("Go SDK generation complete", zap.String("output", outputFilePath))
		return finalSDKOutputPath, nil
	}

	// TypeScript SDK Generation
	if strings.ToLower(language) == "typescript" {
		tsPackagePath := filepath.Join(outputDir, targetPackageName)
		if err := utils.EnsureDir(tsPackagePath); err != nil {
			generationError = fmt.Errorf("failed to ensure TypeScript package directory '%s' exists: %w", tsPackagePath, err)
			return "", generationError
		}
		finalSDKOutputPath = tsPackagePath // The directory is the output
		s.logger.Info("Generating TypeScript SDK", zap.String("spec", openAPISpecPath), zap.String("path", tsPackagePath), zap.String("module", targetPackageName))

		openAPISpecBytes, err := os.ReadFile(openAPISpecPath)
		if err != nil {
			generationError = fmt.Errorf("failed to read OpenAPI spec file '%s': %w", openAPISpecPath, err)
			return "", generationError
		}
		openAPISpecString := string(openAPISpecBytes)

		vm := goja.New()
		_, err = vm.RunString(openapiTypescriptMinJS)
		if err != nil {
			generationError = fmt.Errorf("failed to execute JavaScript bundle: %w", err)
			return "", generationError
		}
		generateFunc, ok := goja.AssertFunction(vm.Get("generateTypescriptClient"))
		if !ok {
			generationError = fmt.Errorf("'generateTypescriptClient' function not found in JavaScript bundle")
			return "", generationError
		}

		result, err := generateFunc(goja.Undefined(), vm.ToValue(openAPISpecString), vm.ToValue("{}"))
		if err != nil {
			generationError = fmt.Errorf("JavaScript SDK generation failed: %w", err)
			return "", generationError
		}
		generatedTSCode := result.String()
		outputTSFilePath := filepath.Join(tsPackagePath, "index.ts")
		if err := os.WriteFile(outputTSFilePath, []byte(generatedTSCode), 0644); err != nil {
			generationError = fmt.Errorf("failed to write generated TypeScript SDK to file '%s': %w", outputTSFilePath, err)
			return "", generationError
		}

		s.logger.Info("TypeScript SDK generation complete", zap.String("output", finalSDKOutputPath))
		return finalSDKOutputPath, nil
	}

	// Python SDK Generation
	if strings.ToLower(language) == "python" {
		pythonProjectParentDir := outputDir
		pythonProjectName := targetPackageName
		finalSDKOutputPath = filepath.Join(pythonProjectParentDir, pythonProjectName)

		s.logger.Info("Generating Python SDK", zap.String("spec", openAPISpecPath), zap.String("parentDir", pythonProjectParentDir), zap.String("project", pythonProjectName))

		// Create a temporary directory to place the Python script
		tempDir, err := os.MkdirTemp("", "python-sdk-gen-")
		if err != nil {
			generationError = fmt.Errorf("failed to create temporary directory for Python generation: %w", err)
			return "", generationError
		}
		defer os.RemoveAll(tempDir) // Clean up temp directory

		// Write the embedded Python generator script to the temp directory
		tempScriptPath := filepath.Join(tempDir, "generate_python_sdk.py")
		if err := os.WriteFile(tempScriptPath, []byte(generatePythonSDKScript), 0755); err != nil {
			generationError = fmt.Errorf("failed to write embedded Python script to temp file '%s': %w", tempScriptPath, err)
			return "", generationError
		}

		// Determine Python executable (python3 or python)
		pythonExecutable := "python3"
		if _, err := exec.LookPath(pythonExecutable); err != nil {
			s.logger.Warn("python3 not found, trying python", zap.Error(err))
			pythonExecutable = "python"
			if _, err := exec.LookPath(pythonExecutable); err != nil {
				generationError = fmt.Errorf("neither python3 nor python found in PATH: %w", err)
				return "", generationError
			}
		}

		// Execute the Python script
		cmd := exec.CommandContext(ctx, pythonExecutable, tempScriptPath, openAPISpecPath, pythonProjectParentDir, pythonProjectName)
		cmd.Dir = tempDir

		s.logger.Info("Executing Python SDK generation command",
			zap.String("command", strings.Join(cmd.Args, " ")),
			zap.String("workingDir", cmd.Dir))

		output, err := cmd.CombinedOutput()
		if err != nil {
			s.logger.Error("Python SDK generation script failed",
				zap.Error(err),
				zap.String("output", string(output)),
				zap.String("specPath", openAPISpecPath),
				zap.String("parentDir", pythonProjectParentDir),
				zap.String("packageName", pythonProjectName))
			generationError = fmt.Errorf("python SDK generation script failed: %w. Output:\n%s", err, string(output))
			return "", generationError
		}

		s.logger.Info("Python SDK generation script executed successfully",
			zap.String("output", string(output)),
			zap.String("specPath", openAPISpecPath),
			zap.String("parentDir", pythonProjectParentDir),
			zap.String("packageName", pythonProjectName))

		return finalSDKOutputPath, nil
	}

	// PHP SDK Generation
	if strings.ToLower(language) == "php" {
		phpOutputDir := filepath.Join(outputDir, targetPackageName)
		finalSDKOutputPath = phpOutputDir
		if err := utils.EnsureDir(phpOutputDir); err != nil {
			generationError = fmt.Errorf("failed to ensure PHP output directory '%s' exists: %w", phpOutputDir, err)
			return "", generationError
		}

		phpNamespace := utils.ConvertToPascalCase(targetPackageName)
		composerPackageName := fmt.Sprintf("api2sdk/%s", utils.ConvertToKebabCase(targetPackageName))
		s.logger.Info("Generating PHP SDK", zap.String("spec", openAPISpecPath), zap.String("outputDir", phpOutputDir), zap.String("namespace", phpNamespace))

		// Create a temporary directory to extract/place PHP script and vendor files
		tempDir, err := os.MkdirTemp("", "php-sdk-gen-")
		if err != nil {
			generationError = fmt.Errorf("failed to create temporary directory for PHP generation: %w", err)
			return "", generationError
		}
		defer os.RemoveAll(tempDir) // Clean up temp directory

		// Write the embedded PHP generator script to the temp directory
		tempScriptPath := filepath.Join(tempDir, "generate_php_sdk.php")
		if err := os.WriteFile(tempScriptPath, generatePHPSDKScriptBytes, 0755); err != nil {
			generationError = fmt.Errorf("failed to write embedded PHP script to temp file '%s': %w", tempScriptPath, err)
			return "", generationError
		}

		// Extract the embedded vendor.tar.gz into the temp directory
		vendorDir := filepath.Join(tempDir, "vendor")
		if err := utils.EnsureDir(vendorDir); err != nil {
			generationError = fmt.Errorf("failed to create vendor directory in temp path '%s': %w", vendorDir, err)
			return "", generationError
		}

		gzipReader, err := gzip.NewReader(strings.NewReader(string(phpVendorTarGz)))
		if err != nil {
			generationError = fmt.Errorf("failed to create gzip reader for vendor tarball: %w", err)
			return "", generationError
		}
		defer gzipReader.Close()

		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break // End of tar archive
			}
			if err != nil {
				generationError = fmt.Errorf("failed to read from vendor tarball: %w", err)
				return "", generationError
			}

			targetPath := filepath.Join(tempDir, header.Name) // Extract relative to tempDir

			switch header.Typeflag {
			case tar.TypeDir:
				if err := utils.EnsureDir(targetPath); err != nil {
					generationError = fmt.Errorf("failed to create directory from tarball '%s': %w", targetPath, err)
					return "", generationError
				}
			case tar.TypeReg:
				outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					generationError = fmt.Errorf("failed to create file from tarball '%s': %w", targetPath, err)
					return "", generationError
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					generationError = fmt.Errorf("failed to write file content from tarball '%s': %w", targetPath, err)
					return "", generationError
				}
				outFile.Close()
			default:
				// Skip other types like symlinks for now for simplicity
				s.logger.Info("Skipping unsupported tar entry type", zap.String("type", string(header.Typeflag)), zap.String("name", header.Name))
			}
		}
		s.logger.Info("PHP vendor files extracted", zap.String("to", vendorDir))

		// Execute the PHP script
		cmd := exec.CommandContext(ctx, "php", "-f", tempScriptPath, "--", openAPISpecPath, phpOutputDir, phpNamespace, composerPackageName)
		cmd.Dir = tempDir // Set working directory for the script, so it can find vendor/autoload.php

		output, err := cmd.CombinedOutput()
		if err != nil {
			generationError = fmt.Errorf("PHP SDK generation script failed: %w. Output:\n%s", err, string(output))
			return "", generationError
		}

		s.logger.Info("PHP SDK generation complete", zap.String("output", string(output)))
		return finalSDKOutputPath, nil
	}

	// JAR-based Generators (Java, C#, Rust, Ruby)
	loLanguage := strings.ToLower(language)
	if loLanguage == "java" || loLanguage == "csharp" || loLanguage == "rust" || loLanguage == "ruby" {
		generatorName := ""
		additionalProperties := []string{}
		packageNameProperty := ""
		finalOutputDir := filepath.Join(outputDir, targetPackageName)

		switch loLanguage {
		case "java":
			generatorName = "java"
			javaPackage := fmt.Sprintf("com.%s.%s", utils.ConvertToAlphanumeric(utils.GetOrgFromPkg(targetPackageName), "mypackage"), utils.ConvertToAlphanumeric(utils.GetNameFromPkg(targetPackageName), "client"))
			additionalProperties = append(additionalProperties, fmt.Sprintf("apiPackage=%s.api", javaPackage))
			additionalProperties = append(additionalProperties, fmt.Sprintf("modelPackage=%s.model", javaPackage))
			additionalProperties = append(additionalProperties, fmt.Sprintf("invokerPackage=%s.invoker", javaPackage))
			additionalProperties = append(additionalProperties, fmt.Sprintf("groupId=com.%s", utils.ConvertToAlphanumeric(utils.GetOrgFromPkg(targetPackageName), "mypackage")))
			additionalProperties = append(additionalProperties, fmt.Sprintf("artifactId=%s", utils.ConvertToKebabCase(utils.GetNameFromPkg(targetPackageName))))
			finalOutputDir = outputDir
		case "csharp":
			generatorName = "csharp-netcore"
			csharpNamespace := fmt.Sprintf("%s.%s", utils.ConvertToPascalCase(utils.GetOrgFromPkg(targetPackageName)), utils.ConvertToPascalCase(utils.GetNameFromPkg(targetPackageName)))
			packageNameProperty = fmt.Sprintf("packageName=%s", csharpNamespace)
			finalOutputDir = filepath.Join(outputDir, csharpNamespace)
		case "rust":
			generatorName = "rust"
			rustPackageName := utils.ConvertToSnakeCase(targetPackageName)
			packageNameProperty = fmt.Sprintf("packageName=%s", rustPackageName)
			finalOutputDir = filepath.Join(outputDir, rustPackageName)
		case "ruby":
			generatorName = "ruby"
			gemName := utils.ConvertToSnakeCase(targetPackageName)
			moduleName := utils.ConvertToPascalCase(targetPackageName)
			packageNameProperty = fmt.Sprintf("gemName=%s,moduleName=%s,gemVersion=1.0.0", gemName, moduleName)
			finalOutputDir = filepath.Join(outputDir, gemName)
		}
		finalSDKOutputPath = finalOutputDir

		if err := utils.EnsureDir(finalOutputDir); err != nil {
			generationError = fmt.Errorf("failed to ensure output directory '%s' for %s: %w", finalOutputDir, language, err)
			return "", generationError
		}

		s.logger.Info("Generating SDK with OpenAPI Generator CLI", zap.String("lang", language), zap.String("spec", openAPISpecPath), zap.String("outputDir", finalOutputDir))

		// Create a temporary directory to extract the JAR
		tempDir, err := os.MkdirTemp("", "openapi-generator-cli-")
		if err != nil {
			generationError = fmt.Errorf("failed to create temporary directory for OpenAPI Generator CLI: %w", err)
			return "", generationError
		}
		defer os.RemoveAll(tempDir) // Clean up temp directory

		// Write the embedded JAR to the temp directory
		tempJarPath := filepath.Join(tempDir, "openapi-generator-cli.jar")
		if err := os.WriteFile(tempJarPath, openapiGeneratorCLIJar, 0755); err != nil {
			generationError = fmt.Errorf("failed to write embedded OpenAPI Generator CLI JAR to temp file '%s': %w", tempJarPath, err)
			return "", generationError
		}

		// Construct and execute the command
		args := []string{
			"-jar", tempJarPath,
			"generate",
			"-i", openAPISpecPath,
			"-g", generatorName,
			"-o", finalOutputDir,
		}
		if packageNameProperty != "" {
			args = append(args, "--additional-properties", packageNameProperty)
		}
		for _, prop := range additionalProperties {
			args = append(args, "--additional-properties", prop)
		}

		cmd := exec.CommandContext(ctx, "java", args...)
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			generationError = fmt.Errorf("%s SDK generation using OpenAPI Generator CLI failed: %w. Output:\n%s", language, err, string(output))
			return "", generationError
		}

		s.logger.Info("SDK generation complete", zap.String("language", language), zap.String("output", string(output)))
		return finalSDKOutputPath, nil
	}

	generationError = fmt.Errorf("language '%s' is not supported for SDK generation. Supported: 'go', 'typescript', 'python', 'php', 'java', 'csharp', 'rust', 'ruby'", language)
	return "", generationError
}

// ConvertPostmanToOpenAPI converts a Postman collection string to an OpenAPI v3 string using a JavaScript library.
func (s *SDKService) ConvertPostmanToOpenAPI(ctx context.Context, postmanCollectionJSON string, optionsJSON string) (string, error) {
	s.logger.Info("Attempting to convert Postman collection to OpenAPI")
	startTime := time.Now()

	vm := goja.New()

	// Execute the bundled JavaScript code
	_, err := vm.RunString(jsBundle)
	if err != nil {
		s.logger.Error("failed to execute JavaScript bundle in goja VM", zap.Error(err))
		return "", fmt.Errorf("failed to execute JavaScript bundle: %w", err)
	}

	// Get the conversion function from the VM
	convertFunc, ok := goja.AssertFunction(vm.Get("convertPostmanToOpenAPI"))
	if !ok {
		s.logger.Error("'convertPostmanToOpenAPI' function not found in JavaScript bundle")
		return "", fmt.Errorf("'convertPostmanToOpenAPI' function not found in JavaScript bundle")
	}

	// Prepare arguments for the JavaScript function
	postmanVal := vm.ToValue(postmanCollectionJSON)
	optionsVal := vm.ToValue(optionsJSON) // Pass options as a JSON string

	s.logger.Info("Calling 'convertPostmanToOpenAPI' JavaScript function")

	// Call the JavaScript function
	// Since the JS function is async, it returns a Promise.
	// We need to handle the promise to get the result.
	promise, err := convertFunc(goja.Undefined(), postmanVal, optionsVal)
	if err != nil {
		s.logger.Error("Error calling 'convertPostmanToOpenAPI' JavaScript function", zap.Error(err))
		return "", fmt.Errorf("error calling 'convertPostmanToOpenAPI': %w", err)
	}

	// Handle the promise
	// This requires a bit more setup to properly await a promise in goja,
	// typically by passing a callback or using a more complex promise handling mechanism.
	// For simplicity, and assuming the JS function `convertPostmanToOpenAPI`
	// might be adapted or the environment supports a simpler promise resolution,
	// let's try to get the result.
	// A more robust way would be to have the JS function call a Go callback
	// or use goja_nodejs/eventloop for full promise support.

	// Simplified promise handling:
	// If the promise resolves quickly or the JS environment handles it internally for goja.
	// This part might need adjustment based on how `postman-to-openapi` and your JS wrapper handle async operations.

	var openAPISpec string
	var convertErr error

	// Create channels to receive the result or error from the promise
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Access the 'then' and 'catch' methods of the promise
	thenFunc, okThen := goja.AssertFunction(promise.ToObject(vm).Get("then"))
	if !okThen {
		return "", fmt.Errorf("promise does not have a 'then' method")
	}

	// We will use the two-argument version of 'then' for fulfillment and rejection,
	// so a separate 'catch' is not strictly necessary here if the promise chain is simple.
	// _, okCatch := goja.AssertFunction(promise.ToObject(vm).Get("catch"))
	// if !okCatch {
	// 	return "", fmt.Errorf("promise does not have a 'catch' method")
	// }

	// Define Go functions to be called by 'then' and 'catch'
	onFulfilled := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		result := call.Argument(0).String()
		resultChan <- result
		return goja.Undefined()
	})

	onRejected := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		errStr := "unknown error"
		if jsErr, ok := call.Argument(0).(*goja.Object); ok {
			if msg := jsErr.Get("message"); msg != nil {
				errStr = msg.String()
			} else {
				errStr = jsErr.String()
			}
		} else {
			errStr = call.Argument(0).String()
		}
		errorChan <- fmt.Errorf("JavaScript conversion error: %s", errStr)
		return goja.Undefined()
	})

	// Call promise.then(onFulfilled, onRejected)
	_, err = thenFunc(promise, onFulfilled, onRejected)
	if err != nil {
		return "", fmt.Errorf("failed to call promise.then: %w", err)
	}
	// Call promise.catch(onRejected) - typically .then(null, onRejected) is used, or just .catch(onRejected)
	// For goja, ensure the promise chain is correctly established.
	// The above .then(onFulfilled, onRejected) should cover both cases.

	// Wait for the result or an error
	select {
	case openAPISpec = <-resultChan:
		s.logger.Info("Postman to OpenAPI conversion successful", zap.Duration("duration", time.Since(startTime)))
		return openAPISpec, nil
	case convertErr = <-errorChan:
		s.logger.Error("Postman to OpenAPI conversion failed", zap.Error(convertErr), zap.Duration("duration", time.Since(startTime)))
		return "", convertErr
	case <-ctx.Done():
		s.logger.Error("Context cancelled during Postman to OpenAPI conversion", zap.Error(ctx.Err()))
		return "", ctx.Err()
	case <-time.After(30 * time.Second): // Timeout for the conversion
		s.logger.Error("Timeout during Postman to OpenAPI conversion")
		return "", fmt.Errorf("timeout during Postman to OpenAPI conversion")
	}
}

// GetSDKHistory retrieves a paginated list of SDK records for a specific user.
func (s *SDKService) GetSDKHistory(ctx context.Context, userID string, page, limit int) ([]*models.SDK, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	sdks, totalCount, err := s.sdkRepo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get SDK history", zap.Error(err), zap.String("userID", userID))
		return nil, 0, fmt.Errorf("failed to retrieve SDK history: %w", err)
	}

	s.logger.Info("Retrieved SDK history", zap.String("userID", userID), zap.Int("count", len(sdks)), zap.Int64("total", totalCount))
	return sdks, totalCount, nil
}

// DeleteSDK marks an SDK record as deleted (soft delete).
func (s *SDKService) DeleteSDK(ctx context.Context, sdkID string, userID string) error {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(sdkID)
	if err != nil {
		s.logger.Error("Invalid SDK ID format", zap.String("sdkID", sdkID), zap.Error(err))
		return fmt.Errorf("invalid SDK ID format: %w", err)
	}

	// Perform soft delete - this ensures the SDK belongs to the user
	err = s.sdkRepo.SoftDelete(ctx, objectID, userID)
	if err != nil {
		s.logger.Error("Failed to delete SDK", zap.String("sdkID", sdkID), zap.String("userID", userID), zap.Error(err))
		return fmt.Errorf("failed to delete SDK: %w", err)
	}

	// TODO: Consider deleting actual SDK files from filesystem
	// This would require getting the SDK record first to get the FilePath,
	// then removing the file/directory. For now, we only soft-delete the DB record.

	s.logger.Info("SDK soft-deleted successfully", zap.String("sdkID", sdkID), zap.String("userID", userID))
	return nil
}

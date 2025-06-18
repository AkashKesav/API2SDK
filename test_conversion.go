package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/AkashKesav/API2SDK/internal/services"
	"go.uber.org/zap" // Added for logger
	"gopkg.in/yaml.v2"
)

// convertKeysToStrings recursively converts map keys from interface{} to string.
// This is necessary because json.Marshal cannot handle map[interface{}]interface{}.
func convertKeysToStrings(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[fmt.Sprintf("%v", k)] = convertKeysToStrings(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertKeysToStrings(v)
		}
	}
	return i
}

func main() {
	// Initialize a basic logger for the test
	logger, err := zap.NewDevelopment() // Or zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	// Read the test Postman collection (Click API)
	inputFilePath := "internal/services/testdata/click_api.postman_collection.json"
	log.Printf("Using input Postman collection: %s", inputFilePath)
	collectionJSON, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		log.Fatalf("Failed to read test collection '%s': %v", inputFilePath, err)
	}

	// Validate and clean the input JSON
	var tempInterface interface{}
	if err := json.Unmarshal(collectionJSON, &tempInterface); err != nil {
		log.Fatalf("Input Postman collection '%s' is not valid JSON: %v", inputFilePath, err)
	}
	cleanCollectionBytes, err := json.Marshal(tempInterface)
	if err != nil {
		log.Fatalf("Failed to re-marshal clean Postman collection: %v", err)
	}
	// End of validation and cleaning

	// Create SDKService (minimal setup for testing)
	sdkService, err := services.NewSDKService(
		nil,                        // sdkRepo - not needed for conversion test
		nil,                        // mongoClient - not needed for conversion test
		"",                         // dbName - not needed for conversion test
		nil,                        // postmanClient - not needed for conversion test
		logger,                     // logger - pass the initialized logger
		"",                         // openAPIGenPath - not needed for conversion test
		services.GetPyGenScript(),  // pyFS - use getter
		services.GetPhpGenScript(), // phpFS - use getter
		services.GetPhpVendorZip(), // phpVendorFS - use getter
	)
	if err != nil {
		log.Fatalf("Failed to create SDK service: %v", err)
	}

	// Test conversion
	// options := make(map[string]interface{}) // Removed unused variable
	// optionsJSON, _ := json.Marshal(options) // Removed unused variable

	fmt.Println("Testing Postman to OpenAPI conversion...")
	fmt.Printf("Collection size: %d bytes\n", len(collectionJSON))

	openAPISpec, err := sdkService.ConvertPostmanToOpenAPI(
		context.Background(),
		string(cleanCollectionBytes), // Use the cleaned JSON string
	// string(optionsJSON), // Removed optionsJSON as it's not an expected argument
	)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	fmt.Printf("Conversion successful!\n")
	fmt.Printf("OpenAPI spec size: %d bytes\n", len(openAPISpec))

	// The output from p2o is YAML. We need to parse it as YAML first.
	var yamlSpec interface{}
	if err := yaml.Unmarshal([]byte(openAPISpec), &yamlSpec); err != nil {
		log.Fatalf("Generated OpenAPI spec is not valid YAML: %v\nYAML Content:\n%s", err, openAPISpec)
	}
	log.Println("Generated OpenAPI spec successfully parsed as YAML.")

	// Convert map[interface{}]interface{} to map[string]interface{} for JSON marshalling
	convertedSpec := convertKeysToStrings(yamlSpec)

	// For testing purposes, convert YAML to JSON to ensure it's structured correctly
	// and can be handled as JSON if needed later.
	jsonBytes, err := json.Marshal(convertedSpec)
	if err != nil {
		log.Fatalf("Failed to convert YAML to JSON: %v", err)
	}
	openAPISpec = string(jsonBytes) // Replace openAPISpec with the JSON version for further steps
	log.Println("Successfully converted YAML to JSON for output and validation.")

	// Validate it's valid JSON (after conversion from YAML)
	var spec interface{}
	if err := json.Unmarshal([]byte(openAPISpec), &spec); err != nil {
		log.Fatalf("Converted OpenAPI spec is not valid JSON: %v", err)
	}

	// Write the result to a file (now it will be JSON)
	err = ioutil.WriteFile("test_openapi_output.json", []byte(openAPISpec), 0644)
	if err != nil {
		log.Printf("Warning: Failed to write output file: %v", err)
	} else {
		fmt.Println("OpenAPI spec written to test_openapi_output.json")
	}

	fmt.Println("✅ Postman to OpenAPI conversion test passed!")
}

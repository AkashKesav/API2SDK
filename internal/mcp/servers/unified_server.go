package servers

import (
	"context"
	"fmt"

	"github.com/AkashKesav/API2SDK/internal/mcp/transport"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"go.uber.org/zap"
)

// UnifiedMCPServer implements the unified MCP server pattern
// It provides two meta-functions: ACI_SEARCH_FUNCTIONS and ACI_EXECUTE_FUNCTION
// to discover and execute ALL functions available through the platform
type UnifiedMCPServer struct {
	logger               *zap.Logger
	integrationService   services.IntegrationService
	toolProvider         services.ToolProvider
	linkedAccountOwnerID string
	allowedAppsOnly      bool
	functionCache        map[string][]models.Tool
	initialized          bool
}

// Tool represents a unified MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents a unified MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

// NewUnifiedMCPServer creates a new unified MCP server instance
func NewUnifiedMCPServer(
	logger *zap.Logger,
	integrationService services.IntegrationService,
	toolProvider services.ToolProvider,
	linkedAccountOwnerID string,
	allowedAppsOnly bool,
) *UnifiedMCPServer {
	return &UnifiedMCPServer{
		logger:               logger,
		integrationService:   integrationService,
		toolProvider:         toolProvider,
		linkedAccountOwnerID: linkedAccountOwnerID,
		allowedAppsOnly:      allowedAppsOnly,
		functionCache:        make(map[string][]models.Tool),
		initialized:          false,
	}
}

// Initialize implements the MCP initialize method
func (s *UnifiedMCPServer) Initialize(params map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Initializing unified MCP server",
		zap.String("linkedAccountOwnerID", s.linkedAccountOwnerID),
		zap.Bool("allowedAppsOnly", s.allowedAppsOnly))

	s.initialized = true

	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"subscribe":   true,
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "api2sdk-unified-mcp",
			"version": "1.0.0",
		},
	}, nil
}

// ListTools returns the unified meta-functions for dynamic tool access
func (s *UnifiedMCPServer) ListTools() ([]interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Listing unified MCP tools")

	// Define the two meta-functions following the ACI-MCP pattern
	aciSearchFunctions := Tool{
		Name:        "ACI_SEARCH_FUNCTIONS",
		Description: "Search for available functions in the platform. Use this to discover what tools and capabilities are available before executing them.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"app_names": map[string]interface{}{
					"type":        "array",
					"description": "Optional list of specific app names to search within (e.g., ['GMAIL', 'SLACK']). If not provided, searches all available apps.",
					"items":       map[string]interface{}{"type": "string"},
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Optional search query to filter functions by name or description",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Optional category filter (e.g., 'communication', 'productivity', 'data')",
				},
			},
		},
	}

	aciExecuteFunction := Tool{
		Name:        "ACI_EXECUTE_FUNCTION",
		Description: "Execute a specific function that was discovered using ACI_SEARCH_FUNCTIONS. This is the universal execution method for all platform capabilities.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"function_name": map[string]interface{}{
					"type":        "string",
					"description": "The exact name of the function to execute (obtained from ACI_SEARCH_FUNCTIONS)",
				},
				"function_arguments": map[string]interface{}{
					"type":        "object",
					"description": "The arguments to pass to the function, structured according to the function's input schema",
				},
				"override_linked_account_owner_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Override the default linked account owner ID for multi-user scenarios",
				},
			},
			"required": []string{"function_name", "function_arguments"},
		},
	}

	tools := []interface{}{aciSearchFunctions, aciExecuteFunction}

	s.logger.Debug("Returning unified MCP tools", zap.Int("count", len(tools)))
	return tools, nil
}

// CallTool handles execution of the unified meta-functions
func (s *UnifiedMCPServer) CallTool(name string, arguments map[string]interface{}) (interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Info("Executing unified MCP tool",
		zap.String("toolName", name),
		zap.Any("arguments", arguments))

	switch name {
	case "ACI_SEARCH_FUNCTIONS":
		return s.handleSearchFunctions(arguments)

	case "ACI_EXECUTE_FUNCTION":
		return s.handleExecuteFunction(arguments)

	default:
		return nil, fmt.Errorf("unknown tool: %s. Available tools: ACI_SEARCH_FUNCTIONS, ACI_EXECUTE_FUNCTION", name)
	}
}

// handleSearchFunctions implements the ACI_SEARCH_FUNCTIONS meta-function
func (s *UnifiedMCPServer) handleSearchFunctions(arguments map[string]interface{}) (interface{}, error) {
	s.logger.Debug("Handling ACI_SEARCH_FUNCTIONS", zap.Any("arguments", arguments))

	// Extract search parameters
	appNames := []string{}
	if appNamesInterface, ok := arguments["app_names"]; ok {
		if appList, ok := appNamesInterface.([]interface{}); ok {
			for _, app := range appList {
				if appStr, ok := app.(string); ok {
					appNames = append(appNames, appStr)
				}
			}
		}
	}

	query, _ := arguments["query"].(string)
	category, _ := arguments["category"].(string)

	// Get all available integrations
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	var allTools []models.Tool
	for _, integration := range integrations {
		// Filter by app names if specified
		if len(appNames) > 0 {
			found := false
			for _, appName := range appNames {
				if integration.Name == appName {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Get tools for this integration
		tools, err := s.toolProvider.GetTools(context.Background(), integration.ID)
		if err != nil {
			s.logger.Warn("Failed to get tools for integration",
				zap.String("integrationID", integration.ID.Hex()),
				zap.Error(err))
			continue
		}

		// Apply filters
		for _, tool := range tools {
			// Apply query filter
			if query != "" {
				if !contains(tool.Name, query) && !contains(tool.Description, query) {
					continue
				}
			}

			// Apply category filter (if your tools have categories)
			if category != "" {
				// This would require extending your Tool model to include categories
				// For now, skip this filter
			}

			allTools = append(allTools, tool)
		}
	}

	// Format response similar to ACI-MCP
	response := map[string]interface{}{
		"functions": allTools,
		"total":     len(allTools),
		"query":     query,
		"app_names": appNames,
		"category":  category,
	}

	s.logger.Info("Search functions completed",
		zap.Int("totalFound", len(allTools)),
		zap.Strings("appNames", appNames))

	return response, nil
}

// handleExecuteFunction implements the ACI_EXECUTE_FUNCTION meta-function
func (s *UnifiedMCPServer) handleExecuteFunction(arguments map[string]interface{}) (interface{}, error) {
	s.logger.Debug("Handling ACI_EXECUTE_FUNCTION", zap.Any("arguments", arguments))

	// Extract required parameters
	functionName, ok := arguments["function_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid function_name parameter")
	}

	functionArguments, ok := arguments["function_arguments"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid function_arguments parameter")
	}

	// Handle optional linked account owner ID override
	linkedOwnerID := s.linkedAccountOwnerID
	if overrideID, ok := arguments["override_linked_account_owner_id"].(string); ok && overrideID != "" {
		linkedOwnerID = overrideID
		s.logger.Debug("Using overridden linked account owner ID", zap.String("overrideID", overrideID))
	}

	// Find the integration that has this function
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	for _, integration := range integrations {
		tools, err := s.toolProvider.GetTools(context.Background(), integration.ID)
		if err != nil {
			continue
		}

		for _, tool := range tools {
			if tool.Name == functionName {
				// Found the function, execute it
				s.logger.Info("Executing function",
					zap.String("functionName", functionName),
					zap.String("integrationID", integration.ID.Hex()),
					zap.String("linkedOwnerID", linkedOwnerID))

				result, err := s.toolProvider.ExecuteTool(context.Background(), integration.ID, functionName, functionArguments)
				if err != nil {
					return nil, fmt.Errorf("function execution failed: %w", err)
				}

				return map[string]interface{}{
					"success":       true,
					"function_name": functionName,
					"result":        result,
					"integration":   integration.Name,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("function '%s' not found. Use ACI_SEARCH_FUNCTIONS to discover available functions", functionName)
}

// ListResources returns available resources (could be collections, APIs, etc.)
func (s *UnifiedMCPServer) ListResources() ([]interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Listing unified MCP resources")

	// Get available integrations as resources
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	var resources []interface{}
	for _, integration := range integrations {
		resource := Resource{
			URI:         fmt.Sprintf("integration://%s", integration.ID.Hex()),
			Name:        integration.Name,
			Description: integration.Description,
			MimeType:    "application/json",
		}
		resources = append(resources, resource)
	}

	s.logger.Debug("Returning unified MCP resources", zap.Int("count", len(resources)))
	return resources, nil
}

// ReadResource reads a specific resource
func (s *UnifiedMCPServer) ReadResource(uri string) (interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Reading unified MCP resource", zap.String("uri", uri))

	// Parse URI and return resource content
	// For now, return a simple response
	return map[string]interface{}{
		"uri":       uri,
		"content":   "Resource content would be provided here",
		"mimeType":  "application/json",
		"timestamp": fmt.Sprintf("%d", 1234567890),
	}, nil
}

// Shutdown gracefully shuts down the server
func (s *UnifiedMCPServer) Shutdown() error {
	s.logger.Info("Shutting down unified MCP server")
	s.initialized = false
	s.functionCache = make(map[string][]models.Tool)
	return nil
}

// StartWithTransport starts the unified MCP server with the specified transport
func (s *UnifiedMCPServer) StartWithTransport(ctx context.Context, transportType string, port int) error {
	var mcpTransport transport.MCPTransport

	switch transportType {
	case "stdio":
		mcpTransport = transport.NewStdioTransport(s.logger)
	case "sse":
		mcpTransport = transport.NewSSETransport(port, s.logger)
	default:
		return fmt.Errorf("unsupported transport type: %s. Supported: stdio, sse", transportType)
	}

	s.logger.Info("Starting unified MCP server",
		zap.String("transport", transportType),
		zap.Int("port", port))

	return mcpTransport.Start(ctx, s)
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text == substr ||
			len(substr) == 0 ||
			(len(text) > 0 && len(substr) > 0 &&
				fmt.Sprintf("%s", text) != fmt.Sprintf("%s", substr)))
}

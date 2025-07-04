package servers

import (
	"context"
	"fmt"
	"strings"

	"github.com/AkashKesav/API2SDK/internal/mcp/transport"
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"go.uber.org/zap"
)

// AppsMCPServer implements the apps-specific MCP server pattern
// It provides direct access to functions from specified apps
type AppsMCPServer struct {
	logger               *zap.Logger
	integrationService   services.IntegrationService
	toolProvider         services.ToolProvider
	linkedAccountOwnerID string
	allowedApps          []string
	toolsCache           []models.Tool
	initialized          bool
}

// NewAppsMCPServer creates a new apps-specific MCP server instance
func NewAppsMCPServer(
	logger *zap.Logger,
	integrationService services.IntegrationService,
	toolProvider services.ToolProvider,
	linkedAccountOwnerID string,
	allowedApps []string,
) *AppsMCPServer {
	return &AppsMCPServer{
		logger:               logger,
		integrationService:   integrationService,
		toolProvider:         toolProvider,
		linkedAccountOwnerID: linkedAccountOwnerID,
		allowedApps:          allowedApps,
		toolsCache:           []models.Tool{},
		initialized:          false,
	}
}

// Initialize implements the MCP initialize method
func (s *AppsMCPServer) Initialize(params map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Initializing apps MCP server",
		zap.String("linkedAccountOwnerID", s.linkedAccountOwnerID),
		zap.Strings("allowedApps", s.allowedApps))

	// Pre-load tools from allowed apps
	if err := s.loadToolsFromApps(); err != nil {
		return nil, fmt.Errorf("failed to load tools from apps: %w", err)
	}

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
			"name":        "api2sdk-apps-mcp",
			"version":     "1.0.0",
			"allowedApps": s.allowedApps,
		},
	}, nil
}

// loadToolsFromApps loads all tools from the specified allowed apps
func (s *AppsMCPServer) loadToolsFromApps() error {
	s.logger.Debug("Loading tools from allowed apps", zap.Strings("apps", s.allowedApps))

	// Get all available integrations
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list integrations: %w", err)
	}

	var allTools []models.Tool
	appFound := make(map[string]bool)

	for _, integration := range integrations {
		// Check if this integration is in our allowed apps list
		appAllowed := false
		for _, allowedApp := range s.allowedApps {
			if strings.EqualFold(integration.Name, allowedApp) {
				appAllowed = true
				appFound[allowedApp] = true
				break
			}
		}

		if !appAllowed {
			continue
		}

		// Get tools for this integration
		tools, err := s.toolProvider.GetTools(context.Background(), integration.ID)
		if err != nil {
			s.logger.Warn("Failed to get tools for integration",
				zap.String("integrationID", integration.ID.Hex()),
				zap.String("integrationName", integration.Name),
				zap.Error(err))
			continue
		}

		s.logger.Debug("Loaded tools from app",
			zap.String("app", integration.Name),
			zap.Int("toolCount", len(tools)))

		allTools = append(allTools, tools...)
	}

	// Check if all requested apps were found
	for _, allowedApp := range s.allowedApps {
		if !appFound[allowedApp] {
			s.logger.Warn("Requested app not found", zap.String("app", allowedApp))
		}
	}

	s.toolsCache = allTools
	s.logger.Info("Loaded tools from apps",
		zap.Int("totalTools", len(allTools)),
		zap.Strings("requestedApps", s.allowedApps))

	return nil
}

// ListTools returns all tools from the specified apps
func (s *AppsMCPServer) ListTools() ([]interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Listing apps MCP tools", zap.Int("cachedTools", len(s.toolsCache)))

	// Convert models.Tool to interface{} for MCP compatibility
	var tools []interface{}
	for _, tool := range s.toolsCache {
		mcpTool := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
		tools = append(tools, mcpTool)
	}

	s.logger.Debug("Returning apps MCP tools", zap.Int("count", len(tools)))
	return tools, nil
}

// CallTool handles direct execution of app-specific functions
func (s *AppsMCPServer) CallTool(name string, arguments map[string]interface{}) (interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Info("Executing apps MCP tool",
		zap.String("toolName", name),
		zap.Any("arguments", arguments))

	// Handle optional linked account owner ID override
	linkedOwnerID := s.linkedAccountOwnerID
	if overrideID, ok := arguments["override_linked_account_owner_id"].(string); ok && overrideID != "" {
		linkedOwnerID = overrideID
		s.logger.Debug("Using overridden linked account owner ID", zap.String("overrideID", overrideID))
		// Remove the override parameter from arguments before passing to the tool
		delete(arguments, "override_linked_account_owner_id")
	}

	// Find the tool in our cache
	var targetTool *models.Tool
	for _, tool := range s.toolsCache {
		if tool.Name == name {
			targetTool = &tool
			break
		}
	}

	if targetTool == nil {
		availableTools := make([]string, len(s.toolsCache))
		for i, tool := range s.toolsCache {
			availableTools[i] = tool.Name
		}
		return nil, fmt.Errorf("tool '%s' not found. Available tools: %v", name, availableTools)
	}

	// Find the integration that owns this tool
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	for _, integration := range integrations {
		// Check if this integration is in our allowed apps
		appAllowed := false
		for _, allowedApp := range s.allowedApps {
			if strings.EqualFold(integration.Name, allowedApp) {
				appAllowed = true
				break
			}
		}

		if !appAllowed {
			continue
		}

		// Check if this integration has the requested tool
		tools, err := s.toolProvider.GetTools(context.Background(), integration.ID)
		if err != nil {
			continue
		}

		for _, tool := range tools {
			if tool.Name == name {
				// Found the tool, execute it
				s.logger.Info("Executing tool",
					zap.String("toolName", name),
					zap.String("integrationID", integration.ID.Hex()),
					zap.String("integrationName", integration.Name),
					zap.String("linkedOwnerID", linkedOwnerID))

				result, err := s.toolProvider.ExecuteTool(context.Background(), integration.ID, name, arguments)
				if err != nil {
					return nil, fmt.Errorf("tool execution failed: %w", err)
				}

				return map[string]interface{}{
					"success":     true,
					"tool_name":   name,
					"result":      result,
					"integration": integration.Name,
					"app":         integration.Name,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("tool '%s' execution failed - integration not found", name)
}

// ListResources returns available resources from the specified apps
func (s *AppsMCPServer) ListResources() ([]interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Listing apps MCP resources")

	// Get integrations for allowed apps only
	integrations, err := s.integrationService.ListIntegrations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}

	var resources []interface{}
	for _, integration := range integrations {
		// Check if this integration is in our allowed apps
		appAllowed := false
		for _, allowedApp := range s.allowedApps {
			if strings.EqualFold(integration.Name, allowedApp) {
				appAllowed = true
				break
			}
		}

		if !appAllowed {
			continue
		}

		resource := map[string]interface{}{
			"uri":         fmt.Sprintf("app://%s", integration.Name),
			"name":        integration.Name,
			"description": fmt.Sprintf("%s integration resources", integration.Name),
			"mimeType":    "application/json",
		}
		resources = append(resources, resource)
	}

	s.logger.Debug("Returning apps MCP resources", zap.Int("count", len(resources)))
	return resources, nil
}

// ReadResource reads a specific resource from the allowed apps
func (s *AppsMCPServer) ReadResource(uri string) (interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	s.logger.Debug("Reading apps MCP resource", zap.String("uri", uri))

	// Parse URI to extract app name
	if !strings.HasPrefix(uri, "app://") {
		return nil, fmt.Errorf("invalid URI format. Expected app://appname")
	}

	appName := strings.TrimPrefix(uri, "app://")

	// Check if the app is in our allowed list
	appAllowed := false
	for _, allowedApp := range s.allowedApps {
		if strings.EqualFold(appName, allowedApp) {
			appAllowed = true
			break
		}
	}

	if !appAllowed {
		return nil, fmt.Errorf("app '%s' not allowed. Allowed apps: %v", appName, s.allowedApps)
	}

	// Find tools for this specific app
	var appTools []models.Tool
	for _, tool := range s.toolsCache {
		// This is a simplified approach - in a real implementation,
		// you might want to track which tools belong to which apps
		appTools = append(appTools, tool)
	}

	return map[string]interface{}{
		"uri":         uri,
		"app":         appName,
		"tools":       appTools,
		"toolCount":   len(appTools),
		"mimeType":    "application/json",
		"allowedApps": s.allowedApps,
	}, nil
}

// Shutdown gracefully shuts down the server
func (s *AppsMCPServer) Shutdown() error {
	s.logger.Info("Shutting down apps MCP server")
	s.initialized = false
	s.toolsCache = []models.Tool{}
	return nil
}

// StartWithTransport starts the apps MCP server with the specified transport
func (s *AppsMCPServer) StartWithTransport(ctx context.Context, transportType string, port int) error {
	var mcpTransport transport.MCPTransport

	switch transportType {
	case "stdio":
		mcpTransport = transport.NewStdioTransport(s.logger)
	case "sse":
		mcpTransport = transport.NewSSETransport(port, s.logger)
	default:
		return fmt.Errorf("unsupported transport type: %s. Supported: stdio, sse", transportType)
	}

	s.logger.Info("Starting apps MCP server",
		zap.String("transport", transportType),
		zap.Int("port", port),
		zap.Strings("allowedApps", s.allowedApps))

	return mcpTransport.Start(ctx, s)
}

// GetAllowedApps returns the list of allowed apps for this server
func (s *AppsMCPServer) GetAllowedApps() []string {
	return s.allowedApps
}

// GetToolsCount returns the number of tools loaded from allowed apps
func (s *AppsMCPServer) GetToolsCount() int {
	return len(s.toolsCache)
}

// RefreshTools reloads tools from the allowed apps
func (s *AppsMCPServer) RefreshTools() error {
	if !s.initialized {
		return fmt.Errorf("server not initialized")
	}

	s.logger.Info("Refreshing tools from allowed apps")
	return s.loadToolsFromApps()
}

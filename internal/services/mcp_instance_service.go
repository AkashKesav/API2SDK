package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MCPInstanceService defines the interface for managing MCP instances.
type MCPInstanceService interface {
	CreateMCPInstance(ctx context.Context, mcpInstance *models.MCPInstance) (*models.MCPInstance, error)
	GetMCPInstance(ctx context.Context, id primitive.ObjectID) (*models.MCPInstance, error)
	ListMCPInstances(ctx context.Context, userID string) ([]*models.MCPInstance, error)
	DeleteMCPInstance(ctx context.Context, id primitive.ObjectID) error
	GetTools(ctx context.Context, instanceID primitive.ObjectID) ([]models.Tool, error)
	ExecuteToolCall(ctx context.Context, instanceID primitive.ObjectID, toolName string, args map[string]interface{}) (interface{}, error)
	GetResources(ctx context.Context, instanceID primitive.ObjectID) ([]models.Resource, error)
}

// mcpInstanceService is the concrete implementation of MCPInstanceService.
type mcpInstanceService struct {
	repo               repositories.MCPInstanceRepository
	integrationService IntegrationService
	toolProvider       ToolProvider
	toolCache          map[primitive.ObjectID][]models.Tool
}

// NewMCPInstanceService creates a new MCPInstanceService.
func NewMCPInstanceService(repo repositories.MCPInstanceRepository, integrationService IntegrationService, toolProvider ToolProvider) MCPInstanceService {
	return &mcpInstanceService{
		repo:               repo,
		integrationService: integrationService,
		toolProvider:       toolProvider,
		toolCache:          make(map[primitive.ObjectID][]models.Tool),
	}
}

// CreateMCPInstance creates a new MCP instance.
func (s *mcpInstanceService) CreateMCPInstance(ctx context.Context, mcpInstance *models.MCPInstance) (*models.MCPInstance, error) {
	return s.repo.Create(ctx, mcpInstance)
}

// GetMCPInstance retrieves an MCP instance by its ID.
func (s *mcpInstanceService) GetMCPInstance(ctx context.Context, id primitive.ObjectID) (*models.MCPInstance, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMCPInstances retrieves all MCP instances for a given user.
func (s *mcpInstanceService) ListMCPInstances(ctx context.Context, userID string) ([]*models.MCPInstance, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// DeleteMCPInstance deletes an MCP instance by its ID.
func (s *mcpInstanceService) DeleteMCPInstance(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

var aciSearchFunctionsTool = models.Tool{
	Name:        "ACI_SEARCH_FUNCTIONS",
	Description: "Search for available functions in ACI.",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app_names": map[string]interface{}{
				"type":        "array",
				"description": "A list of app names to search for functions in.",
				"items":       map[string]interface{}{"type": "string"},
			},
		},
	},
}

var aciExecuteFunctionTool = models.Tool{
	Name:        "ACI_EXECUTE_FUNCTION",
	Description: "Execute a function in ACI.",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"function_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the function to execute.",
			},
			"function_arguments": map[string]interface{}{
				"type":        "object",
				"description": "The arguments to pass to the function.",
			},
		},
		"required": []string{"function_name", "function_arguments"},
	},
}

// GetTools returns a list of all tools provided by the integration.
func (s *mcpInstanceService) GetTools(ctx context.Context, instanceID primitive.ObjectID) ([]models.Tool, error) {
	// Check cache first
	if tools, ok := s.toolCache[instanceID]; ok {
		return tools, nil
	}

	mcpInstance, err := s.GetMCPInstance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP instance: %w", err)
	}

	if mcpInstance.ToolsURL != "" {
		resp, err := http.Get(mcpInstance.ToolsURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tools from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch tools: received non-200 status code")
		}

		var tools []models.Tool
		if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
			return nil, fmt.Errorf("failed to parse tools from response: %w", err)
		}

		s.toolCache[instanceID] = tools
		return tools, nil
	}

	if mcpInstance.ServerType == "unified" {
		tools := []models.Tool{aciSearchFunctionsTool, aciExecuteFunctionTool}
		s.toolCache[instanceID] = tools
		return tools, nil
	}

	integrationId, err := primitive.ObjectIDFromHex(mcpInstance.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid integration ID: %w", err)
	}

	tools, err := s.toolProvider.GetTools(ctx, integrationId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}

	s.toolCache[instanceID] = tools
	return tools, nil
}

// ExecuteToolCall runs a specific tool with the given arguments.
func (s *mcpInstanceService) ExecuteToolCall(ctx context.Context, instanceID primitive.ObjectID, toolName string, args map[string]interface{}) (interface{}, error) {
	mcpInstance, err := s.GetMCPInstance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP instance: %w", err)
	}

	integrationId, err := primitive.ObjectIDFromHex(mcpInstance.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid integration ID: %w", err)
	}

	if mcpInstance.ServerType == "unified" {
		if toolName == aciSearchFunctionsTool.Name {
			// This is a simplified implementation. A real implementation would use the arguments to filter the tools.
			tools, err := s.toolProvider.GetTools(ctx, integrationId)
			if err != nil {
				return nil, fmt.Errorf("failed to search for tools: %w", err)
			}
			return tools, nil
		}

		if toolName == aciExecuteFunctionTool.Name {
			funcName, ok := args["function_name"].(string)
			if !ok {
				return nil, fmt.Errorf("missing or invalid function_name")
			}

			funcArgs, ok := args["function_arguments"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("missing or invalid function_arguments")
			}

			result, err := s.toolProvider.ExecuteTool(ctx, integrationId, funcName, funcArgs)
			if err != nil {
				return nil, fmt.Errorf("failed to execute tool: %w", err)
			}
			return result, nil
		}

		return nil, fmt.Errorf("unknown tool for unified server")
	}

	result, err := s.toolProvider.ExecuteTool(ctx, integrationId, toolName, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool: %w", err)
	}

	return result, nil
}

// GetResources returns a list of all resources provided by the integration.
func (s *mcpInstanceService) GetResources(ctx context.Context, instanceID primitive.ObjectID) ([]models.Resource, error) {
	mcpInstance, err := s.GetMCPInstance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP instance: %w", err)
	}

	return mcpInstance.Resources, nil
}

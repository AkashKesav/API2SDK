package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AkashKesav/API2SDK/internal/crypto"
	"github.com/AkashKesav/API2SDK/internal/mcp"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// MCPController handles the requests for the unified MCP system.
type MCPController struct {
	mcpInstanceService services.MCPInstanceService
	integrationService services.IntegrationService
	mcpManager         *mcp.MCPManager
	logger             *zap.Logger
}

// NewMCPController creates a new MCPController.
func NewMCPController(
	mcpInstanceService services.MCPInstanceService,
	integrationService services.IntegrationService,
	mcpManager *mcp.MCPManager,
	logger *zap.Logger,
) *MCPController {
	return &MCPController{
		mcpInstanceService: mcpInstanceService,
		integrationService: integrationService,
		mcpManager:         mcpManager,
		logger:             logger,
	}
}

// StartServer starts a new MCP server
func (c *MCPController) StartServer(ctx fiber.Ctx) error {
	var config mcp.MCPServerConfig
	if err := ctx.Bind().JSON(&config); err != nil {
		c.logger.Error("Failed to parse MCP server config", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate the configuration
	if config.Type == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "server type is required",
		})
	}

	if config.TransportType == "" {
		config.TransportType = "sse" // Default to SSE
	}

	if config.TransportType == "sse" && config.Port == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "port is required for SSE transport",
		})
	}

	if config.Type == mcp.ServerTypeApps && len(config.AllowedApps) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "allowed_apps is required for apps server type",
		})
	}

	c.logger.Info("Starting new MCP server",
		zap.String("type", string(config.Type)),
		zap.String("transport", config.TransportType),
		zap.Int("port", config.Port))

	server, err := c.mcpManager.StartServer(&config)
	if err != nil {
		c.logger.Error("Failed to start MCP server", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start server: %s", err.Error()),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "MCP server started successfully",
		"server":  server,
	})
}

// StopServer stops a running MCP server
func (c *MCPController) StopServer(ctx fiber.Ctx) error {
	serverID := ctx.Params("serverId")
	if serverID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "server ID is required",
		})
	}

	c.logger.Info("Stopping MCP server", zap.String("serverID", serverID))

	err := c.mcpManager.StopServer(serverID)
	if err != nil {
		c.logger.Error("Failed to stop MCP server", zap.String("serverID", serverID), zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to stop server: %s", err.Error()),
		})
	}

	return ctx.JSON(fiber.Map{
		"message":   "MCP server stopped successfully",
		"server_id": serverID,
	})
}

// GetServer returns information about a specific MCP server
func (c *MCPController) GetServer(ctx fiber.Ctx) error {
	serverID := ctx.Params("serverId")
	if serverID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "server ID is required",
		})
	}

	server, err := c.mcpManager.GetServer(serverID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Server not found: %s", err.Error()),
		})
	}

	return ctx.JSON(fiber.Map{
		"server": server,
	})
}

// ListServers returns all running MCP servers
func (c *MCPController) ListServers(ctx fiber.Ctx) error {
	servers := c.mcpManager.ListServers()
	return ctx.JSON(fiber.Map{
		"servers": servers,
		"count":   len(servers),
	})
}

// GetMCPMetrics returns metrics about the MCP system
func (c *MCPController) GetMCPMetrics(ctx fiber.Ctx) error {
	metrics := c.mcpManager.GetMetrics()
	return ctx.JSON(fiber.Map{
		"metrics": metrics,
	})
}

// StopAllServers stops all running MCP servers
func (c *MCPController) StopAllServers(ctx fiber.Ctx) error {
	c.logger.Info("Stopping all MCP servers")

	err := c.mcpManager.StopAllServers()
	if err != nil {
		c.logger.Error("Failed to stop all servers", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to stop all servers: %s", err.Error()),
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "All MCP servers stopped successfully",
	})
}

// CleanupServers removes stopped servers from the manager
func (c *MCPController) CleanupServers(ctx fiber.Ctx) error {
	c.mcpManager.CleanupStoppedServers()
	return ctx.JSON(fiber.Map{
		"message": "Stopped servers cleaned up successfully",
	})
}

// Legacy methods for backward compatibility

// HandleRequest handles all incoming requests to the MCP (legacy method).
func (c *MCPController) HandleRequest(ctx fiber.Ctx) error {
	path := ctx.Params("*")
	if strings.HasSuffix(path, "/list_tools") {
		return c.ListTools(ctx)
	}
	if strings.HasSuffix(path, "/call_tool") {
		return c.CallTool(ctx)
	}

	instanceId := ctx.Params("instanceId")
	// Convert the instanceId to a primitive.ObjectID
	objId, err := primitive.ObjectIDFromHex(instanceId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid instance ID")
	}

	// Get the MCP instance from the database
	mcpInstance, err := c.mcpInstanceService.GetMCPInstance(ctx.Context(), objId)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to get MCP instance")
	}

	// Get the integration from the database
	integrationId, err := primitive.ObjectIDFromHex(mcpInstance.IntegrationID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("Invalid integration ID")
	}
	integration, err := c.integrationService.GetIntegration(ctx.Context(), integrationId)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to get integration")
	}

	// Decrypt the API key
	decryptedAPIKey, err := crypto.Decrypt(string(integration.APIKey))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to decrypt API key")
	}

	// Forward the request to the integration's API
	return forwardRequest(ctx, integration.BaseURL, decryptedAPIKey)
}

// ListTools returns a list of all tools provided by the integration (legacy method).
func (c *MCPController) ListTools(ctx fiber.Ctx) error {
	instanceId := ctx.Params("instanceId")
	objId, err := primitive.ObjectIDFromHex(instanceId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid instance ID")
	}

	tools, err := c.mcpInstanceService.GetTools(ctx.Context(), objId)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return ctx.JSON(tools)
}

// CallTool runs a specific tool with the given arguments (legacy method).
func (c *MCPController) CallTool(ctx fiber.Ctx) error {
	instanceId := ctx.Params("instanceId")
	objId, err := primitive.ObjectIDFromHex(instanceId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid instance ID")
	}

	var body struct {
		ToolName  string                 `json:"tool_name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := ctx.Bind().JSON(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	result, err := c.mcpInstanceService.ExecuteToolCall(ctx.Context(), objId, body.ToolName, body.Arguments)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return ctx.JSON(result)
}

// forwardRequest handles the logic of forwarding the request to the target API.
func forwardRequest(c fiber.Ctx, baseURL string, apiKey string) error {
	// Construct the target URL
	path := c.Params("*")
	targetURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), path)

	// Create a new request
	req, err := http.NewRequestWithContext(c.Context(), string(c.Method()), targetURL, strings.NewReader(string(c.Body())))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create request")
	}

	// Whitelist of headers to forward
	headerWhitelist := map[string]bool{
		"Content-Type": true,
		"Accept":       true,
		// Add other safe headers here
	}

	// Copy whitelisted headers from the original request
	c.Request().Header.VisitAll(func(key, value []byte) {
		if headerWhitelist[string(key)] {
			req.Header.Set(string(key), string(value))
		}
	})

	// Add the API key to the request if it exists
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("Failed to forward request")
	}
	defer resp.Body.Close()

	// Copy headers from the response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Copy the status code
	c.Status(resp.StatusCode)

	// Copy the body
	if _, err := io.Copy(c.Response().BodyWriter(), resp.Body); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to copy response body")
	}

	return nil
}

// StreamTool handles SSE connections for streaming tool results (legacy method).
func (c *MCPController) StreamTool(ctx fiber.Ctx) error {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")

	ctx.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer w.Flush()

		instanceId := ctx.Params("instanceId")
		objId, err := primitive.ObjectIDFromHex(instanceId)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Invalid instance ID\"}\n\n")
			return
		}

		mcpInstance, err := c.mcpInstanceService.GetMCPInstance(ctx.Context(), objId)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Failed to get MCP instance\"}\n\n")
			return
		}

		if _, err := primitive.ObjectIDFromHex(mcpInstance.IntegrationID); err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Invalid integration ID\"}\n\n")
			return
		}

		toolName := ctx.Query("tool_name")
		if toolName == "" {
			fmt.Fprintf(w, "data: {\"error\": \"Missing tool_name query parameter\"}\n\n")
			return
		}

		var arguments map[string]interface{}
		if err := json.Unmarshal([]byte(ctx.Query("arguments", "{}")), &arguments); err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Invalid arguments query parameter\"}\n\n")
			return
		}

		result, err := c.mcpInstanceService.ExecuteToolCall(ctx.Context(), objId, toolName, arguments)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Failed to execute tool: %s\"}\n\n", err.Error())
			return
		}

		resultBytes, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Failed to marshal result: %s\"}\n\n", err.Error())
			return
		}

		fmt.Fprintf(w, "data: %s\n\n", string(resultBytes))
		w.Flush()
	})

	return nil
}

package routes

import (
	"github.com/AkashKesav/API2SDK/internal/controllers"
	"github.com/gofiber/fiber/v3"
)

// MCPRouter is the router for the unified MCP system.
type MCPRouter struct {
	app           fiber.Router
	mcpController *controllers.MCPController
}

// NewMCPRouter creates a new MCPRouter.
func NewMCPRouter(app fiber.Router, mcpController *controllers.MCPController) *MCPRouter {
	return &MCPRouter{
		app:           app,
		mcpController: mcpController,
	}
}

// SetupRoutes sets up the routes for the MCP system.
func (r *MCPRouter) SetupRoutes() {
	// New unified MCP server management endpoints
	unified := r.app.Group("/unified")
	{
		// Server management
		unified.Post("/servers", r.mcpController.StartServer)            // Start a new MCP server
		unified.Get("/servers", r.mcpController.ListServers)             // List all running servers
		unified.Get("/servers/:serverId", r.mcpController.GetServer)     // Get server info
		unified.Delete("/servers/:serverId", r.mcpController.StopServer) // Stop a specific server
		unified.Delete("/servers", r.mcpController.StopAllServers)       // Stop all servers

		// Metrics and management
		unified.Get("/metrics", r.mcpController.GetMCPMetrics)   // Get MCP system metrics
		unified.Post("/cleanup", r.mcpController.CleanupServers) // Cleanup stopped servers
	}

	// Legacy MCP instance endpoints for backward compatibility
	legacy := r.app.Group("/instances")
	{
		legacy.Get("/:instanceId/sse", r.mcpController.StreamTool)
		legacy.All("/:instanceId/*", r.mcpController.HandleRequest)
	}

	// For backward compatibility, also support the old format
	r.app.Get("/:instanceId/sse", r.mcpController.StreamTool)
	r.app.All("/:instanceId/*", r.mcpController.HandleRequest)
}

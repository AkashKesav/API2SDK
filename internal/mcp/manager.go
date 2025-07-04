package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/AkashKesav/API2SDK/internal/mcp/servers"
	"github.com/AkashKesav/API2SDK/internal/services"
	"go.uber.org/zap"
)

// MCPServerType represents the type of MCP server
type MCPServerType string

const (
	ServerTypeUnified MCPServerType = "unified"
	ServerTypeApps    MCPServerType = "apps"
)

// MCPServerConfig holds configuration for starting an MCP server
type MCPServerConfig struct {
	Type                 MCPServerType `json:"type"`
	TransportType        string        `json:"transport_type"` // "stdio" or "sse"
	Port                 int           `json:"port,omitempty"`
	LinkedAccountOwnerID string        `json:"linked_account_owner_id"`
	AllowedApps          []string      `json:"allowed_apps,omitempty"` // For apps server
	AllowedAppsOnly      bool          `json:"allowed_apps_only"`      // For unified server
}

// RunningServer represents a running MCP server instance
type RunningServer struct {
	ID          string             `json:"id"`
	Type        MCPServerType      `json:"type"`
	Transport   string             `json:"transport"`
	Port        int                `json:"port,omitempty"`
	Status      string             `json:"status"`
	AllowedApps []string           `json:"allowed_apps,omitempty"`
	Cancel      context.CancelFunc `json:"-"`
	StartedAt   int64              `json:"started_at"`
}

// MCPManager manages multiple MCP server instances
type MCPManager struct {
	logger             *zap.Logger
	integrationService services.IntegrationService
	toolProvider       services.ToolProvider
	runningServers     map[string]*RunningServer
	mu                 sync.RWMutex
}

// NewMCPManager creates a new MCP manager
func NewMCPManager(
	logger *zap.Logger,
	integrationService services.IntegrationService,
	toolProvider services.ToolProvider,
) *MCPManager {
	return &MCPManager{
		logger:             logger,
		integrationService: integrationService,
		toolProvider:       toolProvider,
		runningServers:     make(map[string]*RunningServer),
	}
}

// StartServer starts a new MCP server with the given configuration
func (m *MCPManager) StartServer(config *MCPServerConfig) (*RunningServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique server ID
	serverID := fmt.Sprintf("%s_%s_%d", config.Type, config.TransportType, len(m.runningServers)+1)

	// Check if port is already in use (for SSE servers)
	if config.TransportType == "sse" {
		for _, server := range m.runningServers {
			if server.Port == config.Port && server.Status == "running" {
				return nil, fmt.Errorf("port %d already in use by server %s", config.Port, server.ID)
			}
		}
	}

	m.logger.Info("Starting MCP server",
		zap.String("serverID", serverID),
		zap.String("type", string(config.Type)),
		zap.String("transport", config.TransportType),
		zap.Int("port", config.Port))

	// Create context for this server
	ctx, cancel := context.WithCancel(context.Background())

	// Create running server entry
	runningServer := &RunningServer{
		ID:          serverID,
		Type:        config.Type,
		Transport:   config.TransportType,
		Port:        config.Port,
		Status:      "starting",
		AllowedApps: config.AllowedApps,
		Cancel:      cancel,
		StartedAt:   int64(1234567890), // This should be time.Now().Unix() in production
	}

	m.runningServers[serverID] = runningServer

	// Start the appropriate server type
	go func() {
		defer func() {
			m.mu.Lock()
			if server, exists := m.runningServers[serverID]; exists {
				server.Status = "stopped"
			}
			m.mu.Unlock()
		}()

		var err error
		switch config.Type {
		case ServerTypeUnified:
			err = m.startUnifiedServer(ctx, config)
		case ServerTypeApps:
			err = m.startAppsServer(ctx, config)
		default:
			err = fmt.Errorf("unsupported server type: %s", config.Type)
		}

		if err != nil {
			m.logger.Error("Failed to start MCP server",
				zap.String("serverID", serverID),
				zap.Error(err))
			m.mu.Lock()
			if server, exists := m.runningServers[serverID]; exists {
				server.Status = "failed"
			}
			m.mu.Unlock()
			return
		}

		// Mark as running
		m.mu.Lock()
		if server, exists := m.runningServers[serverID]; exists {
			server.Status = "running"
		}
		m.mu.Unlock()

		m.logger.Info("MCP server started successfully", zap.String("serverID", serverID))

		// Wait for context cancellation
		<-ctx.Done()
		m.logger.Info("MCP server stopped", zap.String("serverID", serverID))
	}()

	return runningServer, nil
}

// startUnifiedServer starts a unified MCP server
func (m *MCPManager) startUnifiedServer(ctx context.Context, config *MCPServerConfig) error {
	server := servers.NewUnifiedMCPServer(
		m.logger,
		m.integrationService,
		m.toolProvider,
		config.LinkedAccountOwnerID,
		config.AllowedAppsOnly,
	)

	return server.StartWithTransport(ctx, config.TransportType, config.Port)
}

// startAppsServer starts an apps-specific MCP server
func (m *MCPManager) startAppsServer(ctx context.Context, config *MCPServerConfig) error {
	if len(config.AllowedApps) == 0 {
		return fmt.Errorf("allowed_apps must be specified for apps server type")
	}

	server := servers.NewAppsMCPServer(
		m.logger,
		m.integrationService,
		m.toolProvider,
		config.LinkedAccountOwnerID,
		config.AllowedApps,
	)

	return server.StartWithTransport(ctx, config.TransportType, config.Port)
}

// StopServer stops a running MCP server
func (m *MCPManager) StopServer(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.runningServers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if server.Status != "running" {
		return fmt.Errorf("server %s is not running (status: %s)", serverID, server.Status)
	}

	m.logger.Info("Stopping MCP server", zap.String("serverID", serverID))

	// Cancel the server context
	server.Cancel()
	server.Status = "stopping"

	return nil
}

// GetServer returns information about a specific server
func (m *MCPManager) GetServer(serverID string) (*RunningServer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.runningServers[serverID]
	if !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	// Return a copy to avoid external modification
	return &RunningServer{
		ID:          server.ID,
		Type:        server.Type,
		Transport:   server.Transport,
		Port:        server.Port,
		Status:      server.Status,
		AllowedApps: server.AllowedApps,
		StartedAt:   server.StartedAt,
	}, nil
}

// ListServers returns all running servers
func (m *MCPManager) ListServers() map[string]*RunningServer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*RunningServer)
	for id, server := range m.runningServers {
		result[id] = &RunningServer{
			ID:          server.ID,
			Type:        server.Type,
			Transport:   server.Transport,
			Port:        server.Port,
			Status:      server.Status,
			AllowedApps: server.AllowedApps,
			StartedAt:   server.StartedAt,
		}
	}

	return result
}

// GetServerStatus returns the status of all servers
func (m *MCPManager) GetServerStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for id, server := range m.runningServers {
		status[id] = server.Status
	}

	return status
}

// StopAllServers stops all running servers
func (m *MCPManager) StopAllServers() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Stopping all MCP servers")

	for serverID, server := range m.runningServers {
		if server.Status == "running" {
			m.logger.Info("Stopping server", zap.String("serverID", serverID))
			server.Cancel()
			server.Status = "stopping"
		}
	}

	return nil
}

// CleanupStoppedServers removes stopped servers from the manager
func (m *MCPManager) CleanupStoppedServers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for serverID, server := range m.runningServers {
		if server.Status == "stopped" || server.Status == "failed" {
			delete(m.runningServers, serverID)
			m.logger.Debug("Cleaned up stopped server", zap.String("serverID", serverID))
		}
	}
}

// GetMetrics returns metrics about the MCP manager
func (m *MCPManager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	transportCounts := make(map[string]int)

	for _, server := range m.runningServers {
		statusCounts[server.Status]++
		typeCounts[string(server.Type)]++
		transportCounts[server.Transport]++
	}

	return map[string]interface{}{
		"total_servers":    len(m.runningServers),
		"status_counts":    statusCounts,
		"type_counts":      typeCounts,
		"transport_counts": transportCounts,
	}
}

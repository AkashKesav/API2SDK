package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// MCPMessage represents a standard MCP protocol message
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPServer interface that MCP servers must implement
type MCPServer interface {
	Initialize(params map[string]interface{}) (map[string]interface{}, error)
	ListTools() ([]interface{}, error)
	CallTool(name string, arguments map[string]interface{}) (interface{}, error)
	ListResources() ([]interface{}, error)
	ReadResource(uri string) (interface{}, error)
	Shutdown() error
}

// MCPTransport interface for different transport mechanisms
type MCPTransport interface {
	Start(ctx context.Context, server MCPServer) error
	Stop() error
	SendMessage(message MCPMessage) error
}

// StdioTransport implements MCP over stdio
type StdioTransport struct {
	logger *zap.Logger
	reader *bufio.Scanner
	writer io.Writer
	server MCPServer
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(logger *zap.Logger) *StdioTransport {
	return &StdioTransport{
		logger: logger,
		reader: bufio.NewScanner(os.Stdin),
		writer: os.Stdout,
	}
}

// Start begins the stdio transport loop
func (t *StdioTransport) Start(ctx context.Context, server MCPServer) error {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.server = server

	t.logger.Info("Starting MCP stdio transport")

	// Send initialization
	initMessage := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
		},
	}

	if err := t.SendMessage(initMessage); err != nil {
		return fmt.Errorf("failed to send initialization: %w", err)
	}

	// Start message processing loop
	go t.messageLoop()

	return nil
}

// messageLoop processes incoming messages
func (t *StdioTransport) messageLoop() {
	for t.reader.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := t.reader.Text()
		if line == "" {
			continue
		}

		var message MCPMessage
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			t.logger.Error("Failed to parse MCP message", zap.Error(err), zap.String("line", line))
			continue
		}

		t.handleMessage(message)
	}

	if err := t.reader.Err(); err != nil {
		t.logger.Error("Stdio scanner error", zap.Error(err))
	}
}

// handleMessage processes individual MCP messages
func (t *StdioTransport) handleMessage(message MCPMessage) {
	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      message.ID,
	}

	switch message.Method {
	case "initialize":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		result, err := t.server.Initialize(params)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Internal error during initialization",
				Data:    err.Error(),
			}
		} else {
			response.Result = result
		}

	case "tools/list":
		tools, err := t.server.ListTools()
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to list tools",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"tools": tools,
			}
		}

	case "tools/call":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		name, _ := params["name"].(string)
		arguments := make(map[string]interface{})
		if args, ok := params["arguments"].(map[string]interface{}); ok {
			arguments = args
		}

		result, err := t.server.CallTool(name, arguments)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Tool execution failed",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("%v", result),
					},
				},
			}
		}

	case "resources/list":
		resources, err := t.server.ListResources()
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to list resources",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"resources": resources,
			}
		}

	case "resources/read":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		uri, _ := params["uri"].(string)
		result, err := t.server.ReadResource(uri)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to read resource",
				Data:    err.Error(),
			}
		} else {
			response.Result = result
		}

	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: "Method not found",
			Data:    message.Method,
		}
	}

	if err := t.SendMessage(response); err != nil {
		t.logger.Error("Failed to send response", zap.Error(err))
	}
}

// SendMessage sends an MCP message via stdio
func (t *StdioTransport) SendMessage(message MCPMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = fmt.Fprintln(t.writer, string(data))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Stop stops the stdio transport
func (t *StdioTransport) Stop() error {
	if t.cancel != nil {
		t.cancel()
	}
	return nil
}

// SSETransport implements MCP over Server-Sent Events using Fiber v3
type SSETransport struct {
	logger *zap.Logger
	port   int
	server MCPServer
	app    *fiber.App
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewSSETransport creates a new SSE transport using Fiber v3
func NewSSETransport(port int, logger *zap.Logger) *SSETransport {
	return &SSETransport{
		logger: logger,
		port:   port,
	}
}

// Start begins the SSE transport server using Fiber v3
func (t *SSETransport) Start(ctx context.Context, server MCPServer) error {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.server = server

	// Create Fiber v3 app
	t.app = fiber.New(fiber.Config{
		ReadTimeout:  0, // Disable timeout for SSE
		WriteTimeout: 0,
	})

	// Setup CORS for SSE
	t.app.Use(func(c fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Headers", "Content-Type")
		c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusOK)
		}

		return c.Next()
	})

	// SSE endpoint for real-time communication
	t.app.Get("/sse", t.handleSSE)

	// HTTP endpoint for message posting
	t.app.Post("/message", t.handleMessage)

	// Health check endpoint
	t.app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(map[string]interface{}{
			"status":      "healthy",
			"transport":   "sse",
			"mcp_version": "2024-11-05",
		})
	})

	t.logger.Info("Starting MCP SSE transport with Fiber v3", zap.Int("port", t.port))

	// Start server in goroutine
	go func() {
		if err := t.app.Listen(fmt.Sprintf(":%d", t.port)); err != nil {
			t.logger.Error("SSE transport server error", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-t.ctx.Done()
	return t.app.Shutdown()
}

// handleSSE handles SSE connections using Fiber v3
func (t *SSETransport) handleSSE(c fiber.Ctx) error {
	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Send initialization message
	initMessage := MCPMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
		},
	}

	data, err := json.Marshal(initMessage)
	if err != nil {
		return err
	}

	// Send SSE formatted message using Fiber's response writer
	c.Response().SetBodyString(fmt.Sprintf("data: %s\n\n", data))

	// Keep connection alive until client disconnects or context is cancelled
	select {
	case <-t.ctx.Done():
		return nil
	case <-c.Context().Done():
		return nil
	}
}

// handleMessage handles HTTP POST messages for SSE using Fiber v3
func (t *SSETransport) handleMessage(c fiber.Ctx) error {
	var message MCPMessage
	if err := c.Bind().JSON(&message); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	// Process the MCP message
	response := t.processMessage(message)

	return c.JSON(response)
}

// processMessage processes MCP messages for SSE transport
func (t *SSETransport) processMessage(message MCPMessage) MCPMessage {
	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      message.ID,
	}

	switch message.Method {
	case "initialize":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		result, err := t.server.Initialize(params)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Internal error during initialization",
				Data:    err.Error(),
			}
		} else {
			response.Result = result
		}

	case "tools/list":
		tools, err := t.server.ListTools()
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to list tools",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"tools": tools,
			}
		}

	case "tools/call":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		name, _ := params["name"].(string)
		arguments := make(map[string]interface{})
		if args, ok := params["arguments"].(map[string]interface{}); ok {
			arguments = args
		}

		result, err := t.server.CallTool(name, arguments)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Tool execution failed",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("%v", result),
					},
				},
			}
		}

	case "resources/list":
		resources, err := t.server.ListResources()
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to list resources",
				Data:    err.Error(),
			}
		} else {
			response.Result = map[string]interface{}{
				"resources": resources,
			}
		}

	case "resources/read":
		params := make(map[string]interface{})
		if message.Params != nil {
			if p, ok := message.Params.(map[string]interface{}); ok {
				params = p
			}
		}

		uri, _ := params["uri"].(string)
		result, err := t.server.ReadResource(uri)
		if err != nil {
			response.Error = &MCPError{
				Code:    -32603,
				Message: "Failed to read resource",
				Data:    err.Error(),
			}
		} else {
			response.Result = result
		}

	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: "Method not found",
			Data:    message.Method,
		}
	}

	return response
}

// SendMessage sends an MCP message via SSE (typically not used in SSE pattern)
func (t *SSETransport) SendMessage(message MCPMessage) error {
	// In SSE, messages are typically sent through the SSE stream
	// This method could be used for server-initiated messages if needed
	return nil
}

// Stop stops the SSE transport
func (t *SSETransport) Stop() error {
	if t.app != nil {
		if err := t.app.Shutdown(); err != nil {
			t.logger.Error("Error shutting down SSE transport", zap.Error(err))
		}
	}

	if t.cancel != nil {
		t.cancel()
	}

	return nil
}

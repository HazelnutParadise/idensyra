package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/HazelnutParadise/idensyra/mcp"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents parameters for initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the result of initialize method
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// Capabilities represents server capabilities
type Capabilities struct {
	Tools map[string]interface{} `json:"tools,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CallToolParams represents parameters for tools/call method
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// MCPHTTPServer wraps the MCP server for HTTP access
type MCPHTTPServer struct {
	server     *mcp.Server
	httpServer *http.Server
	mu         sync.Mutex
	app        *App
}

// NewMCPHTTPServer creates a new HTTP MCP server
func NewMCPHTTPServer(app *App) *MCPHTTPServer {
	return &MCPHTTPServer{
		app: app,
	}
}

// Start initializes and starts the MCP HTTP server
func (m *MCPHTTPServer) Start(port int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get workspace root
	workspaceRoot := "."
	if globalWorkspace != nil {
		globalWorkspace.mu.RLock()
		workspaceRoot = globalWorkspace.rootPath
		globalWorkspace.mu.RUnlock()
	}

	// Make it absolute
	absWorkspace, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace path: %v", err)
	}

	// Create config with default permissions
	config := mcp.DefaultConfig()

	// Confirmation function that shows dialog in UI
	confirmFunc := func(operation, details string) bool {
		// In GUI mode, we can show a dialog - for now auto-approve
		log.Printf("[MCP] Operation: %s - %s (auto-approved)", operation, details)
		return true
	}

	// Execute Go code using the app's method
	executeGoFunc := func(code string, colorBG string) string {
		return m.app.ExecuteCodeWithColorBG(code, colorBG)
	}

	// Execute Python file using the app's method
	executePyFunc := func(filePath string) (string, error) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		// Get relative path for the app method
		relPath, _ := filepath.Rel(absWorkspace, filePath)
		if relPath == "" {
			relPath = filepath.Base(filePath)
		}
		result, err := m.app.ExecutePythonFileContent(relPath, string(content))
		return result, err
	}

	// Execute cell code
	executeCellFunc := func(language, code string) (string, error) {
		switch language {
		case "go":
			return executeGoFunc(code, "dark"), nil
		case "python":
			// For Python, we need to create a temp file
			tmpFile := filepath.Join(absWorkspace, fmt.Sprintf(".tmp_mcp_py_%d.py", os.Getpid()))
			defer os.Remove(tmpFile)
			if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
				return "", fmt.Errorf("failed to create temp file: %v", err)
			}
			return executePyFunc(tmpFile)
		case "markdown":
			return "Markdown cell (no execution)", nil
		default:
			return "", fmt.Errorf("unsupported language: %s", language)
		}
	}

	// Workspace operations using the app's methods
	openWorkspaceFunc := func(path string) error {
		return m.app.OpenWorkspace(path)
	}

	saveWorkspaceFunc := func(path string) error {
		return m.app.CreateWorkspace(path)
	}

	saveChangesFunc := func() error {
		return m.app.SaveAllFiles()
	}

	// Set active file function that uses the app's method
	setActiveFileFunc := func(path string) error {
		return m.app.SetActiveFile(path)
	}

	// Import file function that uses the app's method
	importFileFunc := func(sourcePath, targetDir string) error {
		return m.app.ImportSpecificFileToWorkspace(sourcePath, targetDir)
	}

	// Create MCP server
	m.server = mcp.NewServer(
		config,
		absWorkspace,
		confirmFunc,
		executeGoFunc,
		executePyFunc,
		executeCellFunc,
		openWorkspaceFunc,
		saveWorkspaceFunc,
		saveChangesFunc,
		setActiveFileFunc,
		importFileFunc,
	)

	// Create HTTP handler
	mux := http.NewServeMux()

	// MCP protocol endpoint
	mux.HandleFunc("/", m.handleMCPProtocol)

	// Create HTTP server
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		log.Printf("[MCP] Starting MCP HTTP server on http://localhost:%d", port)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[MCP] HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (m *MCPHTTPServer) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.httpServer != nil {
		log.Println("[MCP] Stopping MCP HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()
		return m.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleMCPProtocol handles MCP JSON-RPC 2.0 protocol requests
func (m *MCPHTTPServer) handleMCPProtocol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &RPCError{
				Code:    -32600,
				Message: "Invalid Request - only POST is allowed",
			},
		})
		return
	}

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &RPCError{
				Code:    -32700,
				Message: "Parse error",
				Data:    err.Error(),
			},
		})
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32600,
				Message: "Invalid Request - jsonrpc must be '2.0'",
			},
		})
		return
	}

	// Handle different MCP methods
	var response JSONRPCResponse
	response.JSONRPC = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "initialize":
		response.Result = m.handleInitialize(req.Params)

	case "tools/list":
		response.Result = m.handleToolsList()

	case "tools/call":
		result, err := m.handleToolsCall(r.Context(), req.Params)
		if err != nil {
			response.Error = &RPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			}
		} else {
			response.Result = result
		}

	default:
		response.Error = &RPCError{
			Code:    -32601,
			Message: "Method not found",
			Data:    fmt.Sprintf("Unknown method: %s", req.Method),
		}
	}

	json.NewEncoder(w).Encode(response)
}

// handleInitialize handles the initialize method
func (m *MCPHTTPServer) handleInitialize(params json.RawMessage) InitializeResult {
	// Parse params if needed (for now, we don't need specific client capabilities)
	var initParams InitializeParams
	if params != nil {
		json.Unmarshal(params, &initParams)
	}

	return InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: Capabilities{
			Tools: map[string]interface{}{},
		},
		ServerInfo: ServerInfo{
			Name:    "idensyra-mcp-server",
			Version: "1.0.0",
		},
	}
}

// handleToolsList handles the tools/list method
func (m *MCPHTTPServer) handleToolsList() map[string]interface{} {
	tools := m.server.ListTools()

	// Convert to MCP tools format
	mcpTools := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		mcpTools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]interface{}{
		"tools": mcpTools,
	}
}

// handleToolsCall handles the tools/call method
func (m *MCPHTTPServer) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var callParams CallToolParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid params: %v", err)
	}

	// Create tool request
	toolReq := &mcp.ToolRequest{
		Name:      callParams.Name,
		Arguments: callParams.Arguments,
	}

	// Execute tool
	resp, err := m.server.HandleRequest(ctx, toolReq)
	if err != nil {
		return nil, err
	}

	// Convert response to MCP format
	return map[string]interface{}{
		"content": resp.Content,
		"isError": resp.IsError,
	}, nil
}

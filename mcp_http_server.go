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
	)

	// Create HTTP handler
	mux := http.NewServeMux()

	// Handle MCP tool calls
	mux.HandleFunc("/mcp/call", m.handleToolCall)

	// Handle MCP tool list
	mux.HandleFunc("/mcp/tools", m.handleListTools)

	// Health check
	mux.HandleFunc("/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

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

// handleToolCall handles MCP tool call requests
func (m *MCPHTTPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req mcp.ToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := m.server.HandleRequest(r.Context(), &req)
	if err != nil {
		log.Printf("[MCP] Error handling request: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleListTools returns the list of available tools
func (m *MCPHTTPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := m.server.ListTools()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
	})
}

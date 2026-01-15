package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MCPServer wraps the MCP SDK server for HTTP access via SSE
type MCPServer struct {
	server     *sdk.Server
	httpServer *http.Server
	app        *App

	pendingResults map[string]chan string
	pendingMu      sync.Mutex
}

// NewMCPServer creates a new MCP server using the official SDK
func NewMCPServer(app *App) *MCPServer {
	return &MCPServer{
		app:            app,
		pendingResults: make(map[string]chan string),
	}
}

// waitForExecutionResult waits for a result published by frontend for given requestId
func (m *MCPServer) waitForExecutionResult(requestId string, timeout time.Duration) (string, error) {
	m.pendingMu.Lock()
	ch := make(chan string, 1)
	m.pendingResults[requestId] = ch
	m.pendingMu.Unlock()

	select {
	case res := <-ch:
		// cleanup
		m.pendingMu.Lock()
		delete(m.pendingResults, requestId)
		m.pendingMu.Unlock()
		return res, nil
	case <-time.After(timeout):
		m.pendingMu.Lock()
		delete(m.pendingResults, requestId)
		m.pendingMu.Unlock()
		return "", fmt.Errorf("timeout waiting for execution result")
	}
}

// dispatchUIAction sends an MCP action to the frontend and waits for a response.
func (m *MCPServer) dispatchUIAction(action string, payload map[string]any, timeout time.Duration) (string, error) {
	requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
	if payload == nil {
		payload = map[string]any{}
	}
	payload["action"] = action
	payload["request_id"] = requestId

	// Emit asynchronously to avoid blocking the MCP handler if the UI event loop stalls.
	go runtime.EventsEmit(m.app.ctx, "mcp:ui_action", payload)
	return m.waitForExecutionResult(requestId, timeout)
}

// deliverExecutionResult is called by App when frontend finishes executing and posts result
func (m *MCPServer) deliverExecutionResult(requestId string, result string) {
	m.pendingMu.Lock()
	ch, ok := m.pendingResults[requestId]
	m.pendingMu.Unlock()
	if ok {
		select {
		case ch <- result:
		default:
		}
	}
}

// Start initializes and starts the MCP server with SSE HTTP transport
func (m *MCPServer) Start(port int) error {
	// Get workspace root
	workspaceRoot := "."
	if globalWorkspace != nil {
		globalWorkspace.mu.RLock()
		workspaceRoot = globalWorkspace.workDir
		globalWorkspace.mu.RUnlock()
	}

	// Make it absolute
	absWorkspace, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace path: %v", err)
	}

	// Create MCP server with SDK
	impl := &sdk.Implementation{
		Name:    "idensyra",
		Version: "1.0.0",
	}

	opts := &sdk.ServerOptions{
		Instructions: "Idensyra MCP Server - AI agent workspace interaction with file operations, code execution, and notebook management",
	}

	m.server = sdk.NewServer(impl, opts)

	// Register all tools
	m.registerFileTools(absWorkspace)
	m.registerCodeExecutionTools(absWorkspace)
	m.registerWorkspaceTools(absWorkspace)

	// Create SSE handler
	handler := sdk.NewSSEHandler(func(*http.Request) *sdk.Server {
		return m.server
	}, nil)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/mcp/result", m.handleMCPResult)

	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		log.Printf("[MCP] Starting MCP server on http://localhost:%d", port)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[MCP] HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (m *MCPServer) Stop() error {
	if m.httpServer != nil {
		log.Println("[MCP] Stopping MCP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()
		return m.httpServer.Shutdown(ctx)
	}
	return nil
}

type mcpResultPayload struct {
	RequestID string `json:"request_id"`
	Result    string `json:"result"`
}

func (m *MCPServer) handleMCPResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var payload mcpResultPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if payload.RequestID == "" {
		http.Error(w, "missing request_id", http.StatusBadRequest)
		return
	}

	m.deliverExecutionResult(payload.RequestID, payload.Result)
	w.WriteHeader(http.StatusNoContent)
}

// registerFileTools registers file operation tools
func (m *MCPServer) registerFileTools(workspace string) {
	// read_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "read_file",
		Description: "Operates on Idensyra workspace - Read the content of a file in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file relative to workspace root",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		content, err := m.dispatchUIAction("read_file", map[string]any{"path": path}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: content},
			},
		}, nil, nil
	})

	// write_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "write_file",
		Description: "Operates on Idensyra workspace - Write content to a file in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file relative to workspace root",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"path", "content"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)
		content := args["content"].(string)

		res, err := m.dispatchUIAction("write_file", map[string]any{"path": path, "content": content}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error updating file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// create_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "create_file",
		Description: "Create a new file in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the new file relative to workspace root",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Initial content of the file",
				},
			},
			"required": []string{"path", "content"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)
		content := args["content"].(string)

		res, err := m.dispatchUIAction("create_file", map[string]any{"path": path, "content": content}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// delete_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "delete_file",
		Description: "Delete a file from the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to delete relative to workspace root",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		res, err := m.dispatchUIAction("delete_file", map[string]any{"path": path}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error deleting file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// rename_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "rename_file",
		Description: "Rename or move a file in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"old_path": map[string]interface{}{
					"type":        "string",
					"description": "Current path of the file relative to workspace root",
				},
				"new_path": map[string]interface{}{
					"type":        "string",
					"description": "New path of the file relative to workspace root",
				},
			},
			"required": []string{"old_path", "new_path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		oldPath := args["old_path"].(string)
		newPath := args["new_path"].(string)

		res, err := m.dispatchUIAction("rename_file", map[string]any{"old_path": oldPath, "new_path": newPath}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error renaming file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// list_files tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "list_files",
		Description: "Operates on Idensyra workspace - List files in a directory in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the directory relative to workspace root (empty for root)",
				},
			},
			"required": []string{},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		dir := ""
		if p, ok := args["dir_path"].(string); ok {
			dir = p
		}

		res, err := m.dispatchUIAction("list_files", map[string]any{"dir_path": dir}, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error listing files via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// import_file_to_workspace tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "import_file_to_workspace",
		Description: "Import a specific file from the computer into the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the source file on the computer",
				},
				"target_dir": map[string]interface{}{
					"type":        "string",
					"description": "Target directory in workspace (empty string for root)",
				},
			},
			"required": []string{"source_path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		sourcePath := args["source_path"].(string)
		targetDir := ""
		if td, ok := args["target_dir"].(string); ok {
			targetDir = td
		}

		res, err := m.dispatchUIAction("import_file_to_workspace", map[string]any{"source_path": sourcePath, "target_dir": targetDir}, 60*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error importing file via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})
}

// registerCodeExecutionTools registers code execution tools
func (m *MCPServer) registerCodeExecutionTools(workspace string) {
	// execute_go_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "execute_go_file",
		Description: "Execute a Go file in the workspace using Yaegi interpreter",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the Go file relative to workspace root",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		// Switch to this file in UI (best-effort)
		_ = m.app.SetActiveFile(path)

		content, err := m.app.GetFileContent(path)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file: %v", err)
		}

		// Dispatch to frontend to run as if user pressed the Run button
		requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
		runtime.EventsEmit(m.app.ctx, "mcp:execute_go_file", map[string]any{"request_id": requestId, "path": path, "content": content})
		res, err := m.waitForExecutionResult(requestId, 30*time.Second)
		if err != nil {
			return &sdk.CallToolResult{
				Content: []sdk.Content{
					&sdk.TextContent{Text: fmt.Sprintf("Dispatched execute_go_file for %s to the frontend (no result): %v", path, err)},
				},
			}, nil, nil
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// execute_python_file tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "execute_python_file",
		Description: "Execute a Python file in the workspace",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the Python file relative to workspace root",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		// Switch to this file in UI
		_ = m.app.SetActiveFile(path)

		content, err := m.app.GetFileContent(path)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file: %v", err)
		}

		// Dispatch to frontend to run as if user pressed the Run button
		requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
		runtime.EventsEmit(m.app.ctx, "mcp:execute_python_file", map[string]any{"request_id": requestId, "path": path, "content": content})
		res, err := m.waitForExecutionResult(requestId, 30*time.Second)
		if err != nil {
			return &sdk.CallToolResult{
				Content: []sdk.Content{
					&sdk.TextContent{Text: fmt.Sprintf("Dispatched execute_python_file for %s to the frontend (no result): %v", path, err)},
				},
			}, nil, nil
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// execute_go_code tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "execute_go_code",
		Description: "Execute Go code directly without a file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "Go code to execute",
				},
			},
			"required": []string{"code"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		code := args["code"].(string)

		// Dispatch to frontend to run as if user pressed the Run button
		requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
		runtime.EventsEmit(m.app.ctx, "mcp:execute_go_code", map[string]any{"request_id": requestId, "code": code})
		res, err := m.waitForExecutionResult(requestId, 30*time.Second)
		if err != nil {
			return &sdk.CallToolResult{
				Content: []sdk.Content{
					&sdk.TextContent{Text: fmt.Sprintf("Dispatched execute_go_code to frontend (no result): %v", err)},
				},
			}, nil, nil
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// execute_python_code tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "execute_python_code",
		Description: "Execute Python code directly without a file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "Python code to execute",
				},
			},
			"required": []string{"code"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		code := args["code"].(string)

		// Use an in-memory temp name and dispatch to frontend for execution
		tmpFile := filepath.Join(workspace, fmt.Sprintf(".tmp_mcp_py_%d.py", os.Getpid()))
		requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
		runtime.EventsEmit(m.app.ctx, "mcp:execute_python_code", map[string]any{"request_id": requestId, "tmp_file": tmpFile, "code": code})
		res, err := m.waitForExecutionResult(requestId, 30*time.Second)
		if err != nil {
			return &sdk.CallToolResult{
				Content: []sdk.Content{
					&sdk.TextContent{Text: fmt.Sprintf("Dispatched execute_python_code to frontend (no result): %v", err)},
				},
			}, nil, nil
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})
}

// registerWorkspaceTools registers workspace management tools
func (m *MCPServer) registerWorkspaceTools(workspace string) {
	// open_workspace tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "open_workspace",
		Description: "Open a workspace directory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the workspace directory",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		res, err := m.dispatchUIAction("open_workspace", map[string]any{"path": path}, 60*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error opening workspace via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// save_workspace tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "save_workspace",
		Description: "Save the current temporary workspace to a specified path",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path where to save the workspace",
				},
			},
			"required": []string{"path"},
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		path := args["path"].(string)

		res, err := m.dispatchUIAction("save_workspace", map[string]any{"path": path}, 60*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error saving workspace via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// save_all_files tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "save_all_files",
		Description: "Save all unsaved changes in the workspace",
		InputSchema: map[string]interface{}{
			"type":                 "object",
			"properties":           map[string]interface{}{},
			"required":             []string{},
			"additionalProperties": true,
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		res, err := m.dispatchUIAction("save_all_files", nil, 60*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error saving files via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})

	// get_workspace_info tool
	sdk.AddTool(m.server, &sdk.Tool{
		Name:        "get_workspace_info",
		Description: "Operates on Idensyra workspace - Get information about the current workspace",
		InputSchema: map[string]interface{}{
			"type":                 "object",
			"properties":           map[string]interface{}{},
			"required":             []string{},
			"additionalProperties": true,
		},
	}, func(ctx context.Context, req *sdk.CallToolRequest, args map[string]interface{}) (*sdk.CallToolResult, any, error) {
		res, err := m.dispatchUIAction("get_workspace_info", nil, 30*time.Second)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting workspace info via UI: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: res},
			},
		}, nil, nil
	})
}

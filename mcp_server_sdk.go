package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer wraps the MCP SDK server for HTTP access via SSE
type MCPServer struct {
	server     *sdk.Server
	httpServer *http.Server
	app        *App
}

// NewMCPServer creates a new MCP server using the official SDK
func NewMCPServer(app *App) *MCPServer {
	return &MCPServer{
		app: app,
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

		// Switch to this file in UI (ignore error from UI layer)
		_ = m.app.SetActiveFile(path)

		content, err := m.app.GetFileContent(path)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file: %v", err)
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

		if err := m.app.UpdateFileContent(path, content); err != nil {
			return nil, nil, fmt.Errorf("error updating file: %v", err)
		}

		// Do not auto-save to disk here; keep file modified in workspace and let frontend decide when to persist.

		// Switch to this file in UI (ignore error)
		_ = m.app.SetActiveFile(path)

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("File updated in workspace (unsaved): %s", path)},
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

		if err := m.app.CreateNewFile(path); err != nil {
			return nil, nil, fmt.Errorf("error creating file: %v", err)
		}
		if content != "" {
			if err := m.app.UpdateFileContent(path, content); err != nil {
				return nil, nil, fmt.Errorf("error updating new file: %v", err)
			}
			// Do not call SaveFile here; do not require a persistent workspace to create files.
			// Let the frontend/user explicitly save when appropriate.
		}

		// Switch to this file in UI (ignore error)
		_ = m.app.SetActiveFile(path)

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("File created successfully: %s", path)},
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

		if err := m.app.DeleteFile(path); err != nil {
			return nil, nil, fmt.Errorf("error deleting file: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("File deleted successfully: %s", path)},
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

		if err := m.app.RenameFile(oldPath, newPath); err != nil {
			return nil, nil, fmt.Errorf("error renaming file: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("File renamed successfully: %s -> %s", oldPath, newPath)},
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

		files := m.app.GetWorkspaceFiles()
		var result string
		for _, f := range files {
			if dir == "" {
				if !strings.Contains(f.Name, "/") {
					if f.IsDir {
						result += fmt.Sprintf("[DIR]  %s\n", f.Name)
					} else {
						result += fmt.Sprintf("[FILE] %s (%d bytes)\n", f.Name, f.Size)
					}
				}
			} else {
				if strings.HasPrefix(f.Name, dir+"/") {
					rest := strings.TrimPrefix(f.Name, dir+"/")
					if !strings.Contains(rest, "/") {
						if f.IsDir {
							result += fmt.Sprintf("[DIR]  %s\n", rest)
						} else {
							result += fmt.Sprintf("[FILE] %s (%d bytes)\n", rest, f.Size)
						}
					}
				}
			}

		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: result},
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

		if err := m.app.ImportSpecificFileToWorkspace(sourcePath, targetDir); err != nil {
			return nil, nil, fmt.Errorf("error importing file: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("File imported successfully from %s", sourcePath)},
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

		// Switch to this file in UI
		_ = m.app.SetActiveFile(path)

		content, err := m.app.GetFileContent(path)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file: %v", err)
		}

		result := m.app.ExecuteCodeWithColorBG(content, "dark")

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: result},
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

		result, err := m.app.ExecutePythonFileContent(path, content)
		if err != nil {
			return nil, nil, fmt.Errorf("execution error: %v\nOutput: %s", err, result)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: result},
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
		result := m.app.ExecuteCodeWithColorBG(code, "dark")

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: result},
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

		// Use an in-memory temp name and pass the code directly to App
		tmpFile := filepath.Join(workspace, fmt.Sprintf(".tmp_mcp_py_%d.py", os.Getpid()))

		result, err := m.app.ExecutePythonFileContent(tmpFile, code)
		if err != nil {
			return nil, nil, fmt.Errorf("execution error: %v\nOutput: %s", err, result)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: result},
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

		if _, err := m.app.OpenWorkspaceAt(path); err != nil {
			return nil, nil, fmt.Errorf("error opening workspace: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("Workspace opened: %s", path)},
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

		if _, err := m.app.CreateWorkspaceAt(path); err != nil {
			return nil, nil, fmt.Errorf("error saving workspace: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: fmt.Sprintf("Workspace saved to: %s", path)},
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
		if err := m.app.SaveAllFiles(); err != nil {
			return nil, nil, fmt.Errorf("error saving files: %v", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: "All files saved successfully"},
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
		info := fmt.Sprintf("Workspace root: %s", workspace)

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: info},
			},
		}, nil, nil
	})
}

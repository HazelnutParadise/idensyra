package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CodeExecution provides code execution tools for MCP
type CodeExecution struct {
	config        *Config
	workspaceRoot string
	confirmFunc   func(operation, details string) bool
	executeGoFunc func(code string, colorBG string) string
	executePyFunc func(filePath string) (string, error)
}

// NewCodeExecution creates a new CodeExecution instance
func NewCodeExecution(
	config *Config,
	workspaceRoot string,
	confirmFunc func(operation, details string) bool,
	executeGoFunc func(code string, colorBG string) string,
	executePyFunc func(filePath string) (string, error),
) *CodeExecution {
	return &CodeExecution{
		config:        config,
		workspaceRoot: workspaceRoot,
		confirmFunc:   confirmFunc,
		executeGoFunc: executeGoFunc,
		executePyFunc: executePyFunc,
	}
}

// ExecuteGoFile executes a Go file using the Yaegi interpreter
func (ce *CodeExecution) ExecuteGoFile(ctx context.Context, path string) (*ToolResponse, error) {
	if ce.config.ExecuteGo == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Go execution permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if ce.config.ExecuteGo == PermissionAsk && ce.confirmFunc != nil {
		if !ce.confirmFunc("Execute Go", fmt.Sprintf("Execute Go file: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Go execution cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(ce.workspaceRoot, path)
	
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", err)}},
			IsError: true,
		}, err
	}

	code := string(content)
	// Strip package main and import statements if present
	// This allows executing file content that has package declaration
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "package main") {
		// Remove "package main" line
		lines := strings.Split(code, "\n")
		if len(lines) > 0 {
			code = strings.Join(lines[1:], "\n")
			code = strings.TrimSpace(code)
		}
	}
	
	// Execute using the provided Go execution function
	if ce.executeGoFunc == nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Go execution function not available"}},
			IsError: true,
		}, fmt.Errorf("execution function not available")
	}

	result := ce.executeGoFunc(code, "dark")

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// ExecuteGoCode executes Go code directly
func (ce *CodeExecution) ExecuteGoCode(ctx context.Context, code string) (*ToolResponse, error) {
	if ce.config.ExecuteGo == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Go execution permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if ce.config.ExecuteGo == PermissionAsk && ce.confirmFunc != nil {
		if !ce.confirmFunc("Execute Go", "Execute Go code") {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Go execution cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	if ce.executeGoFunc == nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Go execution function not available"}},
			IsError: true,
		}, fmt.Errorf("execution function not available")
	}

	result := ce.executeGoFunc(code, "dark")

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// ExecutePythonFile executes a Python file
func (ce *CodeExecution) ExecutePythonFile(ctx context.Context, path string) (*ToolResponse, error) {
	if ce.config.ExecutePython == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Python execution permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if ce.config.ExecutePython == PermissionAsk && ce.confirmFunc != nil {
		if !ce.confirmFunc("Execute Python", fmt.Sprintf("Execute Python file: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Python execution cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(ce.workspaceRoot, path)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File not found: %s", path)}},
			IsError: true,
		}, err
	}

	// Execute using the provided Python execution function
	if ce.executePyFunc != nil {
		result, err := ce.executePyFunc(fullPath)
		if err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, result)}},
				IsError: true,
			}, err
		}
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: result}},
		}, nil
	}

	// Fallback to direct python3 execution
	cmd := exec.CommandContext(ctx, "python3", fullPath)
	cmd.Dir = ce.workspaceRoot
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, string(output))}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: string(output)}},
	}, nil
}

// ExecutePythonCode executes Python code directly
func (ce *CodeExecution) ExecutePythonCode(ctx context.Context, code string) (*ToolResponse, error) {
	if ce.config.ExecutePython == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Python execution permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if ce.config.ExecutePython == PermissionAsk && ce.confirmFunc != nil {
		if !ce.confirmFunc("Execute Python", "Execute Python code") {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Python execution cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	// Create a temporary file
	tmpFile := filepath.Join(ce.workspaceRoot, fmt.Sprintf(".tmp_py_%d.py", time.Now().UnixNano()))
	defer os.Remove(tmpFile)
	
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating temp file: %v", err)}},
			IsError: true,
		}, err
	}

	// Execute using the provided Python execution function
	if ce.executePyFunc != nil {
		result, err := ce.executePyFunc(tmpFile)
		if err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, result)}},
				IsError: true,
			}, err
		}
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: result}},
		}, nil
	}

	// Fallback to direct python3 execution
	cmd := exec.CommandContext(ctx, "python3", tmpFile)
	cmd.Dir = ce.workspaceRoot
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, string(output))}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: string(output)}},
	}, nil
}

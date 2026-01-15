package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CodeExecution provides code execution tools for MCP
type CodeExecution struct {
	config               *Config
	workspaceRoot        string
	confirmFunc          func(operation, details string) bool
	executeGoFunc        func(code string, colorBG string) string
	executePyFunc        func(filePath string) (string, error)
	executePyContentFunc func(filename string, content string) (string, error)
	readFileFunc         func(path string) (string, error)
	setActiveFileFunc    func(path string) error
}

// NewCodeExecution creates a new CodeExecution instance
func NewCodeExecution(
	config *Config,
	workspaceRoot string,
	confirmFunc func(operation, details string) bool,
	executeGoFunc func(code string, colorBG string) string,
	executePyFunc func(filePath string) (string, error),
	executePyContentFunc func(filename string, content string) (string, error),
	readFileFunc func(path string) (string, error),
	setActiveFileFunc func(path string) error,
) *CodeExecution {
	return &CodeExecution{
		config:               config,
		workspaceRoot:        workspaceRoot,
		confirmFunc:          confirmFunc,
		executeGoFunc:        executeGoFunc,
		executePyFunc:        executePyFunc,
		executePyContentFunc: executePyContentFunc,
		readFileFunc:         readFileFunc,
		setActiveFileFunc:    setActiveFileFunc,
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

	// Require read backend to fetch file content; do not depend on workspace being created
	if ce.readFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Read backend not available"}}, IsError: true}, fmt.Errorf("read backend not available")
	}

	content, err := ce.readFileFunc(path)
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

	// Switch to the file being executed
	if ce.setActiveFileFunc != nil {
		_ = ce.setActiveFileFunc(path)
	}

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

	// Prefer content-based execution via callback (does not require workspace)
	if ce.readFileFunc != nil && ce.executePyContentFunc != nil {
		content, err := ce.readFileFunc(path)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", err)}}, IsError: true}, err
		}
		res, err := ce.executePyContentFunc(path, content)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, res)}}, IsError: true}, err
		}
		if ce.setActiveFileFunc != nil {
			_ = ce.setActiveFileFunc(path)
		}
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
	}

	// If we have a readFileFunc and a file-based executor, write to system temp and call it
	if ce.readFileFunc != nil && ce.executePyFunc != nil {
		content, err := ce.readFileFunc(path)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", err)}}, IsError: true}, err
		}
		tmp, err := os.CreateTemp("", "mcp_py_*.py")
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating temp file: %v", err)}}, IsError: true}, err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.WriteString(content); err != nil {
			tmp.Close()
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing temp file: %v", err)}}, IsError: true}, err
		}
		tmp.Close()
		res, err := ce.executePyFunc(tmp.Name())
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, res)}}, IsError: true}, err
		}
		if ce.setActiveFileFunc != nil {
			_ = ce.setActiveFileFunc(path)
		}
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
	}

	// If a file-based executor exists and workspaceRoot is available, try direct invocation
	if ce.executePyFunc != nil && ce.workspaceRoot != "" {
		fullPath := filepath.Join(ce.workspaceRoot, path)
		res, err := ce.executePyFunc(fullPath)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, res)}}, IsError: true}, err
		}
		if ce.setActiveFileFunc != nil {
			_ = ce.setActiveFileFunc(path)
		}
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Python execution backend not available"}}, IsError: true}, fmt.Errorf("python execution backend not available")
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

	// Prefer content-based executor callback
	if ce.executePyContentFunc != nil {
		res, err := ce.executePyContentFunc(".tmp_mcp.py", code)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, res)}}, IsError: true}, err
		}
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
	}

	// If only file-based executor is available, write to system temp and call it
	if ce.executePyFunc != nil {
		tmp, err := os.CreateTemp("", "mcp_py_*.py")
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating temp file: %v", err)}}, IsError: true}, err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.WriteString(code); err != nil {
			tmp.Close()
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing temp file: %v", err)}}, IsError: true}, err
		}
		tmp.Close()
		res, err := ce.executePyFunc(tmp.Name())
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing Python: %v\n%s", err, res)}}, IsError: true}, err
		}
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Python execution backend not available"}}, IsError: true}, fmt.Errorf("python execution backend not available")
}

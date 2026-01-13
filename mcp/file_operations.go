package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// FileOperations provides file manipulation tools for MCP
type FileOperations struct {
	config           *Config
	workspaceRoot    string
	confirmFunc      func(operation, details string) bool
	setActiveFileFunc func(path string) error
}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations(config *Config, workspaceRoot string, confirmFunc func(operation, details string) bool, setActiveFileFunc func(path string) error) *FileOperations {
	return &FileOperations{
		config:           config,
		workspaceRoot:    workspaceRoot,
		confirmFunc:      confirmFunc,
		setActiveFileFunc: setActiveFileFunc,
	}
}

// ReadFile reads the content of a file
func (fo *FileOperations) ReadFile(ctx context.Context, path string) (*ToolResponse, error) {
	fullPath := filepath.Join(fo.workspaceRoot, path)
	
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the file being read
	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: string(content)}},
	}, nil
}

// WriteFile writes content to a file
func (fo *FileOperations) WriteFile(ctx context.Context, path string, content string) (*ToolResponse, error) {
	if fo.config.FileEdit == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "File edit permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if fo.config.FileEdit == PermissionAsk && fo.confirmFunc != nil {
		if !fo.confirmFunc("File Edit", fmt.Sprintf("Edit file: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "File edit cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(fo.workspaceRoot, path)
	
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing file: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the file being edited
	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File written successfully: %s", path)}},
	}, nil
}

// CreateFile creates a new file
func (fo *FileOperations) CreateFile(ctx context.Context, path string, content string) (*ToolResponse, error) {
	if fo.config.FileCreate == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "File create permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if fo.config.FileCreate == PermissionAsk && fo.confirmFunc != nil {
		if !fo.confirmFunc("File Create", fmt.Sprintf("Create file: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "File create cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(fo.workspaceRoot, path)
	
	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File already exists: %s", path)}},
			IsError: true,
		}, fmt.Errorf("file already exists")
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating file: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the newly created file
	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File created successfully: %s", path)}},
	}, nil
}

// DeleteFile deletes a file
func (fo *FileOperations) DeleteFile(ctx context.Context, path string) (*ToolResponse, error) {
	if fo.config.FileDelete == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "File delete permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if fo.config.FileDelete == PermissionAsk && fo.confirmFunc != nil {
		if !fo.confirmFunc("File Delete", fmt.Sprintf("Delete file: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "File delete cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(fo.workspaceRoot, path)
	
	if err := os.Remove(fullPath); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error deleting file: %v", err)}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File deleted successfully: %s", path)}},
	}, nil
}

// RenameFile renames a file
func (fo *FileOperations) RenameFile(ctx context.Context, oldPath string, newPath string) (*ToolResponse, error) {
	if fo.config.FileRename == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "File rename permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if fo.config.FileRename == PermissionAsk && fo.confirmFunc != nil {
		if !fo.confirmFunc("File Rename", fmt.Sprintf("Rename file: %s -> %s", oldPath, newPath)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "File rename cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	oldFullPath := filepath.Join(fo.workspaceRoot, oldPath)
	newFullPath := filepath.Join(fo.workspaceRoot, newPath)
	
	// Create parent directories for new path if needed
	if err := os.MkdirAll(filepath.Dir(newFullPath), 0755); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, err
	}

	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error renaming file: %v", err)}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File renamed successfully: %s -> %s", oldPath, newPath)}},
	}, nil
}

// ListFiles lists all files in a directory
func (fo *FileOperations) ListFiles(ctx context.Context, dirPath string) (*ToolResponse, error) {
	fullPath := filepath.Join(fo.workspaceRoot, dirPath)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error listing files: %v", err)}},
			IsError: true,
		}, err
	}

	var result string
	for _, entry := range entries {
		if entry.IsDir() {
			result += fmt.Sprintf("[DIR]  %s\n", entry.Name())
		} else {
			info, _ := entry.Info()
			result += fmt.Sprintf("[FILE] %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

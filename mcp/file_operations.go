package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOperations provides file manipulation tools for MCP
type FileOperations struct {
	config            *Config
	workspaceRoot     string
	confirmFunc       func(operation, details string) bool
	setActiveFileFunc func(path string) error

	// Optional backend callbacks (use these to call frontend App APIs)
	readFileFunc   func(path string) (string, error)
	writeFileFunc  func(path string, content string) error
	createFileFunc func(path string, content string) error
	deleteFileFunc func(path string) error
	renameFileFunc func(oldPath, newPath string) error
	listFilesFunc  func(dirPath string) (string, error)
}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations(
	config *Config,
	workspaceRoot string,
	confirmFunc func(operation, details string) bool,
	setActiveFileFunc func(path string) error,
	readFileFunc func(path string) (string, error),
	writeFileFunc func(path string, content string) error,
	createFileFunc func(path string, content string) error,
	deleteFileFunc func(path string) error,
	renameFileFunc func(oldPath, newPath string) error,
	listFilesFunc func(dirPath string) (string, error),
) *FileOperations {
	return &FileOperations{
		config:            config,
		workspaceRoot:     workspaceRoot,
		confirmFunc:       confirmFunc,
		setActiveFileFunc: setActiveFileFunc,
		readFileFunc:      readFileFunc,
		writeFileFunc:     writeFileFunc,
		createFileFunc:    createFileFunc,
		deleteFileFunc:    deleteFileFunc,
		renameFileFunc:    renameFileFunc,
		listFilesFunc:     listFilesFunc,
	}
}

// helper: validate and clean a relative path (rejects absolute and ..)
func safeCleanRelativePath(input string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(input))
	if clean == "." || clean == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", input)
	}
	for _, part := range strings.Split(clean, string(os.PathSeparator)) {
		if part == ".." {
			return "", fmt.Errorf("invalid path: %s", input)
		}
	}
	return filepath.ToSlash(clean), nil
}

// ReadFile reads the content of a file
func (fo *FileOperations) ReadFile(ctx context.Context, path string) (*ToolResponse, error) {
	cleanPath, err := safeCleanRelativePath(path)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid path: %v", err)}}, IsError: true}, err
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.readFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Read backend not available"}}, IsError: true}, fmt.Errorf("read backend not available")
	}

	content, err := fo.readFileFunc(cleanPath)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", err)}}, IsError: true}, err
	}

	// Switch to the file being read
	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(cleanPath)
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: content}}}, nil
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

	cleanPath, err := safeCleanRelativePath(path)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid path: %v", err)}}, IsError: true}, err
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.writeFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Write backend not available"}}, IsError: true}, fmt.Errorf("write backend not available")
	}

	if err := fo.writeFileFunc(cleanPath, content); err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing file: %v", err)}}, IsError: true}, err
	}

	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(cleanPath)
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File written successfully: %s", cleanPath)}}}, nil
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

	cleanPath, err := safeCleanRelativePath(path)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid path: %v", err)}}, IsError: true}, err
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.createFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Create backend not available"}}, IsError: true}, fmt.Errorf("create backend not available")
	}

	if err := fo.createFileFunc(cleanPath, content); err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating file: %v", err)}}, IsError: true}, err
	}

	if fo.setActiveFileFunc != nil {
		_ = fo.setActiveFileFunc(cleanPath)
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File created successfully: %s", cleanPath)}}}, nil
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

	cleanPath, err := safeCleanRelativePath(path)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid path: %v", err)}}, IsError: true}, err
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.deleteFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Delete backend not available"}}, IsError: true}, fmt.Errorf("delete backend not available")
	}

	if err := fo.deleteFileFunc(cleanPath); err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error deleting file: %v", err)}}, IsError: true}, err
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File deleted successfully: %s", cleanPath)}}}, nil
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

	// Validate paths
	cleanOld, err := safeCleanRelativePath(oldPath)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid old path: %v", err)}}, IsError: true}, err
	}
	cleanNew, err := safeCleanRelativePath(newPath)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid new path: %v", err)}}, IsError: true}, err
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.renameFileFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "Rename backend not available"}}, IsError: true}, fmt.Errorf("rename backend not available")
	}

	if err := fo.renameFileFunc(cleanOld, cleanNew); err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error renaming file: %v", err)}}, IsError: true}, err
	}

	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File renamed successfully: %s -> %s", cleanOld, cleanNew)}}}, nil
}

// ListFiles lists all files in a directory
func (fo *FileOperations) ListFiles(ctx context.Context, dirPath string) (*ToolResponse, error) {
	cleanDir := ""
	if strings.TrimSpace(dirPath) != "" {
		cd, err := safeCleanRelativePath(dirPath)
		if err != nil {
			return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid path: %v", err)}}, IsError: true}, err
		}
		cleanDir = cd
	}

	// Require backend callback; do NOT fallback to filesystem
	if fo.listFilesFunc == nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: "List backend not available"}}, IsError: true}, fmt.Errorf("list backend not available")
	}

	res, err := fo.listFilesFunc(cleanDir)
	if err != nil {
		return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error listing files: %v", err)}}, IsError: true}, err
	}
	return &ToolResponse{Content: []ContentBlock{{Type: "text", Text: res}}}, nil
}

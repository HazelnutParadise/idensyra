package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// WorkspaceManagement provides workspace management tools for MCP
type WorkspaceManagement struct {
	config            *Config
	confirmFunc       func(operation, details string) bool
	currentWorkspace  string
	openWorkspaceFunc func(path string) error
	saveWorkspaceFunc func(path string) error
	saveChangesFunc   func() error
	importFileFunc    func(sourcePath, targetDir string) error
}

// NewWorkspaceManagement creates a new WorkspaceManagement instance
func NewWorkspaceManagement(
	config *Config,
	confirmFunc func(operation, details string) bool,
	openWorkspaceFunc func(path string) error,
	saveWorkspaceFunc func(path string) error,
	saveChangesFunc func() error,
	importFileFunc func(sourcePath, targetDir string) error,
) *WorkspaceManagement {
	return &WorkspaceManagement{
		config:            config,
		confirmFunc:       confirmFunc,
		openWorkspaceFunc: openWorkspaceFunc,
		saveWorkspaceFunc: saveWorkspaceFunc,
		saveChangesFunc:   saveChangesFunc,
		importFileFunc:    importFileFunc,
	}
}

// OpenWorkspace opens a workspace directory
func (wm *WorkspaceManagement) OpenWorkspace(ctx context.Context, path string) (*ToolResponse, error) {
	if wm.config.WorkspaceOpen == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Workspace open permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if wm.config.WorkspaceOpen == PermissionAsk && wm.confirmFunc != nil {
		if !wm.confirmFunc("Workspace Open", fmt.Sprintf("Open workspace: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Workspace open cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error accessing directory: %v", err)}},
			IsError: true,
		}, err
	}

	if !info.IsDir() {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Path is not a directory: %s", path)}},
			IsError: true,
		}, fmt.Errorf("not a directory")
	}

	// Use the provided open workspace function if available
	if wm.openWorkspaceFunc != nil {
		if err := wm.openWorkspaceFunc(path); err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error opening workspace: %v", err)}},
				IsError: true,
			}, err
		}
	}

	wm.currentWorkspace = path

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Workspace opened successfully: %s", path)}},
	}, nil
}

// SaveTempWorkspace saves the temporary workspace to a specified path
func (wm *WorkspaceManagement) SaveTempWorkspace(ctx context.Context, targetPath string) (*ToolResponse, error) {
	if wm.config.WorkspaceSave == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Workspace save permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if wm.config.WorkspaceSave == PermissionAsk && wm.confirmFunc != nil {
		if !wm.confirmFunc("Workspace Save", fmt.Sprintf("Save workspace to: %s", targetPath)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Workspace save cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	// Create the target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, err
	}

	// Use the provided save workspace function if available
	if wm.saveWorkspaceFunc != nil {
		if err := wm.saveWorkspaceFunc(targetPath); err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error saving workspace: %v", err)}},
				IsError: true,
			}, err
		}
	}

	wm.currentWorkspace = targetPath

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Workspace saved successfully to: %s", targetPath)}},
	}, nil
}

// SaveChanges saves all unsaved changes in the current workspace
func (wm *WorkspaceManagement) SaveChanges(ctx context.Context) (*ToolResponse, error) {
	if wm.config.WorkspaceModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Workspace save changes permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if wm.config.WorkspaceModify == PermissionAsk && wm.confirmFunc != nil {
		if !wm.confirmFunc("Save Changes", "Save all unsaved changes in the workspace") {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Save changes cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	// Use the provided save changes function if available
	if wm.saveChangesFunc != nil {
		if err := wm.saveChangesFunc(); err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error saving changes: %v", err)}},
				IsError: true,
			}, err
		}
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: "All changes saved successfully"}},
	}, nil
}

// GetWorkspaceInfo returns information about the current workspace
func (wm *WorkspaceManagement) GetWorkspaceInfo(ctx context.Context) (*ToolResponse, error) {
	if wm.currentWorkspace == "" {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "No workspace is currently open"}},
		}, nil
	}

	info, err := os.Stat(wm.currentWorkspace)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error accessing workspace: %v", err)}},
			IsError: true,
		}, err
	}

	// Count files in workspace
	fileCount := 0
	err = filepath.Walk(wm.currentWorkspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})

	result := fmt.Sprintf("Workspace: %s\nExists: %v\nFiles: %d", wm.currentWorkspace, info.IsDir(), fileCount)

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// CreateWorkspaceDirectory creates a new directory in the workspace
func (wm *WorkspaceManagement) CreateWorkspaceDirectory(ctx context.Context, relativePath string) (*ToolResponse, error) {
	if wm.config.WorkspaceModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Workspace modify permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if wm.config.WorkspaceModify == PermissionAsk && wm.confirmFunc != nil {
		if !wm.confirmFunc("Create Directory", fmt.Sprintf("Create directory: %s", relativePath)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Create directory cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	if wm.currentWorkspace == "" {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "No workspace is currently open"}},
			IsError: true,
		}, fmt.Errorf("no workspace open")
	}

	fullPath := filepath.Join(wm.currentWorkspace, relativePath)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Directory created successfully: %s", relativePath)}},
	}, nil
}

// ImportFileToWorkspace imports an external file into the workspace
func (wm *WorkspaceManagement) ImportFileToWorkspace(ctx context.Context, sourcePath, targetDir string) (*ToolResponse, error) {
	if wm.config.WorkspaceModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Workspace modify permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if wm.config.WorkspaceModify == PermissionAsk && wm.confirmFunc != nil {
		if !wm.confirmFunc("Import File", fmt.Sprintf("Import file: %s to directory: %s", sourcePath, targetDir)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Import file cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	// Check if source file exists
	if _, err := os.Stat(sourcePath); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Source file does not exist: %v", err)}},
			IsError: true,
		}, err
	}

	// Use the provided import file function if available
	if wm.importFileFunc != nil {
		if err := wm.importFileFunc(sourcePath, targetDir); err != nil {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error importing file: %v", err)}},
				IsError: true,
			}, err
		}
	} else {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Import file function not available"}},
			IsError: true,
		}, fmt.Errorf("import file function not available")
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("File imported successfully: %s", filepath.Base(sourcePath))}},
	}, nil
}

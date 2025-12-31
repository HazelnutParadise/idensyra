package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WorkspaceFile represents a single file in the workspace
type WorkspaceFile struct {
	Name         string `json:"name"`
	Content      string `json:"content"`
	Modified     bool   `json:"modified"`
	SavedContent string `json:"-"`
	IsNew        bool   `json:"-"`
}

// Workspace manages a workspace with multiple Go files
type Workspace struct {
	mu          sync.RWMutex
	workDir     string // Actual workspace directory (or temp if not set)
	isTemp      bool   // true if using temp directory
	files       map[string]*WorkspaceFile
	activeFile  string
	initialized bool
	modified    bool
}

var globalWorkspace *Workspace

// InitWorkspace initializes the workspace with a temporary directory
func (a *App) InitWorkspace() error {
	if globalWorkspace != nil && globalWorkspace.initialized {
		return nil // Already initialized
	}

	tempDir, err := os.MkdirTemp("", "idensyra-workspace-*")
	if err != nil {
		return fmt.Errorf("failed to create temp workspace: %w", err)
	}

	globalWorkspace = &Workspace{
		workDir:     tempDir,
		isTemp:      true,
		files:       make(map[string]*WorkspaceFile),
		initialized: true,
	}

	// Create default main.go file
	defaultFile := &WorkspaceFile{
		Name:         "main.go",
		Content:      defaultCode,
		SavedContent: defaultCode,
		IsNew:        false,
		Modified:     false,
	}
	globalWorkspace.files["main.go"] = defaultFile
	globalWorkspace.activeFile = "main.go"

	// Write default file to directory
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(preCode+"\n"+defaultCode+"\n"+endCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write default file: %w", err)
	}

	return nil
}

// GetWorkspaceFiles returns all files in the workspace, sorted by name
func (a *App) GetWorkspaceFiles() []WorkspaceFile {
	if globalWorkspace == nil {
		return []WorkspaceFile{}
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	refreshWorkspaceFromDiskLocked()

	files := make([]WorkspaceFile, 0, len(globalWorkspace.files))
	for _, file := range globalWorkspace.files {
		files = append(files, *file)
	}

	// Sort files by name for consistent ordering
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files
}

func refreshWorkspaceFromDiskLocked() {
	if globalWorkspace == nil || !globalWorkspace.initialized {
		return
	}

	dirPath := globalWorkspace.workDir
	if dirPath == "" {
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	diskFiles := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		diskFiles[name] = struct{}{}

		filePath := filepath.Join(dirPath, name)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var contentStr string
		if isImageFile(name) {
			contentStr = base64.StdEncoding.EncodeToString(content)
		} else {
			contentStr = string(content)
			if strings.HasSuffix(name, ".go") {
				contentStr = strings.TrimPrefix(contentStr, preCode+"\n")
				contentStr = strings.TrimSuffix(contentStr, "\n"+endCode)
			}
		}

		if existing, exists := globalWorkspace.files[name]; exists {
			if existing.Modified || existing.IsNew {
				continue
			}
			if existing.Content != contentStr {
				existing.Content = contentStr
				existing.SavedContent = contentStr
				existing.Modified = false
				existing.IsNew = false
			}
			continue
		}

		globalWorkspace.files[name] = &WorkspaceFile{
			Name:         name,
			Content:      contentStr,
			SavedContent: contentStr,
			IsNew:        false,
			Modified:     false,
		}
	}

	for name, file := range globalWorkspace.files {
		if _, exists := diskFiles[name]; exists {
			continue
		}
		if file.Modified || file.IsNew {
			continue
		}
		delete(globalWorkspace.files, name)
	}

	if globalWorkspace.activeFile != "" {
		if _, exists := globalWorkspace.files[globalWorkspace.activeFile]; !exists {
			globalWorkspace.activeFile = ""
		}
	}

	if globalWorkspace.activeFile == "" {
		for name := range globalWorkspace.files {
			globalWorkspace.activeFile = name
			break
		}
	}

	updateWorkspaceModifiedLocked()
}

// GetActiveFile returns the name of the currently active file
func (a *App) GetActiveFile() string {
	if globalWorkspace == nil {
		return ""
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	return globalWorkspace.activeFile
}

// SetActiveFile sets the active file in the workspace
func (a *App) SetActiveFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[filename]; !exists {
		return fmt.Errorf("file not found: %s", filename)
	}

	globalWorkspace.activeFile = filename
	return nil
}

// GetFileContent returns the content of a specific file
// For image files, returns base64 encoded data
func (a *App) GetFileContent(filename string) (string, error) {
	if globalWorkspace == nil {
		return "", fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	file, exists := globalWorkspace.files[filename]
	if !exists {
		return "", fmt.Errorf("file not found: %s", filename)
	}

	return file.Content, nil
}

// isImageFile checks if a file is an image based on extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// UpdateFileContent updates the content of a file
func (a *App) UpdateFileContent(filename string, content string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	file, exists := globalWorkspace.files[filename]
	if !exists {
		return fmt.Errorf("file not found: %s", filename)
	}

	if file.Content == content {
		return nil
	}

	file.Content = content
	file.Modified = file.IsNew || file.Content != file.SavedContent
	updateWorkspaceModifiedLocked()

	return nil
}

// SaveFile saves a file to disk (in workspace directory)
func (a *App) SaveFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// If using temp workspace, prompt to open/create workspace
	if globalWorkspace.isTemp {
		return fmt.Errorf("temporary workspace: please open or create a workspace first")
	}

	file, exists := globalWorkspace.files[filename]
	if !exists {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Only wrap with preCode/endCode for .go files
	var fullContent []byte
	if isImageFile(filename) {
		// Decode base64 for image files
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		fullContent = decoded
	} else if strings.HasSuffix(filename, ".go") {
		fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
	} else {
		fullContent = []byte(file.Content)
	}

	filePath := filepath.Join(globalWorkspace.workDir, filename)
	err := os.WriteFile(filePath, fullContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	file.Modified = false
	file.IsNew = false
	file.SavedContent = file.Content

	updateWorkspaceModifiedLocked()

	return nil
}

// SaveAllFiles saves all files in the workspace
func (a *App) SaveAllFiles() error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// If using temp workspace, prompt to open/create workspace
	if globalWorkspace.isTemp {
		return fmt.Errorf("temporary workspace: please open or create a workspace first")
	}

	for filename, file := range globalWorkspace.files {
		// Only wrap with preCode/endCode for .go files
		var fullContent []byte
		if isImageFile(filename) {
			// Decode base64 for image files
			decoded, err := base64.StdEncoding.DecodeString(file.Content)
			if err != nil {
				return fmt.Errorf("failed to decode image %s: %w", filename, err)
			}
			fullContent = decoded
		} else if strings.HasSuffix(filename, ".go") {
			fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
		} else {
			fullContent = []byte(file.Content)
		}

		filePath := filepath.Join(globalWorkspace.workDir, filename)
		err := os.WriteFile(filePath, fullContent, 0644)
		if err != nil {
			return fmt.Errorf("failed to save file %s: %w", filename, err)
		}
		file.Modified = false
		file.IsNew = false
		file.SavedContent = file.Content
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// CreateNewFile creates a new Go file in the workspace with the default template
func (a *App) CreateNewFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	// No longer enforce .go extension - allow any filename

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[filename]; exists {
		return fmt.Errorf("file already exists: %s", filename)
	}

	// Create new file with appropriate default content based on extension
	var content string
	if strings.HasSuffix(filename, ".go") {
		content = defaultCode
	} else {
		content = "" // Empty content for non-Go files
	}

	newFile := &WorkspaceFile{
		Name:         filename,
		Content:      content,
		SavedContent: "",
		IsNew:        true,
		Modified:     true,
	}
	globalWorkspace.files[filename] = newFile
	updateWorkspaceModifiedLocked()

	return nil
}

// DeleteFile removes a file from the workspace
func (a *App) DeleteFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// Don't allow deleting the last file
	if len(globalWorkspace.files) <= 1 {
		return fmt.Errorf("cannot delete the last file in workspace")
	}

	if _, exists := globalWorkspace.files[filename]; !exists {
		return fmt.Errorf("file not found: %s", filename)
	}

	delete(globalWorkspace.files, filename)

	// Remove from disk regardless of temp workspace to prevent re-adding on refresh
	if globalWorkspace.workDir != "" {
		filePath := filepath.Join(globalWorkspace.workDir, filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// If deleted file was active, switch to first available file
	if globalWorkspace.activeFile == filename {
		for name := range globalWorkspace.files {
			globalWorkspace.activeFile = name
			break
		}
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// OpenWorkspace opens an existing workspace from a directory
func (a *App) OpenWorkspace() (string, error) {
	if globalWorkspace == nil {
		return "", fmt.Errorf("workspace not initialized")
	}

	// Prompt user to select workspace directory
	selectedPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Workspace Folder",
	})
	if err != nil {
		return "", fmt.Errorf("failed to select directory: %w", err)
	}
	if selectedPath == "" {
		return "", nil // User cancelled
	}

	dirPath := selectedPath

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dirPath)
	}

	// Find all files in the directory (skip directories and hidden files)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	workspaceFiles := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		var contentStr string

		// For image files, encode as base64
		if isImageFile(entry.Name()) {
			contentStr = base64.StdEncoding.EncodeToString(content)
		} else {
			contentStr = string(content)
			// Only remove preCode/endCode for .go files
			if strings.HasSuffix(entry.Name(), ".go") {
				contentStr = strings.TrimPrefix(contentStr, preCode+"\n")
				contentStr = strings.TrimSuffix(contentStr, "\n"+endCode)
			}
		}

		workspaceFiles[entry.Name()] = contentStr
	}

	if len(workspaceFiles) == 0 {
		return "", fmt.Errorf("no files found in selected directory")
	}

	// Clean up old temp workspace if it was temp
	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if globalWorkspace.isTemp {
		os.RemoveAll(globalWorkspace.workDir)
	}

	// Set new workspace directory
	globalWorkspace.workDir = dirPath
	globalWorkspace.isTemp = false
	globalWorkspace.files = make(map[string]*WorkspaceFile)

	for filename, content := range workspaceFiles {
		globalWorkspace.files[filename] = &WorkspaceFile{
			Name:         filename,
			Content:      content,
			SavedContent: content,
			IsNew:        false,
			Modified:     false,
		}
	}

	// Set first file as active
	for name := range globalWorkspace.files {
		globalWorkspace.activeFile = name
		break
	}

	globalWorkspace.modified = false

	return dirPath, nil
}

// CreateWorkspace creates a new workspace in a user-selected directory
func (a *App) CreateWorkspace() (string, error) {
	if globalWorkspace == nil {
		return "", fmt.Errorf("workspace not initialized")
	}

	// Let user select where to create workspace
	selectedPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Create New Workspace (select location and enter folder name)",
		DefaultFilename: "my-idensyra-project",
	})

	if err != nil {
		return "", fmt.Errorf("failed to select location: %w", err)
	}
	if selectedPath == "" {
		return "", nil // User cancelled
	}

	// Create directory
	err = os.MkdirAll(selectedPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Clean up old temp workspace if it was temp
	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if globalWorkspace.isTemp {
		os.RemoveAll(globalWorkspace.workDir)
	}

	// Set new workspace directory
	globalWorkspace.workDir = selectedPath
	globalWorkspace.isTemp = false

	// Save all current files to new workspace
	for filename, file := range globalWorkspace.files {
		fullContent := preCode + "\n" + file.Content + "\n" + endCode
		filePath := filepath.Join(selectedPath, filename)
		err = os.WriteFile(filePath, []byte(fullContent), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write file %s: %w", filename, err)
		}
		file.Modified = false
		file.IsNew = false
		file.SavedContent = file.Content
	}

	updateWorkspaceModifiedLocked()

	return selectedPath, nil
}

// ImportFileToWorkspace imports an external file into the workspace
func (a *App) ImportFileToWorkspace() error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	// Let user select a file
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import File to Workspace",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
			{
				DisplayName: "Go Files (*.go)",
				Pattern:     "*.go",
			},
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "Markdown Files (*.md)",
				Pattern:     "*.md",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}
	if filePath == "" {
		return nil // User cancelled
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Get base filename
	filename := filepath.Base(filePath)

	// Keep original filename as-is

	// Check if file already exists
	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[filename]; exists {
		// Generate unique name
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		counter := 1
		for {
			newName := fmt.Sprintf("%s_%d%s", base, counter, ext)
			if _, exists := globalWorkspace.files[newName]; !exists {
				filename = newName
				break
			}
			counter++
		}
	}

	// Convert content to string (base64 for images)
	var contentStr string
	if isImageFile(filename) {
		contentStr = base64.StdEncoding.EncodeToString(content)
	} else {
		contentStr = string(content)
	}

	globalWorkspace.files[filename] = &WorkspaceFile{
		Name:         filename,
		Content:      contentStr,
		SavedContent: "",
		IsNew:        true,
		Modified:     true,
	}
	updateWorkspaceModifiedLocked()

	return nil
}

// ExportCurrentFile exports the current active file to a user-selected location
func (a *App) ExportCurrentFile() error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	if globalWorkspace.activeFile == "" {
		return fmt.Errorf("no active file")
	}

	file, exists := globalWorkspace.files[globalWorkspace.activeFile]
	if !exists {
		return fmt.Errorf("active file not found")
	}

	// Let user select where to save
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export File",
		DefaultFilename: globalWorkspace.activeFile,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	if filename == "" {
		return nil // User cancelled
	}

	// Write file content based on file type
	var fullContent []byte
	if isImageFile(globalWorkspace.activeFile) {
		// Decode base64 for image files
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		fullContent = decoded
	} else if strings.HasSuffix(globalWorkspace.activeFile, ".go") {
		fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
	} else {
		fullContent = []byte(file.Content)
	}

	// Write to selected path
	err = os.WriteFile(filename, fullContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// IsWorkspaceModified checks if the workspace has unsaved changes
func (a *App) IsWorkspaceModified() bool {
	if globalWorkspace == nil {
		return false
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	return globalWorkspace.modified
}

func updateWorkspaceModifiedLocked() {
	if globalWorkspace == nil {
		return
	}

	globalWorkspace.modified = false
	for _, file := range globalWorkspace.files {
		if file.Modified {
			globalWorkspace.modified = true
			break
		}
	}
}

// GetWorkspaceInfo returns information about the workspace
func (a *App) GetWorkspaceInfo() map[string]interface{} {
	if globalWorkspace == nil {
		return map[string]interface{}{
			"initialized": false,
		}
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	return map[string]interface{}{
		"initialized": globalWorkspace.initialized,
		"workDir":     globalWorkspace.workDir,
		"isTemp":      globalWorkspace.isTemp,
		"fileCount":   len(globalWorkspace.files),
		"activeFile":  globalWorkspace.activeFile,
		"modified":    globalWorkspace.modified,
	}
}

// CleanupWorkspace removes the temporary workspace directory if temp
func (a *App) CleanupWorkspace() error {
	if globalWorkspace == nil || !globalWorkspace.initialized {
		return nil
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// Only remove if it's a temp workspace
	if globalWorkspace.isTemp {
		err := os.RemoveAll(globalWorkspace.workDir)
		if err != nil {
			return fmt.Errorf("failed to cleanup workspace: %w", err)
		}
	}

	globalWorkspace.initialized = false
	globalWorkspace.files = nil
	globalWorkspace = nil

	return nil
}

// domReady is called after the front-end dom is ready
func (a *App) domReady(ctx context.Context) {
	// Initialize workspace when DOM is ready
	err := a.InitWorkspace()
	if err != nil {
		fmt.Printf("Failed to initialize workspace: %v\n", err)
	}
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if globalWorkspace == nil {
		return false
	}

	// Check if we're in temporary workspace mode with files
	globalWorkspace.mu.RLock()
	isTemp := globalWorkspace.isTemp
	hasFiles := len(globalWorkspace.files) > 0
	hasModified := false
	for _, file := range globalWorkspace.files {
		if file.Modified {
			hasModified = true
			break
		}
	}
	globalWorkspace.mu.RUnlock()

	// Warn user if they're in temporary workspace and haven't saved to disk
	if isTemp && hasFiles {
		var message string
		if hasModified {
			message = "您在臨時工作區中有未儲存的變更。\n關閉後這些檔案將會遺失。\n\n是否先建立工作區並儲存檔案？\n\n點「是(Y)」: 建立工作區並儲存\n點「否(N)」: 直接關閉不儲存"
		} else {
			message = "您在臨時工作區中。\n關閉後工作區檔案將會遺失。\n\n是否先建立工作區並儲存檔案？\n\n點「是(Y)」: 建立工作區並儲存\n點「否(N)」: 直接關閉不儲存"
		}

		selection, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "儲存工作區",
			Message: message,
			Buttons: []string{"是(Y)", "否(N)"},
		})

		if err != nil {
			fmt.Printf("Dialog error: %v\n", err)
			return true // Prevent closing on error
		}

		fmt.Printf("Dialog selection: '%s'\n", selection)

		// Handle different possible return values
		if selection == "是(Y)" || selection == "Y" || selection == "Yes" {
			// Try to create workspace
			path, err := a.CreateWorkspace()
			if err != nil {
				runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "儲存失敗",
					Message: "無法建立工作區：" + err.Error() + "\n\n按確定後程式將繼續關閉。",
				})
				// Even if save failed, allow closing since user chose to save
			}
			if path == "" {
				// User cancelled the folder selection - don't close
				return true
			}
			// Successfully saved or user accepted the error, allow closing
		}
		// If selection is "否(N)", "N", "No", or anything else, continue to close
	}

	// Cleanup and allow close
	a.CleanupWorkspace()
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Ensure cleanup
	a.CleanupWorkspace()
	fmt.Println("Idensyra is shutting down...")
}

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// WorkspaceFile represents a single file in the workspace
type WorkspaceFile struct {
	Name         string `json:"name"`
	Content      string `json:"content"`
	Modified     bool   `json:"modified"`
	Size         int64  `json:"size"`
	TooLarge     bool   `json:"tooLarge"`
	IsBinary     bool   `json:"isBinary"`
	IsDir        bool   `json:"isDir"`
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

type workspaceScanEntry struct {
	path        string
	displayName string
	isDir       bool
	size        int64
}

const (
	concurrentReadThreshold = 4 * 1024 * 1024
	concurrentReadChunkSize = 2 * 1024 * 1024
	concurrentReadWorkers   = 4
	maxPreviewBytes         = 55 * 1024 * 1024
)

func emitWorkspaceOpenProgress(ctx context.Context, phase string, fileName string, bytesRead int64, totalBytes int64, processedFiles int, totalFiles int, message string) {
	if ctx == nil {
		return
	}
	payload := map[string]any{
		"phase":          phase,
		"fileName":       fileName,
		"bytesRead":      bytesRead,
		"totalBytes":     totalBytes,
		"processedFiles": processedFiles,
		"totalFiles":     totalFiles,
	}
	if message != "" {
		payload["message"] = message
	}
	runtime.EventsEmit(ctx, "workspace:open-progress", payload)
}

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
		Size:         int64(len(defaultCode)),
		TooLarge:     false,
		IsDir:        false,
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

	diskFiles := make(map[string]struct{})
	_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if path == dirPath {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return nil
		}
		if isHiddenPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		displayName := filepath.ToSlash(relPath)
		info, err := d.Info()
		if err != nil {
			return nil
		}

		diskFiles[displayName] = struct{}{}

		if d.IsDir() {
			if existing, exists := globalWorkspace.files[displayName]; exists {
				existing.Size = info.Size()
				existing.TooLarge = false
				existing.IsBinary = false
				existing.IsDir = true
				existing.Content = ""
				existing.SavedContent = ""
				existing.Modified = false
				existing.IsNew = false
				return nil
			}

			globalWorkspace.files[displayName] = &WorkspaceFile{
				Name:         displayName,
				Content:      "",
				Size:         info.Size(),
				TooLarge:     false,
				IsBinary:     false,
				IsDir:        true,
				SavedContent: "",
				IsNew:        false,
				Modified:     false,
			}
			return nil
		}

		var contentStr string
		var isBinary bool
		tooLarge := isFileTooLarge(info.Size())
		if !tooLarge {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			isBinary = shouldTreatAsBinary(displayName, content)
			if isBinary {
				contentStr = base64.StdEncoding.EncodeToString(content)
			} else {
				contentStr = string(content)
				if strings.HasSuffix(displayName, ".go") {
					contentStr = strings.TrimPrefix(contentStr, preCode+"\n")
					contentStr = strings.TrimSuffix(contentStr, "\n"+endCode)
				}
			}
		}

		if existing, exists := globalWorkspace.files[displayName]; exists {
			if existing.Modified || existing.IsNew {
				return nil
			}
			existing.Size = info.Size()
			existing.TooLarge = tooLarge
			existing.IsBinary = isBinary
			existing.IsDir = false
			if existing.Content != contentStr {
				existing.Content = contentStr
				existing.SavedContent = contentStr
				existing.Modified = false
				existing.IsNew = false
			}
			return nil
		}

		globalWorkspace.files[displayName] = &WorkspaceFile{
			Name:         displayName,
			Content:      contentStr,
			Size:         info.Size(),
			TooLarge:     tooLarge,
			IsBinary:     isBinary,
			IsDir:        false,
			SavedContent: contentStr,
			IsNew:        false,
			Modified:     false,
		}
		return nil
	})

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
		for name, file := range globalWorkspace.files {
			if file.IsDir {
				continue
			}
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

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[cleanName]; !exists {
		return fmt.Errorf("file not found: %s", cleanName)
	}
	if file := globalWorkspace.files[cleanName]; file.IsDir {
		return fmt.Errorf("path is a directory: %s", cleanName)
	}

	globalWorkspace.activeFile = cleanName
	return nil
}

// GetFileContent returns the content of a specific file
// For image files, returns base64 encoded data
func (a *App) GetFileContent(filename string) (string, error) {
	if globalWorkspace == nil {
		return "", fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return "", err
	}

	globalWorkspace.mu.RLock()
	defer globalWorkspace.mu.RUnlock()

	file, exists := globalWorkspace.files[cleanName]
	if !exists {
		return "", fmt.Errorf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return "", fmt.Errorf("path is a directory: %s", cleanName)
	}

	if file.TooLarge {
		return "", fmt.Errorf("file too large to preview")
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

func isBinaryPreviewFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico",
		".mp4", ".webm", ".mov", ".avi", ".mkv", ".m4v", ".mpg", ".mpeg",
		".mp3", ".wav", ".flac", ".ogg", ".aac", ".m4a",
		".pdf",
		".xlsx", ".xlsm", ".xltx", ".xltm":
		return true
	default:
		return false
	}
}

func isBinaryContent(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	if bytes.IndexByte(content, 0x00) != -1 {
		return true
	}
	if !utf8.Valid(content) {
		return true
	}
	var nonPrintable int
	for _, b := range content {
		if b < 9 || (b > 13 && b < 32) {
			nonPrintable++
		}
	}
	return float64(nonPrintable)/float64(len(content)) > 0.3
}

func shouldTreatAsBinary(filename string, content []byte) bool {
	if isBinaryPreviewFile(filename) {
		return true
	}
	return isBinaryContent(content)
}

// UpdateFileContent updates the content of a file
func (a *App) UpdateFileContent(filename string, content string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	file, exists := globalWorkspace.files[cleanName]
	if !exists {
		return fmt.Errorf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return fmt.Errorf("path is a directory: %s", cleanName)
	}
	if file.IsBinary {
		return fmt.Errorf("binary files cannot be edited: %s", cleanName)
	}
	if file.TooLarge {
		return fmt.Errorf("file too large to edit")
	}

	if file.Content == content {
		return nil
	}

	file.Content = content
	file.Size = int64(len(content))
	file.Modified = file.IsNew || file.Content != file.SavedContent
	updateWorkspaceModifiedLocked()

	return nil
}

// SaveFile saves a file to disk (in workspace directory)
func (a *App) SaveFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// If using temp workspace, prompt to open/create workspace
	if globalWorkspace.isTemp {
		return fmt.Errorf("temporary workspace: please open or create a workspace first")
	}

	file, exists := globalWorkspace.files[cleanName]
	if !exists {
		return fmt.Errorf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return fmt.Errorf("path is a directory: %s", cleanName)
	}
	if file.TooLarge {
		return fmt.Errorf("file too large to save")
	}

	// Only wrap with preCode/endCode for .go files
	var fullContent []byte
	if file.IsBinary {
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return fmt.Errorf("failed to decode file: %w", err)
		}
		fullContent = decoded
	} else if strings.HasSuffix(cleanName, ".go") {
		fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
	} else {
		fullContent = []byte(file.Content)
	}

	filePath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanName))
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	err = os.WriteFile(filePath, fullContent, 0644)
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
		if file.IsDir || file.TooLarge {
			continue
		}
		// Only wrap with preCode/endCode for .go files
		var fullContent []byte
		if file.IsBinary {
			decoded, err := base64.StdEncoding.DecodeString(file.Content)
			if err != nil {
				return fmt.Errorf("failed to decode file %s: %w", filename, err)
			}
			fullContent = decoded
		} else if strings.HasSuffix(filename, ".go") {
			fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
		} else {
			fullContent = []byte(file.Content)
		}

		filePath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(filename))
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
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

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[cleanName]; exists {
		return fmt.Errorf("file already exists: %s", cleanName)
	}

	// Create new file with appropriate default content based on extension
	var content string
	if strings.HasSuffix(cleanName, ".go") {
		content = defaultCode
	} else {
		content = "" // Empty content for non-Go files
	}

	newFile := &WorkspaceFile{
		Name:         cleanName,
		Content:      content,
		Size:         int64(len(content)),
		TooLarge:     false,
		IsBinary:     isBinaryPreviewFile(cleanName),
		IsDir:        false,
		SavedContent: "",
		IsNew:        true,
		Modified:     true,
	}
	globalWorkspace.files[cleanName] = newFile

	if globalWorkspace.workDir != "" {
		fullPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanName))
		parentDir := filepath.Dir(fullPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
	}

	parentParts := strings.Split(cleanName, "/")
	if len(parentParts) > 1 {
		for i := 1; i < len(parentParts); i++ {
			dirPath := strings.Join(parentParts[:i], "/")
			if _, exists := globalWorkspace.files[dirPath]; !exists {
				globalWorkspace.files[dirPath] = &WorkspaceFile{
					Name:         dirPath,
					Content:      "",
					Size:         0,
					TooLarge:     false,
					IsDir:        true,
					SavedContent: "",
					IsNew:        false,
					Modified:     false,
				}
			}
		}
	}
	updateWorkspaceModifiedLocked()

	return nil
}

// DeleteFile removes a file from the workspace
func (a *App) DeleteFile(filename string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	// Don't allow deleting the last file
	if len(globalWorkspace.files) <= 1 {
		return fmt.Errorf("cannot delete the last file in workspace")
	}

	file, exists := globalWorkspace.files[cleanName]
	if !exists {
		return fmt.Errorf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return fmt.Errorf("path is a directory: %s", cleanName)
	}

	delete(globalWorkspace.files, cleanName)

	// Remove from disk regardless of temp workspace to prevent re-adding on refresh
	if globalWorkspace.workDir != "" {
		filePath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanName))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// If deleted file was active, switch to first available file
	if globalWorkspace.activeFile == cleanName {
		for name := range globalWorkspace.files {
			globalWorkspace.activeFile = name
			break
		}
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// RenameFile renames a file in the workspace
func (a *App) RenameFile(oldName string, newName string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanOld, err := cleanRelativePath(oldName)
	if err != nil {
		return err
	}
	cleanNew, err := cleanRelativePath(newName)
	if err != nil {
		return err
	}

	if cleanNew == "" {
		return fmt.Errorf("new filename cannot be empty")
	}
	if cleanOld == cleanNew {
		return fmt.Errorf("new filename is the same as the old filename")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	file, exists := globalWorkspace.files[cleanOld]
	if !exists {
		return fmt.Errorf("file not found: %s", cleanOld)
	}
	if file.IsDir {
		return fmt.Errorf("path is a directory: %s", cleanOld)
	}
	if _, exists := globalWorkspace.files[cleanNew]; exists {
		return fmt.Errorf("file already exists: %s", cleanNew)
	}

	if globalWorkspace.workDir != "" {
		oldPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanOld))
		newPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanNew))
		parentDir := filepath.Dir(newPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to rename file: %w", err)
		}
	}

	delete(globalWorkspace.files, cleanOld)
	file.Name = cleanNew
	globalWorkspace.files[cleanNew] = file

	if globalWorkspace.activeFile == cleanOld {
		globalWorkspace.activeFile = cleanNew
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// CreateFolder creates a new folder in the workspace
func (a *App) CreateFolder(folderPath string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanPath, err := cleanRelativePath(folderPath)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	if _, exists := globalWorkspace.files[cleanPath]; exists {
		return fmt.Errorf("path already exists: %s", cleanPath)
	}

	if globalWorkspace.workDir != "" {
		fullPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanPath))
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}

	globalWorkspace.files[cleanPath] = &WorkspaceFile{
		Name:         cleanPath,
		Content:      "",
		Size:         0,
		TooLarge:     false,
		IsDir:        true,
		SavedContent: "",
		IsNew:        false,
		Modified:     false,
	}

	parentParts := strings.Split(cleanPath, "/")
	if len(parentParts) > 1 {
		for i := 1; i < len(parentParts); i++ {
			dirPath := strings.Join(parentParts[:i], "/")
			if _, exists := globalWorkspace.files[dirPath]; !exists {
				globalWorkspace.files[dirPath] = &WorkspaceFile{
					Name:         dirPath,
					Content:      "",
					Size:         0,
					TooLarge:     false,
					IsDir:        true,
					SavedContent: "",
					IsNew:        false,
					Modified:     false,
				}
			}
		}
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// DeleteFolder removes a folder and its contents from the workspace
func (a *App) DeleteFolder(folderPath string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanPath, err := cleanRelativePath(folderPath)
	if err != nil {
		return err
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	file, exists := globalWorkspace.files[cleanPath]
	if !exists || !file.IsDir {
		return fmt.Errorf("folder not found: %s", cleanPath)
	}

	prefix := cleanPath + "/"
	for name := range globalWorkspace.files {
		if name == cleanPath || strings.HasPrefix(name, prefix) {
			delete(globalWorkspace.files, name)
		}
	}

	if globalWorkspace.workDir != "" {
		fullPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanPath))
		if err := os.RemoveAll(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete folder: %w", err)
		}
	}

	if strings.HasPrefix(globalWorkspace.activeFile, prefix) || globalWorkspace.activeFile == cleanPath {
		globalWorkspace.activeFile = ""
	}

	updateWorkspaceModifiedLocked()
	return nil
}

// RenameFolder renames a folder and updates its contents
func (a *App) RenameFolder(oldPath string, newPath string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanOld, err := cleanRelativePath(oldPath)
	if err != nil {
		return err
	}
	cleanNew, err := cleanRelativePath(newPath)
	if err != nil {
		return err
	}
	if cleanOld == cleanNew {
		return fmt.Errorf("new folder name is the same as the old name")
	}

	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	file, exists := globalWorkspace.files[cleanOld]
	if !exists || !file.IsDir {
		return fmt.Errorf("folder not found: %s", cleanOld)
	}
	if _, exists := globalWorkspace.files[cleanNew]; exists {
		return fmt.Errorf("path already exists: %s", cleanNew)
	}

	if globalWorkspace.workDir != "" {
		oldFull := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanOld))
		newFull := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(cleanNew))
		parentDir := filepath.Dir(newFull)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := os.Rename(oldFull, newFull); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to rename folder: %w", err)
		}
	}

	updates := make(map[string]*WorkspaceFile)
	oldPrefix := cleanOld + "/"
	newPrefix := cleanNew + "/"
	for name, entry := range globalWorkspace.files {
		if name == cleanOld || strings.HasPrefix(name, oldPrefix) {
			newName := strings.Replace(name, cleanOld, cleanNew, 1)
			entry.Name = newName
			updates[newName] = entry
			delete(globalWorkspace.files, name)
		}
	}
	globalWorkspace.files[cleanNew] = file
	for name, entry := range updates {
		globalWorkspace.files[name] = entry
	}

	if globalWorkspace.activeFile == cleanOld {
		globalWorkspace.activeFile = cleanNew
	} else if strings.HasPrefix(globalWorkspace.activeFile, oldPrefix) {
		globalWorkspace.activeFile = strings.Replace(globalWorkspace.activeFile, oldPrefix, newPrefix, 1)
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

	entries := make([]workspaceScanEntry, 0)
	var totalBytes int64
	totalFiles := 0
	_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == dirPath {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return nil
		}
		if isHiddenPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		displayName := filepath.ToSlash(relPath)
		info, err := d.Info()
		if err != nil {
			return nil
		}

		entry := workspaceScanEntry{
			path:        path,
			displayName: displayName,
			isDir:       d.IsDir(),
			size:        info.Size(),
		}
		entries = append(entries, entry)

		if !entry.isDir {
			totalBytes += entry.size
			totalFiles++
		}
		return nil
	})

	if len(entries) == 0 {
		emitWorkspaceOpenProgress(a.ctx, "error", "", 0, 0, 0, 0, "no files found in selected directory")
		return "", fmt.Errorf("no files found in selected directory")
	}

	emitWorkspaceOpenProgress(a.ctx, "start", "Opening workspace...", 0, totalBytes, 0, totalFiles, "")

	workspaceFiles := make(map[string]*WorkspaceFile)
	var processedBytes int64
	processedFiles := 0
	for _, entry := range entries {
		if entry.isDir {
			workspaceFiles[entry.displayName] = &WorkspaceFile{
				Name:         entry.displayName,
				Content:      "",
				Size:         entry.size,
				TooLarge:     false,
				IsDir:        true,
				SavedContent: "",
				IsNew:        false,
				Modified:     false,
			}
			continue
		}

		var contentStr string
		var isBinary bool
		tooLarge := isFileTooLarge(entry.size)
		if !tooLarge {
			content, err := os.ReadFile(entry.path)
			if err != nil {
				processedBytes += entry.size
				processedFiles++
				emitWorkspaceOpenProgress(a.ctx, "progress", entry.displayName, processedBytes, totalBytes, processedFiles, totalFiles, "")
				continue
			}

			isBinary = shouldTreatAsBinary(entry.displayName, content)
			if isBinary {
				contentStr = base64.StdEncoding.EncodeToString(content)
			} else {
				contentStr = string(content)
				if strings.HasSuffix(entry.displayName, ".go") {
					contentStr = strings.TrimPrefix(contentStr, preCode+"\n")
					contentStr = strings.TrimSuffix(contentStr, "\n"+endCode)
				}
			}
		}

		workspaceFiles[entry.displayName] = &WorkspaceFile{
			Name:         entry.displayName,
			Content:      contentStr,
			Size:         entry.size,
			TooLarge:     tooLarge,
			IsBinary:     isBinary,
			IsDir:        false,
			SavedContent: contentStr,
			IsNew:        false,
			Modified:     false,
		}

		processedBytes += entry.size
		processedFiles++
		emitWorkspaceOpenProgress(a.ctx, "progress", entry.displayName, processedBytes, totalBytes, processedFiles, totalFiles, "")
	}

	if len(workspaceFiles) == 0 {
		emitWorkspaceOpenProgress(a.ctx, "error", "", processedBytes, totalBytes, processedFiles, totalFiles, "no files found in selected directory")
		return "", fmt.Errorf("no files found in selected directory")
	}

	emitWorkspaceOpenProgress(a.ctx, "done", "Workspace ready", processedBytes, totalBytes, processedFiles, totalFiles, "")

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

	for filename, file := range workspaceFiles {
		globalWorkspace.files[filename] = file
	}

	// Set first file as active
	for name, file := range globalWorkspace.files {
		if file.IsDir {
			continue
		}
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
		if file.IsDir {
			folderPath := filepath.Join(selectedPath, filepath.FromSlash(filename))
			if err := os.MkdirAll(folderPath, 0755); err != nil {
				return "", fmt.Errorf("failed to create folder %s: %w", filename, err)
			}
		}
	}

	for filename, file := range globalWorkspace.files {
		if file.IsDir {
			continue
		}
		filePath := filepath.Join(selectedPath, filepath.FromSlash(filename))
		if file.TooLarge {
			sourcePath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(filename))
			if err := copyFile(sourcePath, filePath); err != nil {
				return "", fmt.Errorf("failed to copy file %s: %w", filename, err)
			}
		} else {
			var fullContent []byte
			if file.IsBinary {
				decoded, err := base64.StdEncoding.DecodeString(file.Content)
				if err != nil {
					return "", fmt.Errorf("failed to decode file %s: %w", filename, err)
				}
				fullContent = decoded
			} else if strings.HasSuffix(filename, ".go") {
				fullContent = []byte(preCode + "\n" + file.Content + "\n" + endCode)
			} else {
				fullContent = []byte(file.Content)
			}

			err = os.WriteFile(filePath, fullContent, 0644)
			if err != nil {
				return "", fmt.Errorf("failed to write file %s: %w", filename, err)
			}
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
	return a.ImportFileToWorkspaceAt("")
}

// ImportFileToWorkspaceAt imports an external file into the workspace at target folder
func (a *App) ImportFileToWorkspaceAt(targetDir string) error {
	if globalWorkspace == nil {
		return fmt.Errorf("workspace not initialized")
	}

	cleanTarget, err := cleanOptionalRelativePath(targetDir)
	if err != nil {
		return err
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

	filename := filepath.Base(filePath)
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file content
	var progressFn func(readBytes, totalBytes int64)
	if a.ctx != nil {
		var progressMu sync.Mutex
		progressFn = func(readBytes, totalBytes int64) {
			phase := "progress"
			if totalBytes == 0 {
				phase = "done"
			} else if readBytes == 0 {
				phase = "start"
			} else if readBytes >= totalBytes {
				phase = "done"
			}
			progressMu.Lock()
			runtime.EventsEmit(a.ctx, "import:file-progress", map[string]any{
				"phase":      phase,
				"fileName":   filename,
				"bytesRead":  readBytes,
				"totalBytes": totalBytes,
			})
			progressMu.Unlock()
		}
	}

	var content []byte
	tooLarge := isFileTooLarge(info.Size())
	if !tooLarge {
		content, err = readFileConcurrently(filePath, progressFn)
		if err != nil {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "import:file-progress", map[string]any{
					"phase":    "error",
					"fileName": filename,
					"message":  err.Error(),
				})
			}
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Get base filename
	// Keep original filename as-is

	// Check if file already exists
	globalWorkspace.mu.Lock()
	defer globalWorkspace.mu.Unlock()

	finalName := filename
	if cleanTarget != "" {
		finalName = cleanTarget + "/" + filename
	}

	if _, exists := globalWorkspace.files[finalName]; exists {
		// Generate unique name
		ext := filepath.Ext(finalName)
		base := strings.TrimSuffix(finalName, ext)
		counter := 1
		for {
			newName := fmt.Sprintf("%s_%d%s", base, counter, ext)
			if _, exists := globalWorkspace.files[newName]; !exists {
				finalName = newName
				break
			}
			counter++
		}
	}

	// Convert content to string (base64 for binary files)
	var contentStr string
	isBinary := false
	if !tooLarge {
		isBinary = shouldTreatAsBinary(finalName, content)
		if isBinary {
			contentStr = base64.StdEncoding.EncodeToString(content)
		} else {
			contentStr = string(content)
		}
	} else {
		isBinary = isBinaryPreviewFile(finalName)
	}

	if tooLarge && globalWorkspace.workDir != "" {
		targetPath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(finalName))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := copyFile(filePath, targetPath); err != nil {
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "import:file-progress", map[string]any{
					"phase":    "error",
					"fileName": finalName,
					"message":  err.Error(),
				})
			}
			return fmt.Errorf("failed to import large file: %w", err)
		}
		if progressFn != nil {
			progressFn(0, info.Size())
			progressFn(info.Size(), info.Size())
		}
	}

	globalWorkspace.files[finalName] = &WorkspaceFile{
		Name:         finalName,
		Content:      contentStr,
		Size:         info.Size(),
		TooLarge:     tooLarge,
		IsBinary:     isBinary,
		SavedContent: "",
		IsNew:        !tooLarge,
		Modified:     !tooLarge,
	}

	parentParts := strings.Split(finalName, "/")
	if len(parentParts) > 1 {
		for i := 1; i < len(parentParts); i++ {
			dirPath := strings.Join(parentParts[:i], "/")
			if _, exists := globalWorkspace.files[dirPath]; !exists {
				globalWorkspace.files[dirPath] = &WorkspaceFile{
					Name:         dirPath,
					Content:      "",
					Size:         0,
					TooLarge:     false,
					IsDir:        true,
					SavedContent: "",
					IsNew:        false,
					Modified:     false,
				}
			}
		}
	}
	updateWorkspaceModifiedLocked()

	return nil
}

func readFileConcurrently(filePath string, onProgress func(readBytes, totalBytes int64)) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := info.Size()
	if size == 0 {
		if onProgress != nil {
			onProgress(0, 0)
		}
		return []byte{}, nil
	}

	maxInt := int64(int(^uint(0) >> 1))
	if size > maxInt {
		return nil, fmt.Errorf("file too large: %d bytes", size)
	}

	if size < concurrentReadThreshold {
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(size, size)
		}
		return data, nil
	}

	if onProgress != nil {
		onProgress(0, size)
	}

	buf := make([]byte, size)
	chunkSize := int64(concurrentReadChunkSize)
	totalChunks := int((size + chunkSize - 1) / chunkSize)
	workers := concurrentReadWorkers
	if totalChunks < workers {
		workers = totalChunks
	}

	var wg sync.WaitGroup
	var readErr error
	var errOnce sync.Once
	var bytesRead int64
	setErr := func(err error) {
		errOnce.Do(func() {
			readErr = err
		})
	}

	chunkCh := make(chan int, totalChunks)
	for i := 0; i < totalChunks; i++ {
		chunkCh <- i
	}
	close(chunkCh)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for idx := range chunkCh {
				start := int64(idx) * chunkSize
				end := start + chunkSize
				if end > size {
					end = size
				}

				n, err := file.ReadAt(buf[start:end], start)
				if err != nil && err != io.EOF {
					setErr(err)
					return
				}
				if int64(n) != end-start {
					setErr(io.ErrUnexpectedEOF)
					return
				}
				if onProgress != nil {
					totalRead := atomic.AddInt64(&bytesRead, int64(n))
					onProgress(totalRead, size)
				}
			}
		}()
	}
	wg.Wait()

	if readErr != nil {
		return nil, readErr
	}

	if onProgress != nil {
		onProgress(size, size)
	}

	return buf, nil
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
	if file.IsDir {
		return fmt.Errorf("active path is a directory")
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
		return fmt.Errorf("User cancelled")
	}

	// Write file content based on file type
	var fullContent []byte
	if file.TooLarge {
		sourcePath := filepath.Join(globalWorkspace.workDir, filepath.FromSlash(globalWorkspace.activeFile))
		if err := copyFile(sourcePath, filename); err != nil {
			return fmt.Errorf("failed to export file: %w", err)
		}
		return nil
	} else if file.IsBinary {
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return fmt.Errorf("failed to decode file: %w", err)
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

func isFileTooLarge(size int64) bool {
	return size > maxPreviewBytes
}

func cleanRelativePath(input string) (string, error) {
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

func cleanOptionalRelativePath(input string) (string, error) {
	clean := strings.TrimSpace(input)
	if clean == "" {
		return "", nil
	}
	return cleanRelativePath(clean)
}

func isHiddenPath(relPath string) bool {
	clean := filepath.Clean(relPath)
	for _, part := range strings.Split(clean, string(os.PathSeparator)) {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func copyFile(sourcePath string, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return err
	}

	return target.Sync()
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

// GetExcelPreview returns an HTML table preview for the first sheet.
func (a *App) GetExcelPreview(filename string, maxRows int, maxCols int) (string, error) {
	return a.getExcelPreview(filename, "", maxRows, maxCols)
}

// GetExcelSheets returns all worksheet names in the Excel file.
func (a *App) GetExcelSheets(filename string) ([]string, error) {
	excel, err := openExcelFromWorkspace(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = excel.Close()
	}()

	sheets := excel.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found")
	}
	return sheets, nil
}

// GetExcelSheetPreview returns an HTML table preview for the specified sheet.
func (a *App) GetExcelSheetPreview(filename string, sheetName string, maxRows int, maxCols int) (string, error) {
	return a.getExcelPreview(filename, sheetName, maxRows, maxCols)
}

func openExcelFromWorkspace(filename string) (*excelize.File, error) {
	if globalWorkspace == nil {
		return nil, fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return nil, err
	}

	globalWorkspace.mu.RLock()
	file, exists := globalWorkspace.files[cleanName]
	globalWorkspace.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return nil, fmt.Errorf("path is a directory: %s", cleanName)
	}
	if file.TooLarge {
		return nil, fmt.Errorf("file too large to preview")
	}

	if file.Content == "" {
		return nil, fmt.Errorf("file content unavailable")
	}

	data, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	reader := bytes.NewReader(data)
	excel, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	return excel, nil
}

func (a *App) getExcelPreview(filename string, sheetName string, maxRows int, maxCols int) (string, error) {
	if maxRows <= 0 {
		maxRows = 50
	}
	if maxCols <= 0 {
		maxCols = 20
	}

	excel, err := openExcelFromWorkspace(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = excel.Close()
	}()

	sheets := excel.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("no sheets found")
	}

	targetSheet := sheetName
	if targetSheet == "" {
		targetSheet = sheets[0]
	} else {
		found := false
		for _, sheet := range sheets {
			if sheet == targetSheet {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("sheet not found: %s", targetSheet)
		}
	}

	rows, err := excel.Rows(targetSheet)
	if err != nil {
		return "", fmt.Errorf("failed to read rows: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var builder strings.Builder
	builder.WriteString("<table><tbody>")

	rowCount := 0
	for rows.Next() && rowCount < maxRows {
		cols, err := rows.Columns()
		if err != nil {
			return "", fmt.Errorf("failed to read row: %w", err)
		}
		builder.WriteString("<tr>")
		for colIdx, col := range cols {
			if colIdx >= maxCols {
				break
			}
			builder.WriteString("<td>")
			builder.WriteString(html.EscapeString(col))
			builder.WriteString("</td>")
		}
		builder.WriteString("</tr>")
		rowCount++
	}
	builder.WriteString("</tbody></table>")

	return builder.String(), nil
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

	if isTemp && hasFiles {
		var message string
		if hasModified {
			message = "You have unsaved changes in a temporary workspace.\nClosing will discard these files.\n\nDo you want to create a workspace and save them first?\n\nChoose \"Yes (Y)\" to create a workspace and save.\nChoose \"No (N)\" to close without saving."
		} else {
			message = "You are working in a temporary workspace.\nClosing will discard these files.\n\nDo you want to create a workspace and save them first?\n\nChoose \"Yes (Y)\" to create a workspace and save.\nChoose \"No (N)\" to close without saving."
		}

		selection, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "Save Workspace",
			Message: message,
			Buttons: []string{"Yes (Y)", "No (N)"},
		})

		if err != nil {
			fmt.Printf("Dialog error: %v\n", err)
			return true // Prevent closing on error
		}

		fmt.Printf("Dialog selection: '%s'\n", selection)

		if selection == "Yes (Y)" || selection == "Y" || selection == "Yes" {
			path, err := a.CreateWorkspace()
			if err != nil {
				runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
					Type:    runtime.ErrorDialog,
					Title:   "Save Failed",
					Message: "Failed to create workspace: " + err.Error() + "\n\nThe app will continue closing after you confirm.",
				})
			}
			if path == "" {
				return true
			}
		}

		a.CleanupWorkspace()
		return false
	}

	// Warn user if they're in temporary workspace and haven't saved to disk
	if false && isTemp && hasFiles {
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

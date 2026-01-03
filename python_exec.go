package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ExecutePythonFile runs a workspace Python file via py.RunFile and returns HTML output.
func (a *App) ExecutePythonFile(filename string, content string) string {
	if globalWorkspace == nil || !globalWorkspace.initialized {
		return "workspace not initialized"
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err.Error()
	}
	if !strings.HasSuffix(strings.ToLower(cleanName), ".py") {
		return "only .py files can be executed"
	}

	globalWorkspace.mu.RLock()
	file, exists := globalWorkspace.files[cleanName]
	workDir := globalWorkspace.workDir
	globalWorkspace.mu.RUnlock()

	if !exists {
		return fmt.Sprintf("file not found: %s", cleanName)
	}
	if file.IsDir {
		return fmt.Sprintf("path is a directory: %s", cleanName)
	}
	if file.TooLarge {
		return "file too large to execute"
	}
	if file.IsBinary {
		return "binary files cannot be executed"
	}
	if workDir == "" {
		return "workspace directory not set"
	}

	// Create a temporary file with encoding setup instead of modifying the original
	tempFile, err := createTempPythonFile(workDir, cleanName, content)
	if err != nil {
		return fmt.Sprintf("failed to prepare python file: %v", err)
	}
	defer os.Remove(tempFile) // Clean up temp file after execution

	code := buildPythonFileRunner(tempFile)
	return executeGoCode(code, "dark")
}

// pythonEncodingSetup is prepended to Python files to ensure UTF-8 output on Windows
const pythonEncodingSetup = `# -*- coding: utf-8 -*-
import sys
if sys.platform == 'win32':
    if hasattr(sys.stdout, 'reconfigure'):
        sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    if hasattr(sys.stderr, 'reconfigure'):
        sys.stderr.reconfigure(encoding='utf-8', errors='replace')
# End of encoding setup
`

// createTempPythonFile creates a temporary Python file with encoding setup
// without modifying the original file
func createTempPythonFile(workDir string, cleanName string, content string) (string, error) {
	// Create temp file in a temporary directory
	tempDir := filepath.Join(workDir, ".igonb_temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Create temp file with encoding setup prepended
	tempFile, err := os.CreateTemp(tempDir, "*.py")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write encoding setup and then the original content
	if _, err := io.WriteString(tempFile, pythonEncodingSetup); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}
	if _, err := io.WriteString(tempFile, content); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func buildPythonFileRunner(relPath string) string {
	quotedPath := strconv.Quote(relPath)
	return fmt.Sprintf(`import (
	"fmt"
	"github.com/HazelnutParadise/insyra/py"
)

func main() {
	if err := py.RunFile(nil, %s); err != nil {
		fmt.Println(err)
	}
}`, quotedPath)
}

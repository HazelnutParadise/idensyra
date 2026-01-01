package main

import (
	"fmt"
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

	if err := writeExecutionFile(workDir, cleanName, content); err != nil {
		return fmt.Sprintf("failed to prepare python file: %v", err)
	}

	code := buildPythonFileRunner(cleanName)
	return executeGoCode(code, "dark")
}

func writeExecutionFile(workDir string, cleanName string, content string) error {
	fullPath := filepath.Join(workDir, filepath.FromSlash(cleanName))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
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

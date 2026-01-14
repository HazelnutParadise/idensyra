package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ExecutePythonFile runs a workspace Python file via py.RunFile and returns HTML output.
func (a *App) ExecutePythonFile(filename string, content string) string {
	result, _ := a.ExecutePythonFileContent(filename, content)
	return result
}

// ExecutePythonFileContent runs a workspace Python file and returns HTML output with error.
func (a *App) ExecutePythonFileContent(filename string, content string) (string, error) {
	if globalWorkspace == nil || !globalWorkspace.initialized {
		return "workspace not initialized", fmt.Errorf("workspace not initialized")
	}

	cleanName, err := cleanRelativePath(filename)
	if err != nil {
		return err.Error(), err
	}
	if !strings.HasSuffix(strings.ToLower(cleanName), ".py") {
		return "only .py files can be executed", fmt.Errorf("only .py files can be executed")
	}

	globalWorkspace.mu.RLock()
	file, exists := globalWorkspace.files[cleanName]
	workDir := globalWorkspace.workDir
	globalWorkspace.mu.RUnlock()

	if !exists {
		msg := fmt.Sprintf("file not found: %s", cleanName)
		return msg, fmt.Errorf(msg)
	}
	if file.IsDir {
		msg := fmt.Sprintf("path is a directory: %s", cleanName)
		return msg, fmt.Errorf(msg)
	}
	if file.TooLarge {
		return "file too large to execute", fmt.Errorf("file too large to execute")
	}
	if file.IsBinary {
		return "binary files cannot be executed", fmt.Errorf("binary files cannot be executed")
	}
	if workDir == "" {
		return "workspace directory not set", fmt.Errorf("workspace directory not set")
	}

	// Execute python content directly; insyra will handle temp file concerns internally.
	fullContent := pythonEncodingSetup + content
	code := buildPythonFileRunnerFromContent(fullContent)
	return executeGoCode(code, "dark"), nil
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
func buildPythonFileRunnerFromContent(content string) string {
	quoted := strconv.Quote(content)
	return fmt.Sprintf(`import (
	"fmt"
	"github.com/HazelnutParadise/insyra/py"
)

func main() {
	if err := py.RunCode(nil, %s); err != nil {
		fmt.Println(err)
	}
}`, quoted)
}

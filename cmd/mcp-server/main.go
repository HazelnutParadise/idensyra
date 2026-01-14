package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/HazelnutParadise/idensyra/internal"
	"github.com/HazelnutParadise/idensyra/mcp"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var preCode = `package main
`

func main() {
	var (
		workspaceRoot = flag.String("workspace", ".", "Workspace root directory")
		configFile    = flag.String("config", "", "Configuration file path")
	)
	flag.Parse()

	// Get absolute workspace path
	absWorkspace, err := filepath.Abs(*workspaceRoot)
	if err != nil {
		log.Fatalf("Failed to get absolute workspace path: %v", err)
	}

	// Load configuration
	config := mcp.DefaultConfig()
	if *configFile != "" {
		// TODO: Load config from file if needed
		log.Printf("Using default configuration")
	}

	// Create confirmation function (auto-approve for CLI mode)
	confirmFunc := func(operation, details string) bool {
		// In CLI mode, we auto-approve all operations
		// In GUI mode, this would show a dialog
		log.Printf("Operation: %s - %s (auto-approved)", operation, details)
		return true
	}

	// Create execution functions
	executeGoFunc := func(code string, colorBG string) string {
		return executeGoCode(code, colorBG)
	}

	executePyFunc := func(filePath string) (string, error) {
		return executePythonFile(filePath)
	}

	executeCellFunc := func(language, code string) (string, error) {
		switch language {
		case "go":
			return executeGoCode(code, "dark"), nil
		case "python":
			// For Python, we need to create a temp file since executePythonFile expects a file path
			tmpFile := filepath.Join(absWorkspace, fmt.Sprintf(".tmp_py_%d.py", time.Now().UnixNano()))
			defer os.Remove(tmpFile)
			if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
				return "", fmt.Errorf("failed to create temp file: %v", err)
			}
			return executePythonFile(tmpFile)
		case "markdown":
			return "Markdown cell (no execution)", nil
		default:
			return "", fmt.Errorf("unsupported language: %s", language)
		}
	}

	openWorkspaceFunc := func(path string) error {
		log.Printf("Opening workspace: %s", path)
		return nil
	}

	saveWorkspaceFunc := func(path string) error {
		log.Printf("Saving workspace to: %s", path)
		return nil
	}

	saveChangesFunc := func() error {
		log.Printf("Saving changes")
		return nil
	}

	setActiveFileFunc := func(path string) error {
		log.Printf("Setting active file: %s", path)
		// In CLI mode, this just logs. In GUI mode, this would call the actual SetActiveFile function
		return nil
	}

	importFileFunc := func(sourcePath, targetDir string) error {
		log.Printf("Importing file: %s to %s", sourcePath, targetDir)
		// In CLI mode, we just copy the file
		// Read source file
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source file: %v", err)
		}

		// Determine target path
		filename := filepath.Base(sourcePath)
		targetPath := filepath.Join(absWorkspace, targetDir, filename)

		// Create target directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %v", err)
		}

		// Write to target
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write target file: %v", err)
		}

		log.Printf("File imported successfully: %s", filename)
		return nil
	}

	// Provide file backend functions (CLI implementations)
	readFileFunc := func(path string) (string, error) {
		full := filepath.Join(absWorkspace, filepath.FromSlash(path))
		b, err := os.ReadFile(full)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	writeFileFunc := func(path string, content string) error {
		full := filepath.Join(absWorkspace, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}
		return os.WriteFile(full, []byte(content), 0644)
	}
	createFileFunc := func(path string, content string) error {
		full := filepath.Join(absWorkspace, filepath.FromSlash(path))
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("file already exists")
		}
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}
		return os.WriteFile(full, []byte(content), 0644)
	}
	deleteFileFunc := func(path string) error {
		full := filepath.Join(absWorkspace, filepath.FromSlash(path))
		return os.Remove(full)
	}
	renameFileFunc := func(oldPath, newPath string) error {
		oldFull := filepath.Join(absWorkspace, filepath.FromSlash(oldPath))
		newFull := filepath.Join(absWorkspace, filepath.FromSlash(newPath))
		if err := os.MkdirAll(filepath.Dir(newFull), 0755); err != nil {
			return err
		}
		return os.Rename(oldFull, newFull)
	}
	listFilesFunc := func(dir string) (string, error) {
		full := filepath.Join(absWorkspace, filepath.FromSlash(dir))
		entries, err := os.ReadDir(full)
		if err != nil {
			return "", err
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
		return result, nil
	}

	// Create MCP server
	server := mcp.NewServer(
		config,
		absWorkspace,
		confirmFunc,
		executeGoFunc,
		executePyFunc,
		executeCellFunc,
		openWorkspaceFunc,
		saveWorkspaceFunc,
		saveChangesFunc,
		setActiveFileFunc,
		importFileFunc,
		readFileFunc,
		writeFileFunc,
		createFileFunc,
		deleteFileFunc,
		renameFileFunc,
		listFilesFunc,
	)

	log.Printf("MCP Server started. Workspace: %s", absWorkspace)
	log.Printf("Available tools: %d", len(server.ListTools()))

	// Serve on stdin/stdout
	ctx := context.Background()
	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// executeGoCode executes Go code using yaegi interpreter
func executeGoCode(code string, colorBG string) string {
	var buf bytes.Buffer

	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	i.Use(stdlib.Symbols)
	i.Use(internal.Symbols)

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return fmt.Sprintf("Failed to execute code: %v", pipeErr)
	}
	os.Stdout = w
	os.Stderr = w

	outputChan := make(chan string, 1)
	go func() {
		var outputBuf bytes.Buffer
		io.Copy(&outputBuf, r)
		outputChan <- outputBuf.String()
	}()

	execErr := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				if rErr, ok := r.(error); ok {
					err = rErr
				} else {
					err = fmt.Errorf("%v", r)
				}
			}
		}()
		_, err = i.Eval(code)
		return err
	}()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	output := <-outputChan
	result := buf.String() + output

	if execErr != nil {
		result += fmt.Sprintf("\nFailed to execute code: %v", execErr)
	}

	return result
}

// executePythonFile executes a Python file
func executePythonFile(filePath string) (string, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "python3", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

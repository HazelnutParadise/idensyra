package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/HazelnutParadise/idensyra/internal"
	"github.com/HazelnutParadise/insyra"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const version = "0.1.0"

var preCode = `package main
`
var endCode = ``

var defaultCode = `import (
	"fmt"
	"log"
	"github.com/HazelnutParadise/insyra/isr"
	"github.com/HazelnutParadise/insyra"
	"github.com/HazelnutParadise/insyra/datafetch"
	"github.com/HazelnutParadise/insyra/stats"
	"github.com/HazelnutParadise/insyra/parallel"
	"github.com/HazelnutParadise/insyra/plot"
	"github.com/HazelnutParadise/insyra/gplot"
	"github.com/HazelnutParadise/insyra/lpgen"
	"github.com/HazelnutParadise/insyra/csvxl"
	"github.com/HazelnutParadise/insyra/parquet"
	"github.com/HazelnutParadise/insyra/mkt"
	"github.com/HazelnutParadise/insyra/py"

	// No lp package support
	// No other third party package support
)

func main() {
	fmt.Println("Hello, World!")
	log.Println("this is a log message")
	dl := isr.DL.Of(1, 2, 3)
	insyra.Show("My_Data", dl)
}`

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Idensyra is starting...")
	// Workspace initialization is done in domReady to ensure frontend is ready
}

// ExecuteCode executes Go code and returns the result as HTML
func (a *App) ExecuteCode(code string) string {
	return executeGoCode(code, "dark")
}

// ExecuteCodeWithColorBG executes code with specific background color theme
func (a *App) ExecuteCodeWithColorBG(code string, colorBG string) string {
	return executeGoCode(code, colorBG)
}

// GetVersion returns version information
func (a *App) GetVersion() map[string]string {
	return map[string]string{
		"idensyra": version,
		"insyra":   insyra.Version,
	}
}

// GetDefaultCode returns the default code template
func (a *App) GetDefaultCode() string {
	return defaultCode
}

// GetSymbols returns all available symbols for autocomplete
func (a *App) GetSymbols() []string {
	symbols := make([]string, 0)

	for packageFullName, symbol := range internal.Symbols {
		packageName := strings.Split(packageFullName, "/")[len(strings.Split(packageFullName, "/"))-1]
		for funcName := range symbol {
			if funcName != "init" && funcName != "main" && !strings.HasPrefix(funcName, "_") {
				symbols = append(symbols, packageName+"."+funcName)
			}
		}
	}

	for packageFullName, symbol := range stdlib.Symbols {
		packageName := strings.Split(packageFullName, "/")[len(strings.Split(packageFullName, "/"))-1]
		for funcName := range symbol {
			if funcName != "init" && funcName != "main" && !strings.HasPrefix(funcName, "_") {
				symbols = append(symbols, packageName+"."+funcName)
			}
		}
	}

	return symbols
}

// SaveCode saves code to a file using file dialog
func (a *App) SaveCode(code string) error {
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "code.go",
		Title:           "Save Code",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Go Files (*.go)",
				Pattern:     "*.go",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return err
	}
	if filename == "" {
		return nil // User cancelled
	}

	fullCode := preCode + "\n" + code + "\n" + endCode
	return os.WriteFile(filename, []byte(fullCode), 0644)
}

// LoadCode loads code from a file using file dialog
func (a *App) LoadCode() (string, error) {
	filename, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Load Code",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Go Files (*.go)",
				Pattern:     "*.go",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if filename == "" {
		return "", nil // User cancelled
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveResult saves execution result to a file using file dialog
func (a *App) SaveResult(result string) error {
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "result.txt",
		Title:           "Save Result",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return err
	}
	if filename == "" {
		return nil // User cancelled
	}

	return os.WriteFile(filename, []byte(result), 0644)
}

// OpenGitHub opens the GitHub repository in the default browser
func (a *App) OpenGitHub() {
	runtime.BrowserOpenURL(a.ctx, "https://github.com/HazelnutParadise/idensyra")
}

// OpenHazelnutParadise opens the HazelnutParadise website in the default browser
func (a *App) OpenHazelnutParadise() {
	runtime.BrowserOpenURL(a.ctx, "https://hazelnut-paradise.com")
}

// executeGoCode uses yaegi to execute dynamic Go code and capture all output
func executeGoCode(code string, colorBG string) string {
	// Prepare a bytes.Buffer to capture all output
	var buf bytes.Buffer

	// Initialize yaegi interpreter and set Stdout and Stderr
	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	i.Use(stdlib.Symbols)   // Load standard library
	i.Use(internal.Symbols) // Load internal package

	// Redirect standard output and standard error
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Create a channel to receive output
	outputChan := make(chan string)
	go func() {
		var outputBuf bytes.Buffer
		io.Copy(&outputBuf, r)
		outputChan <- outputBuf.String()
	}()

	// Execute code
	_, err := i.Eval(code)
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		return fmt.Sprintf("執行代碼失敗: %v", err)
	}

	// Restore standard output and standard error
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Get all output
	output := <-outputChan

	// Merge yaegi captured output and redirected output
	result := buf.String() + output

	// Convert ANSI to HTML based on color scheme
	if colorBG == "light" || colorBG == "dark" {
		return internal.AnsiToHTMLWithBG(result, colorBG)
	}
	return internal.AnsiToHTML(result)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/HazelnutParadise/idensyra/igonb"
	"github.com/HazelnutParadise/idensyra/internal"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var igonbExecutorMu sync.Mutex
var igonbExecutors = map[string]*igonb.Executor{}

type igonbRunMode int

const (
	igonbRunAll igonbRunMode = iota
	igonbRunUpTo
	igonbRunSingle
)

func getIgonbExecutorKey() string {
	if globalWorkspace == nil {
		return "default"
	}
	globalWorkspace.mu.RLock()
	activeFile := globalWorkspace.activeFile
	workDir := globalWorkspace.workDir
	globalWorkspace.mu.RUnlock()

	if activeFile == "" {
		return "default"
	}
	if workDir == "" {
		return activeFile
	}
	return filepath.Join(workDir, filepath.FromSlash(activeFile))
}

func getIgonbExecutor() (*igonb.Executor, error) {
	key := getIgonbExecutorKey()

	igonbExecutorMu.Lock()
	defer igonbExecutorMu.Unlock()
	if exec, ok := igonbExecutors[key]; ok {
		return exec, nil
	}

	exec, err := igonb.NewExecutorWithSymbols(internal.Symbols)
	if err != nil {
		return nil, err
	}
	igonbExecutors[key] = exec
	return exec, nil
}

// ExecuteIgonbCells executes an igonb notebook up to cellIndex.
// Use -1 to run all cells, and use <= -2 to run a single cell (index = -cellIndex - 2).
func (a *App) ExecuteIgonbCells(content string, cellIndex int) ([]igonb.CellResult, error) {
	nb, err := igonb.Parse([]byte(content))
	if err != nil {
		return nil, err
	}

	mode := igonbRunUpTo
	targetIndex := cellIndex
	if cellIndex == -1 {
		mode = igonbRunAll
		targetIndex = -1
	} else if cellIndex <= -2 {
		mode = igonbRunSingle
		targetIndex = -cellIndex - 2
	}

	if mode != igonbRunAll && (targetIndex < 0 || targetIndex >= len(nb.Cells)) {
		return nil, fmt.Errorf("cell index out of range: %d", targetIndex)
	}

	exec, err := getIgonbExecutor()
	if err != nil {
		return nil, err
	}

	formattedResults := make([]igonb.CellResult, 0)
	results, runErr := runIgonbInWorkspace(exec, nb, mode, targetIndex, func(result igonb.CellResult) {
		formatted := formatIgonbResult(result)
		formattedResults = append(formattedResults, formatted)
		if a != nil && a.ctx != nil {
			runtime.EventsEmit(a.ctx, "igonb:cell-result", formatted)
		}
	})
	if len(formattedResults) != len(results) {
		formattedResults = formatIgonbResults(results)
	}

	return formattedResults, runErr
}

func runIgonbInWorkspace(exec *igonb.Executor, nb *igonb.Notebook, mode igonbRunMode, index int, onResult func(igonb.CellResult)) ([]igonb.CellResult, error) {
	var oldWD string
	var restoreWD bool
	if globalWorkspace != nil {
		globalWorkspace.mu.RLock()
		workspaceDir := globalWorkspace.workDir
		globalWorkspace.mu.RUnlock()
		if workspaceDir != "" {
			if wd, err := os.Getwd(); err == nil {
				if err := os.Chdir(workspaceDir); err == nil {
					oldWD = wd
					restoreWD = true
				}
			}
		}
	}
	if restoreWD {
		defer os.Chdir(oldWD)
	}

	switch mode {
	case igonbRunSingle:
		return exec.RunNotebookCellWithCallback(nb, index, onResult)
	case igonbRunAll:
		return exec.RunNotebookWithCallback(nb, -1, onResult)
	default:
		return exec.RunNotebookWithCallback(nb, index, onResult)
	}
}

// ExecuteIgonb runs all cells and returns formatted results.
func (a *App) ExecuteIgonb(content string) ([]igonb.CellResult, error) {
	return a.ExecuteIgonbCells(content, -1)
}

func formatIgonbResults(results []igonb.CellResult) []igonb.CellResult {
	if len(results) == 0 {
		return results
	}
	formatted := make([]igonb.CellResult, len(results))
	for i, result := range results {
		formatted[i] = formatIgonbResult(result)
	}
	return formatted
}

func formatIgonbResult(result igonb.CellResult) igonb.CellResult {
	if result.Language == "markdown" {
		return result
	}
	if result.Output == "" {
		return result
	}
	result.Output = internal.AnsiToHTMLWithBG(result.Output, "dark")
	return result
}
